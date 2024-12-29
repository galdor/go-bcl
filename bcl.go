package bcl

import (
	"fmt"
	"io"
)

type ElementType string

const (
	ElementTypeBlock ElementType = "block"
	ElementTypeEntry ElementType = "entry"
)

func (t ElementType) WithArticle() string {
	var article string

	switch t {
	case ElementTypeBlock:
		article = "a"
	case ElementTypeEntry:
		article = "an"
	}

	return article + " " + string(t)
}

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

type Value struct {
	Location Span
	Content  any // either Symbol, bool, string, int64 or float64
}

type Symbol string

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

func (elt *Element) Blocks(btype string) []*Element {
	block, ok := elt.Content.(*Block)
	if !ok {
		elt.AddInvalidElementTypeError(ElementTypeBlock)
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
	block, ok := elt.Content.(*Block)
	if !ok {
		elt.AddInvalidElementTypeError(ElementTypeBlock)
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
