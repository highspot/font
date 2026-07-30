package sfnt

import "errors"

// ErrNotImplemented is returned when a stub API has no implementation yet.
var ErrNotImplemented = errors.New("not implemented")

// VariationAxis describes a single axis in the fvar table.
type VariationAxis struct {
	Tag           Tag
	MinValue      float64
	DefaultValue  float64
	MaxValue      float64
	NameID        uint16
	Flags         uint16
}

// TableFvar contains the font variations table, which defines the
// variation axes and named instances for a variable font.
// See https://learn.microsoft.com/en-us/typography/opentype/spec/fvar
type TableFvar struct {
	baseTable
	bytes []byte
}

func parseTableFvar(tag Tag, buf []byte) (Table, error) {
	return &TableFvar{
		baseTable: baseTable(tag),
		bytes:     buf,
	}, nil
}

func (t *TableFvar) Bytes() []byte {
	return t.bytes
}

// HasWidthAxis reports whether this table defines a width variation axis.
// Stub: not yet implemented.
func (t *TableFvar) HasWidthAxis() bool {
	return false
}

// WidthAxis returns the width ('wdth') variation axis if present.
// Stub: not yet implemented.
func (t *TableFvar) WidthAxis() (*VariationAxis, error) {
	return nil, ErrNotImplemented
}

// Axes returns all variation axes defined in this table.
// Stub: not yet implemented.
func (t *TableFvar) Axes() ([]VariationAxis, error) {
	return nil, ErrNotImplemented
}
