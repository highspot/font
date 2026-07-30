package sfnt

// TableHvar contains the horizontal metrics variations table, which
// describes how horizontal glyph metrics vary across variation axes.
// See https://learn.microsoft.com/en-us/typography/opentype/spec/hvar
type TableHvar struct {
	baseTable
	bytes []byte
}

func parseTableHvar(tag Tag, buf []byte) (Table, error) {
	return &TableHvar{
		baseTable: baseTable(tag),
		bytes:     buf,
	}, nil
}

func (t *TableHvar) Bytes() []byte {
	return t.bytes
}
