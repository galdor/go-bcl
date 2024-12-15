package bcl

type Document struct {
	Elements []Element
}

// Either *Block or *Entry
type Element interface {
}

type Block struct {
	Name     string
	Elements []Element
}

type Entry struct {
	Name   string
	Values []Value
}

type Value interface {
}

type Symbol string

func Tokenize(data []byte, source string) ([]*Token, error) {
	t := newTokenizer(data, source)
	return t.Tokenize()
}

func Parse(data []byte, source string) (*Document, error) {
	p := newParser(data, source)
	return p.Parse()
}
