package bcl

import (
	"fmt"
	"math"
	"reflect"
)

type ValueType string

const (
	ValueTypeSymbol  ValueType = "symbol"
	ValueTypeBool    ValueType = "bool"
	ValueTypeString  ValueType = "string"
	ValueTypeInteger ValueType = "integer"
	ValueTypeFloat   ValueType = "float"
)

type ValueReader interface {
	ReadBCLValue(*Value) error
}

type Value struct {
	Location Span
	Content  any // either Symbol, bool, string, int64 or float64
}

type Symbol string

func (v *Value) Type() (t ValueType) {
	switch v.Content.(type) {
	case Symbol:
		t = ValueTypeSymbol
	case bool:
		t = ValueTypeBool
	case string:
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

func (v *Value) Extract(dest any) error {
	vt := v.Type()

	if vr, ok := dest.(ValueReader); ok {
		return vr.ReadBCLValue(v)
	}

	switch ptr := dest.(type) {
	case *bool:
		switch vt {
		case ValueTypeBool:
			*ptr = v.Content.(bool)
		default:
			return v.ValueTypeError(ValueTypeBool)
		}

	case *string:
		switch vt {
		case ValueTypeString:
			*ptr = v.Content.(string)
		case ValueTypeSymbol:
			*ptr = string(v.Content.(Symbol))
		default:
			return v.ValueTypeError(ValueTypeString, ValueTypeSymbol)
		}

	case *int:
		switch vt {
		case ValueTypeInteger:
			i := v.Content.(int64)
			min := int64(math.MinInt)
			max := int64(math.MaxInt)
			if i < min || i > max {
				return v.MinMaxIntegerValueError(min, max)
			}
			*ptr = int(i)
		default:
			return v.ValueTypeError(ValueTypeInteger)
		}

	case *int64:
		switch vt {
		case ValueTypeInteger:
			*ptr = v.Content.(int64)
		default:
			return v.ValueTypeError(ValueTypeInteger)
		}

	case *float64:
		switch vt {
		case ValueTypeFloat:
			*ptr = v.Content.(float64)
		case ValueTypeInteger:
			i := v.Content.(int64)
			min := int64(-1) << 53
			max := int64(1) << 53
			if i < min || i > max {
				return v.MinMaxIntegerValueError(min, max)
			}
			*ptr = float64(i)
		default:
			return v.ValueTypeError(ValueTypeFloat, ValueTypeInteger)
		}

	default:
		// Given a type T, there are two possible destination values:
		//
		// 1. A value of type *T if the caller wants to extract the BCL value to
		// a stack-allocated value.
		//
		// 2. A value of type **T if the caller wants to extract the BCL value to
		// a heap-allocated value (or in most cases because the value is
		// optional, hence the pointer type).
		//
		// 1 was handled at the beginning of the fonction (ReadBCLValue will
		// always have a pointer receiver).
		//
		// 2 is handled here.

		dv := reflect.ValueOf(dest)
		if dv.Kind() == reflect.Pointer && dv.Elem().Kind() == reflect.Pointer {
			dest2 := reflect.New(dv.Elem().Type().Elem())

			if vr, ok := dest2.Interface().(ValueReader); ok {
				if err := vr.ReadBCLValue(v); err != nil {
					return err
				}

				dv.Elem().Set(dest2)
				return nil
			}
		}

		panic(fmt.Sprintf("unhandled value destination of type %T", dest))
	}

	return nil
}

func (v *Value) ValueTypeError(expectedTypes ...ValueType) *InvalidValueTypeError {
	return &InvalidValueTypeError{Type: v.Type(), ExpectedTypes: expectedTypes}
}

func (v *Value) MinIntegerValueError(min int64) *MinIntegerValueError {
	return &MinIntegerValueError{Min: min}
}

func (v *Value) MaxIntegerValueError(max int64) *MaxIntegerValueError {
	return &MaxIntegerValueError{Max: max}
}

func (v *Value) MinMaxIntegerValueError(min, max int64) *MinMaxIntegerValueError {
	return &MinMaxIntegerValueError{Min: min, Max: max}
}
