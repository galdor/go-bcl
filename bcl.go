package bcl

import (
	"fmt"
	"io"
)

type Document struct {
	Source   string
	TopLevel *Block
}

type Element struct {
	Location            Span
	Content             any // *Block or *Entry
	FollowedByEmptyLine bool
}

type Block struct {
	Type     string
	Name     string
	Elements []*Element
}

type Entry struct {
	Name   string
	Values []Value
}

// Either Symbol, bool, string, int64 or float64
type Value interface {
}

type Symbol string

func Parse(data []byte, source string) (*Document, error) {
	p := newParser(data, source)
	return p.Parse()
}

func (doc *Document) Print(w io.Writer) error {
	p := newPrinter(w, doc)
	return p.Print()
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

func (elt *Element) ContentTypeName() (name string) {
	switch elt.Content.(type) {
	case *Block:
		name = "block"
	case *Entry:
		name = "entry"
	default:
		panic(fmt.Sprintf("unhandled element content %#v (%T)", elt, elt))
	}

	return
}

func (doc *Document) Blocks(btype string) []*Block {
	return doc.TopLevel.Blocks(btype)
}

func (doc *Document) Block(btype, name string) *Block {
	return doc.TopLevel.Block(btype, name)
}

func (b *Block) Blocks(btype string) []*Block {
	var blocks []*Block

	for _, elt := range b.Elements {
		if block, ok := elt.Content.(*Block); ok {
			if block.Type == btype {
				blocks = append(blocks, block)
			}
		}
	}

	return blocks
}

func (b *Block) Block(btype, name string) *Block {
	for _, elt := range b.Elements {
		if block, ok := elt.Content.(*Block); ok {
			if block.Type == btype && block.Name == name {
				return block
			}
		}
	}

	return nil
}
