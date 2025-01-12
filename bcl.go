package bcl

import (
	"fmt"
	"io"
	"reflect"
	"slices"

	"maps"
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

func (elt *Element) IsBlock() bool {
	_, ok := elt.Content.(*Block)
	return ok
}

func (elt *Element) IsEntry() bool {
	_, ok := elt.Content.(*Entry)
	return ok
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

func (doc *Document) Block(btype string) *Element {
	return doc.TopLevel.Block(btype)
}

func (doc *Document) NamedBlock(btype, name string) *Element {
	return doc.TopLevel.NamedBlock(btype, name)
}

func (doc *Document) MaybeBlock(btype string) *Element {
	return doc.TopLevel.MaybeBlock(btype)
}

func (doc *Document) MaybeNamedBlock(btype, name string) *Element {
	return doc.TopLevel.MaybeNamedBlock(btype, name)
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

func (elt *Element) uniqueElementNames(eltType *ElementType, names []string) ([]string, bool) {
	block := elt.CheckTypeBlock()
	if block == nil {
		return nil, false
	}

	foundNames := make(map[string]struct{})

	for _, child := range block.Elements {
		if eltType != nil && *eltType != child.Type() {
			continue
		}

		switch content := child.Content.(type) {
		case *Block:
			if slices.Contains(names, content.Type) {
				foundNames[content.Type] = struct{}{}
			}
		case *Entry:
			if slices.Contains(names, content.Name) {
				foundNames[content.Name] = struct{}{}
			}
		}
	}

	return slices.Collect(maps.Keys(foundNames)), true
}

func (elt *Element) CheckElementsOneOf(names ...string) bool {
	foundNames, ok := elt.uniqueElementNames(nil, names)
	if !ok {
		return false
	}

	if len(foundNames) == 0 {
		elt.AddMissingElementError(nil, names)
		return false
	} else if len(foundNames) > 1 {
		elt.AddElementConflictError(nil, foundNames, names)
		return false
	}

	return true
}

func (elt *Element) CheckElementsMaybeOneOf(names ...string) bool {
	foundNames, ok := elt.uniqueElementNames(nil, names)
	if !ok {
		return false
	}

	if len(foundNames) > 1 {
		elt.AddElementConflictError(nil, foundNames, names)
		return false
	}

	return true
}

func (elt *Element) CheckBlocksOneOf(btypes ...string) bool {
	foundNames, ok := elt.uniqueElementNames(ref(ElementTypeBlock), btypes)
	if !ok {
		return false
	}

	if len(foundNames) == 0 {
		elt.AddMissingElementError(ref(ElementTypeBlock), btypes)
		return false
	} else if len(foundNames) > 1 {
		elt.AddElementConflictError(ref(ElementTypeBlock), foundNames, btypes)
		return false
	}

	return true
}

func (elt *Element) CheckBlocksMaybeOneOf(btypes ...string) bool {
	foundNames, ok := elt.uniqueElementNames(ref(ElementTypeBlock), btypes)
	if !ok {
		return false
	}

	if len(foundNames) > 1 {
		elt.AddElementConflictError(ref(ElementTypeBlock), foundNames, btypes)
		return false
	}

	return true
}

func (elt *Element) CheckEntriesOneOf(names ...string) bool {
	foundNames, ok := elt.uniqueElementNames(ref(ElementTypeEntry), names)
	if !ok {
		return false
	}

	if len(foundNames) == 0 {
		elt.AddMissingElementError(ref(ElementTypeEntry), names)
		return false
	} else if len(foundNames) > 1 {
		elt.AddElementConflictError(ref(ElementTypeEntry), foundNames, names)
		return false
	}

	return true
}

func (elt *Element) Element(name string) *Element {
	child := elt.MaybeElement(name)
	if child == nil {
		elt.AddMissingElementError(nil, []string{name})
		return nil
	}

	return child
}

func (elt *Element) MaybeElement(name string) *Element {
	block := elt.CheckTypeBlock()
	if block == nil {
		return nil
	}

	for _, child := range block.Elements {
		switch content := child.Content.(type) {
		case *Block:
			if content.Type == name {
				return child
			}
		case *Entry:
			if content.Name == name {
				return child
			}
		}
	}

	return nil
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

func (elt *Element) Block(btype string) *Element {
	return elt.NamedBlock(btype, "")
}

func (elt *Element) NamedBlock(btype, name string) *Element {
	block := elt.MaybeNamedBlock(btype, name)
	if block == nil {
		elt.AddMissingElementError(ref(ElementTypeBlock), []string{btype})
		return nil
	}

	return block
}

func (elt *Element) MaybeBlock(btype string) *Element {
	return elt.MaybeNamedBlock(btype, "")
}

func (elt *Element) MaybeNamedBlock(btype, name string) *Element {
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

func (elt *Element) BlockName() string {
	block := elt.CheckTypeBlock()
	if block == nil {
		return ""
	}

	if block.Name == "" {
		elt.AddMissingBlockNameError()
	}

	return block.Name
}

func (elt *Element) Entries(name string) []*Element {
	block := elt.CheckTypeBlock()
	if block == nil {
		return nil
	}

	var entries []*Element

	for _, child := range block.Elements {
		if entry, ok := child.Content.(*Entry); ok {
			if entry.Name == name {
				entries = append(entries, child)
			}
		}
	}

	return entries
}

func (elt *Element) Entry(name string) *Element {
	entry := elt.MaybeEntry(name)
	if entry == nil {
		elt.AddMissingElementError(ref(ElementTypeEntry), []string{name})
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

func (elt *Element) CheckEntryMinMaxValues(name string, min, max int) bool {
	entry := elt.Entry(name)
	if entry == nil {
		return false
	}

	return entry.CheckMinMaxValues(min, max)
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

func (elt *Element) CheckMinMaxValues(min, max int) bool {
	entry := elt.CheckTypeEntry()
	if entry == nil {
		return false
	}

	if len(entry.Values) < min {
		elt.AddInvalidEntryMinMaxNbValuesError(min, max)
		return false
	}

	if len(entry.Values) > max {
		elt.AddInvalidEntryMinMaxNbValuesError(min, max)
		return false
	}

	return true
}

func (elt *Element) NbValues() int {
	entry := elt.CheckTypeEntry()
	if entry == nil {
		return -1
	}

	return len(entry.Values)
}

func (elt *Element) EntryValues(name string, dests ...any) bool {
	entry := elt.Entry(name)
	if entry == nil {
		return false
	}

	return entry.Values(dests...)
}

func (elt *Element) EntryValue(name string, dest any) bool {
	entry := elt.Entry(name)
	if entry == nil {
		return false
	}

	return entry.Values(dest)
}

func (elt *Element) MaybeEntryValues(name string, dests ...any) bool {
	entry := elt.MaybeEntry(name)
	if entry == nil {
		return true
	}

	return entry.Values(dests...)
}

func (elt *Element) MaybeEntryValue(name string, dest any) bool {
	entry := elt.MaybeEntry(name)
	if entry == nil {
		return true
	}

	return entry.Values(dest)
}

func (elt *Element) Value(dest any) bool {
	return elt.Values(dest)
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
