package bcl

import (
	"fmt"
)

type ValueType string

const (
	ValueTypeSymbol  ValueType = "symbol"
	ValueTypeBool    ValueType = "bool"
	ValueTypeString  ValueType = "string"
	ValueTypeInteger ValueType = "integer"
	ValueTypeFloat   ValueType = "float"
)

type Value struct {
	Location Span
	Content  any // either Symbol, bool, String, int64 or float64
}

func (v *Value) Type() (t ValueType) {
	switch v.Content.(type) {
	case Symbol:
		t = ValueTypeSymbol
	case bool:
		t = ValueTypeBool
	case String:
		t = ValueTypeString
	case int64:
		t = ValueTypeInteger
	case float64:
		t = ValueTypeFloat

	default:
		panic(fmt.Sprintf("unhandled value %#v (%T)", v.Content, v.Content))
	}

	return
}

type Symbol string

type String struct {
	String string
	Sigil  string
}

func (v *Value) IsOneOf(contents ...any) error {
	valid := false
	t := v.Type()

	for _, content := range contents {
		switch c := content.(type) {
		case bool:
			if t == ValueTypeBool && v.Content.(bool) == c {
				valid = true
				break
			}

		case string:
			if t == ValueTypeString && v.Content.(String).String == c {
				valid = true
				break
			}

			if t == ValueTypeSymbol && string(v.Content.(Symbol)) == c {
				valid = true
				break
			}

		case int:
			if t == ValueTypeInteger && v.Content.(int64) == int64(c) {
				valid = true
				break
			}

		case int64:
			if t == ValueTypeInteger && v.Content.(int64) == c {
				valid = true
				break
			}

		case float64:
			if t == ValueTypeFloat && v.Content.(float64) == c {
				valid = true
				break
			}
		}
	}

	if !valid {
		return NewValueContentError(v, contents...)
	}

	return nil
}

func (v1 *Value) Equal(v2 *Value) bool {
	t := v1.Type()
	if t != v2.Type() {
		return false
	}

	eq := false

	switch t {
	case ValueTypeSymbol:
		eq = v1.Content.(Symbol) == v2.Content.(Symbol)
	case ValueTypeBool:
		eq = v1.Content.(bool) == v2.Content.(bool)
	case ValueTypeString:
		s1 := v1.Content.(String)
		s2 := v2.Content.(String)
		eq = s1.Sigil == s2.Sigil && s1.String == s2.String
	case ValueTypeInteger:
		eq = v1.Content.(int64) == v2.Content.(int64)
	case ValueTypeFloat:
		eq = v1.Content.(float64) == v2.Content.(float64)
	default:
		panic(fmt.Sprintf("unhandled value type %q", t))
	}

	return eq
}
