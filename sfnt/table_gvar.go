package sfnt

// TableGvar contains the glyph variations table, which defines
// how TrueType glyph outlines vary across the font's variation axes.
// See https://learn.microsoft.com/en-us/typography/opentype/spec/gvar
type TableGvar struct {
	baseTable
	bytes []byte
}

func parseTableGvar(tag Tag, buf []byte) (Table, error) {
	return &TableGvar{
		baseTable: baseTable(tag),
		bytes:     buf,
	}, nil
}

func (t *TableGvar) Bytes() []byte {
	return t.bytes
}
