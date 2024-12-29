package bcl

import (
	"bytes"
	"fmt"
)

type ValidationError struct {
	Err      error
	Location Span
}

type ValidationErrors struct {
	Errs  []ValidationError
	Lines []string
}

func (errs *ValidationErrors) Error() string {
	var buf bytes.Buffer

	for i, err := range errs.Errs {
		if i > 0 {
			buf.WriteByte('\n')
		}

		buf.WriteString(err.Err.Error())
		buf.WriteByte('\n')

		err.Location.PrintSource(&buf, errs.Lines)
	}

	return buf.String()
}

func (doc *Document) ValidationErrors() *ValidationErrors {
	var errs []ValidationError

	var walk func(*Element)
	walk = func(elt *Element) {
		for _, eltErr := range elt.validationErrors {
			eltErr2 := fmt.Errorf("invalid %s: %w", elt.Type(), eltErr)

			verr := ValidationError{
				Err:      eltErr2,
				Location: elt.Location,
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

func (elt *Element) AddValidationError(err error) {
	elt.validationErrors = append(elt.validationErrors, err)
}

func (elt *Element) AddSimpleValidationError(format string, args ...any) {
	elt.AddValidationError(&SimpleValidationError{
		Description: fmt.Sprintf(format, args...),
	})
}

func (elt *Element) AddMissingElementError(name string, expectedType ElementType) {
	elt.AddValidationError(&MissingElementError{
		Name:         name,
		ExpectedType: expectedType,
	})
}

func (elt *Element) AddInvalidElementTypeError(expectedType ElementType) {
	elt.AddValidationError(&InvalidElementTypeError{
		ExpectedType: expectedType,
	})
}

type SimpleValidationError struct {
	Description string
}

func (err *SimpleValidationError) Error() string {
	return err.Description
}

type MissingElementError struct {
	Name         string
	ExpectedType ElementType
}

func (err *MissingElementError) Error() string {
	return fmt.Sprintf("missing child %s %q", err.ExpectedType, err.Name)
}

type InvalidElementTypeError struct {
	ExpectedType ElementType
}

func (err *InvalidElementTypeError) Error() string {
	return fmt.Sprintf("element should be %s", err.ExpectedType.WithArticle())
}
