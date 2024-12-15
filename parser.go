package bcl

import (
	"errors"
	"fmt"
)

type parser struct {
	source   string
	data     []byte
	lines    []string
	tokens   []*Token
	endPoint Point
	doc      *Document
}

func newParser(data []byte, source string) *parser {
	return &parser{
		source: source,
		data:   data,
		lines:  splitLines(data),

		doc: &Document{},
	}
}

func (p *parser) Parse() (doc *Document, err error) {
	defer func() {
		if v := recover(); v != nil {
			if verr, ok := v.(error); ok {
				var syntaxErr *SyntaxError

				if errors.As(verr, &syntaxErr) {
					syntaxErr.Lines = p.lines
					err = ParseErrors{verr}
					return
				}
			}

			panic(v)
		}
	}()

	tokenizer := newTokenizer(p.data, p.source)
	tokens := []*Token{}

	for {
		token := tokenizer.readToken()
		if token == nil {
			break
		}

		tokens = append(tokens, token)
	}

	p.tokens = tokens

	if len(tokens) > 0 {
		p.endPoint = tokens[len(tokens)-1].Span.End
	}

	p.doc = new(Document)

	for {
		elt := p.parseElement()
		if elt == nil {
			break
		}

		p.doc.Elements = append(p.doc.Elements, elt)
	}

	doc = p.doc
	return
}

func (p *parser) tokenSyntaxError(token *Token, format string, args ...any) error {
	return p.syntaxErrorAt(token.Span, format, args...)
}

func (p *parser) syntaxErrorAtPoint(point Point, format string, args ...any) error {
	return p.syntaxErrorAt(Span{point, point}, format, args...)
}

func (p *parser) syntaxErrorAt(span Span, format string, args ...any) error {
	return &SyntaxError{
		Source:      p.source,
		Location:    span,
		Description: fmt.Sprintf(format, args...),
	}
}

func (p *parser) peekToken() *Token {
	if len(p.tokens) == 0 {
		return nil
	}

	return p.tokens[0]
}

func (p *parser) readToken() *Token {
	token := p.peekToken()
	if token == nil {
		return nil
	}

	p.skipToken()
	return token
}

func (p *parser) skipToken() {
	p.tokens = p.tokens[1:]
}

func (p *parser) skipEOL() {
	for len(p.tokens) > 0 {
		token := p.peekToken()
		if token.Type != TokenTypeEOL {
			break
		}

		p.tokens = p.tokens[1:]
	}
}

func (p *parser) parseElement() Element {
	p.skipEOL()
	nameToken := p.readToken()
	if nameToken == nil {
		return nil
	}

	if nameToken.Type != TokenTypeSymbol {
		panic(p.tokenSyntaxError(nameToken, "invalid token, expected block "+
			"name or entry name"))
	}

	p.skipEOL()
	token := p.readToken()

	if token == nil {
		panic(p.tokenSyntaxError(nameToken, "truncated element"))
	}

	if token.Type == TokenTypeOpeningBracket {
		elts := p.parseBlockContent()

		return &Block{
			Name:     nameToken.Value.(string),
			Elements: elts,
		}
	}

	values := p.parseEntryValues()

	return &Entry{
		Name:   nameToken.Value.(string),
		Values: append([]Value{p.tokenValue(token)}, values...),
	}
}

func (p *parser) parseBlockContent() []Element {
	var elts []Element

	for {
		p.skipEOL()
		token := p.peekToken()
		if token == nil {
			panic(p.syntaxErrorAtPoint(p.endPoint, "truncated block"))
		}

		if token.Type == TokenTypeClosingBracket {
			p.skipToken()
			break
		}

		elt := p.parseElement()
		if elt == nil {
			panic(p.syntaxErrorAtPoint(p.endPoint, "truncated block"))
		}

		elts = append(elts, elt)
	}

	return elts
}

func (p *parser) parseEntryValues() []Value {
	var values []Value

	for {
		token := p.readToken()
		if token == nil {
			panic(p.syntaxErrorAtPoint(p.endPoint, "truncated entry"))
		}

		if token.Type == TokenTypeEOL {
			break
		}

		values = append(values, p.tokenValue(token))
	}

	return values
}

func (p *parser) tokenValue(t *Token) Value {
	switch t.Type {
	case TokenTypeSymbol:
		s := t.Value.(string)
		switch s {
		case "true":
			return true
		case "false":
			return false
		case "null":
			return nil
		default:
			return Symbol(t.Value.(string))
		}

	case TokenTypeString:
		return t.Value.(string)

	case TokenTypeInteger:
		return t.Value.(int64)

	case TokenTypeFloat:
		return t.Value.(float64)

	default:
		panic(p.tokenSyntaxError(t, "invalid token, expected symbol, "+
			"string, integer or float"))
	}
}
