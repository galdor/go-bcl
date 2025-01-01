package bcl

import (
	"fmt"
	"io"
	"reflect"
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

func (elt *Element) CheckTypeBlock() *Block {
	block, ok := elt.Content.(*Block)
	if !ok {
		elt.AddInvalidElementTypeError(ElementTypeBlock)
		return nil
	}

	return block
}

func (elt *Element) CheckTypeEntry() *Entry {
	entry, ok := elt.Content.(*Entry)
	if !ok {
		elt.AddInvalidElementTypeError(ElementTypeEntry)
		return nil
	}

	return entry
}

func (elt *Element) Blocks(btype string) []*Element {
	block := elt.CheckTypeBlock()
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
	block := elt.MaybeBlock(btype, name)
	if block == nil {
		elt.AddMissingElementError(btype, ElementTypeBlock)
		return nil
	}

	return block
}

func (elt *Element) MaybeBlock(btype, name string) *Element {
	block := elt.CheckTypeBlock()
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

func (elt *Element) Entry(name string) *Element {
	entry := elt.MaybeEntry(name)
	if entry == nil {
		elt.AddMissingElementError(name, ElementTypeEntry)
		return nil
	}

	return entry
}

func (elt *Element) MaybeEntry(name string) *Element {
	block := elt.CheckTypeBlock()
	if block == nil {
		return nil
	}

	for _, child := range block.Elements {
		if entry, ok := child.Content.(*Entry); ok {
			if entry.Name == name {
				return child
			}
		}
	}

	return nil
}

func (elt *Element) CheckEntryMinValues(name string, min int) bool {
	entry := elt.Entry(name)
	if entry == nil {
		return false
	}

	return entry.CheckMinValues(min)
}

func (elt *Element) CheckMinValues(min int) bool {
	entry := elt.CheckTypeEntry()
	if entry == nil {
		return false
	}

	if len(entry.Values) < min {
		elt.AddInvalidEntryMinNbValuesError(min)
		return false
	}

	return true
}

func (elt *Element) EntryValues(name string, dests ...any) bool {
	entry := elt.Entry(name)
	if entry == nil {
		return false
	}

	return entry.Values(dests...)
}

func (elt *Element) MaybeEntryValues(name string, dests ...any) bool {
	entry := elt.MaybeEntry(name)
	if entry == nil {
		return true
	}

	return entry.Values(dests...)
}

func (elt *Element) Values(dests ...any) bool {
	entry := elt.CheckTypeEntry()
	if entry == nil {
		return false
	}

	if len(dests) == 1 {
		v := reflect.ValueOf(dests[0])

		if v.Kind() == reflect.Pointer && v.Elem().Kind() == reflect.Slice {
			t := v.Elem().Type().Elem()
			slice := reflect.MakeSlice(reflect.SliceOf(t), 0, len(entry.Values))

			valid := true

			for _, value := range entry.Values {
				value2 := reflect.New(t)

				err := value.Extract(value2.Interface())
				if err != nil {
					elt.AddInvalidValueError(value, err)
					valid = false
				}

				slice = reflect.Append(slice, value2.Elem())
			}

			if !valid {
				return false
			}

			v.Elem().Set(slice)
			return true
		}
	}

	if len(entry.Values) != len(dests) {
		elt.AddInvalidEntryNbValuesError(len(dests))
		return false
	}

	valid := true

	for i, value := range entry.Values {
		if err := value.Extract(dests[i]); err != nil {
			elt.AddInvalidValueError(value, err)
			valid = false
		}
	}

	return valid
}
