package bcl

import (
	"errors"
	"fmt"
	"io"
)

type ElementType string

const (
	ElementTypeBlock ElementType = "block"
	ElementTypeEntry ElementType = "entry"
)

type Document struct {
	Source   string
	TopLevel *Element

	lines []string
}

type Element struct {
	Location            Span
	Content             any // *Block or *Entry
	FollowedByEmptyLine bool

	validationErrors []error
}

type Block struct {
	Type     string
	Name     string
	Elements []*Element
}

type Entry struct {
	Name   string
	Values []*Value
}

func Parse(data []byte, source string) (*Document, error) {
	p := newParser(data, source)

	doc, err := p.Parse()
	if err != nil {
		return nil, err
	}

	doc.lines = p.lines

	return doc, nil
}

func (doc *Document) Print(w io.Writer) error {
	p := newPrinter(w, doc)
	return p.Print()
}

func (elt *Element) Type() (t ElementType) {
	switch elt.Content.(type) {
	case *Block:
		t = ElementTypeBlock
	case *Entry:
		t = ElementTypeEntry
	default:
		panic(fmt.Sprintf("unhandled element content %#v (%T)", elt, elt))
	}

	return
}

func (elt *Element) Id() (id string) {
	switch content := elt.Content.(type) {
	case *Block:
		if content.Name != "" {
			id = content.Type + "." + content.Name
		}
	case *Entry:
		id = content.Name
	default:
		panic(fmt.Sprintf("unhandled element content %#v (%T)", elt, elt))
	}

	return
}

func (doc *Document) Blocks(btype string) []*Element {
	return doc.TopLevel.Blocks(btype)
}

func (doc *Document) Block(btype, name string) *Element {
	return doc.TopLevel.Block(btype, name)
}

func (elt *Element) EnsureBlock() *Block {
	block, ok := elt.Content.(*Block)
	if !ok {
		elt.AddInvalidElementTypeError(ElementTypeBlock)
		return nil
	}

	return block
}

func (elt *Element) EnsureEntry() *Entry {
	entry, ok := elt.Content.(*Entry)
	if !ok {
		elt.AddInvalidElementTypeError(ElementTypeEntry)
		return nil
	}

	return entry
}

func (elt *Element) Blocks(btype string) []*Element {
	block := elt.EnsureBlock()
	if block == nil {
		return nil
	}

	var blocks []*Element

	for _, child := range block.Elements {
		if block, ok := child.Content.(*Block); ok {
			if block.Type == btype {
				blocks = append(blocks, child)
			}
		}
	}

	return blocks
}

func (elt *Element) Block(btype, name string) *Element {
	block := elt.EnsureBlock()
	if block == nil {
		return nil
	}

	for _, child := range block.Elements {
		if block, ok := child.Content.(*Block); ok {
			if block.Type == btype && block.Name == name {
				return child
			}
		}
	}

	return nil
}

func (elt *Element) EntryValues(name string, dests ...any) error {
	block := elt.EnsureBlock()
	if block == nil {
		return nil
	}

	for _, child := range block.Elements {
		if entry, ok := child.Content.(*Entry); ok {
			if entry.Name == name {
				return child.Values(dests...)
			}
		}
	}

	return elt.AddMissingElementError(name, ElementTypeEntry)
}

func (elt *Element) Values(dests ...any) error {
	entry := elt.EnsureEntry()
	if entry == nil {
		return nil
	}

	if len(entry.Values) != len(dests) {
		return elt.AddInvalidEntryValueCountError(len(dests))
	}

	var errs []error

	for i, value := range entry.Values {
		if err := value.Extract(dests[i]); err != nil {
			verr := elt.AddInvalidValueError(value, err)
			errs = append(errs, verr)
		}
	}

	return errors.Join(errs...)
}
