package sfnt

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
