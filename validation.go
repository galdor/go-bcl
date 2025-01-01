package bcl

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type ValidationError struct {
	Err      error
	Location *Span
}

type ValidationErrors struct {
	Errs  []ValidationError
	Lines []string
}

func (errs *ValidationErrors) Error() string {
	var buf bytes.Buffer

	for _, err := range errs.Errs {
		buf.WriteString("  - ")
		buf.WriteString(err.Err.Error())
		buf.WriteByte('\n')

		if err.Location != nil {
			err.Location.PrintSource(&buf, errs.Lines, "      ")
		}
	}

	return strings.TrimRight(buf.String(), "\n")
}

func (doc *Document) ValidationErrors() *ValidationErrors {
	var errs []ValidationError

	var walk func(*Element)
	walk = func(elt *Element) {
		for _, eltErr := range elt.validationErrors {
			var eltErr2 error
			if elt == doc.TopLevel {
				eltErr2 = eltErr
			} else {
				eltErr2 = fmt.Errorf("invalid %s: %w", elt.Type(), eltErr)
			}

			verr := ValidationError{
				Err: eltErr2,
			}

			if elt != doc.TopLevel {
				verr.Location = &elt.Location
			}

			var invalidValueErr *InvalidValueError
			if errors.As(eltErr, &invalidValueErr) {
				verr.Location = &invalidValueErr.Value.Location
			}

			errs = append(errs, verr)
		}

		if block, ok := elt.Content.(*Block); ok {
			for _, elt2 := range block.Elements {
				walk(elt2)
			}
		}
	}

	walk(doc.TopLevel)

	if len(errs) == 0 {
		return nil
	}

	return &ValidationErrors{
		Errs:  errs,
		Lines: doc.lines,
	}
}

func (elt *Element) AddValidationError(err error) error {
	elt.validationErrors = append(elt.validationErrors, err)
	return err
}

type SimpleValidationError struct {
	Description string
}

func (err *SimpleValidationError) Error() string {
	return err.Description
}

func (elt *Element) AddSimpleValidationError(format string, args ...any) error {
	return elt.AddValidationError(&SimpleValidationError{
		Description: fmt.Sprintf(format, args...),
	})
}

type MissingElementError struct {
	Name         string
	ExpectedType ElementType
}

func (err *MissingElementError) Error() string {
	return fmt.Sprintf("missing child %s %q", err.ExpectedType, err.Name)
}

func (elt *Element) AddMissingElementError(name string, expectedType ElementType) error {
	return elt.AddValidationError(&MissingElementError{
		Name:         name,
		ExpectedType: expectedType,
	})
}

type InvalidElementTypeError struct {
	ExpectedType ElementType
}

func (err *InvalidElementTypeError) Error() string {
	return fmt.Sprintf("element should be %s",
		WordWithArticle(string(err.ExpectedType)))
}

func (elt *Element) AddInvalidElementTypeError(expectedType ElementType) error {
	return elt.AddValidationError(&InvalidElementTypeError{
		ExpectedType: expectedType,
	})
}

type InvalidEntryNbValuesError struct {
	NbValues         int
	ExpectedNbValues []int
}

func (err *InvalidEntryNbValuesError) Error() string {
	ns := make([]string, len(err.ExpectedNbValues))
	for i, n := range err.ExpectedNbValues {
		ns[i] = strconv.Itoa(n)
	}

	return fmt.Sprintf("entry has %d %s but should have %s %s",
		err.NbValues, PluralizeWord("value", err.NbValues),
		WordsEnumerationOr(ns), PluralizeWord("value", len(ns)))
}

func (elt *Element) AddInvalidEntryNbValuesError(expectedNbValues ...int) error {
	return elt.AddValidationError(&InvalidEntryNbValuesError{
		NbValues:         len(elt.Content.(*Entry).Values),
		ExpectedNbValues: expectedNbValues,
	})
}

type InvalidEntryMinNbValuesError struct {
	NbValues int
	Min      int
}

func (err *InvalidEntryMinNbValuesError) Error() string {
	return fmt.Sprintf("entry has %d %s but should have at least %d %s",
		err.NbValues, PluralizeWord("value", err.NbValues),
		err.Min, PluralizeWord("value", err.Min))
}

func (elt *Element) AddInvalidEntryMinNbValuesError(min int) error {
	return elt.AddValidationError(&InvalidEntryMinNbValuesError{
		NbValues: len(elt.Content.(*Entry).Values),
		Min:      min,
	})
}

type InvalidValueError struct {
	Value *Value
	Err   error
}

func (err *InvalidValueError) Unwrap() error {
	return err.Err
}

func (err *InvalidValueError) Error() string {
	return err.Err.Error()
}

func (elt *Element) AddInvalidValueError(v *Value, err error) error {
	return elt.AddValidationError(&InvalidValueError{
		Value: v,
		Err:   err,
	})
}

type InvalidValueTypeError struct {
	Type          ValueType
	ExpectedTypes []ValueType
}

func (err *InvalidValueTypeError) Error() string {
	etWithArticles := make([]string, len(err.ExpectedTypes))
	for i, et := range err.ExpectedTypes {
		etWithArticles[i] = WordWithArticle(string(et))
	}

	return fmt.Sprintf("value is %s but should be %s",
		WordWithArticle(string(err.Type)), WordsEnumerationOr(etWithArticles))
}

type MinIntegerValueError struct {
	Min int64
}

func (err *MinIntegerValueError) Error() string {
	return fmt.Sprintf("integer must be greater or equal to %d", err.Min)
}

type MaxIntegerValueError struct {
	Max int64
}

func (err *MaxIntegerValueError) Error() string {
	return fmt.Sprintf("integer must be lower or equal to %d", err.Max)
}

type MinMaxIntegerValueError struct {
	Min int64
	Max int64
}

func (err *MinMaxIntegerValueError) Error() string {
	return fmt.Sprintf("integer must be between %d and %d", err.Min, err.Max)
}
