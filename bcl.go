package bcl

type Document struct {
	Source   string
	Elements []*Element
}

type Element struct {
	Location Span
	Content  any // *Block or *Entry
}

type Block struct {
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
