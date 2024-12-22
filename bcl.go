package bcl

import (
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
	Name     string
	Label    string
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

func (doc *Document) Blocks(name string) []*Block {
	return doc.TopLevel.Blocks(name)
}

func (doc *Document) Block(name, label string) *Block {
	return doc.TopLevel.Block(name, label)
}

func (b *Block) Blocks(name string) []*Block {
	var blocks []*Block

	for _, elt := range b.Elements {
		if block, ok := elt.Content.(*Block); ok {
			if block.Name == name {
				blocks = append(blocks, block)
			}
		}
	}

	return blocks
}

func (b *Block) Block(name, label string) *Block {
	for _, elt := range b.Elements {
		if block, ok := elt.Content.(*Block); ok {
			if block.Name == name && block.Label == label {
				return block
			}
		}
	}

	return nil
}
