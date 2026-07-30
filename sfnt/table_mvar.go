package sfnt

// TableMvar contains the metrics variations table, which describes
// how font-wide metrics vary across the font's variation axes.
// See https://learn.microsoft.com/en-us/typography/opentype/spec/mvar
type TableMvar struct {
	baseTable
	bytes []byte
}

func parseTableMvar(tag Tag, buf []byte) (Table, error) {
	return &TableMvar{
		baseTable: baseTable(tag),
		bytes:     buf,
	}, nil
}

func (t *TableMvar) Bytes() []byte {
	return t.bytes
}
