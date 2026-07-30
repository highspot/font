package sfnt

// TableStat contains the style attributes table, which describes
// design axes and their values for variable and non-variable fonts.
// Required for variable fonts.
// See https://learn.microsoft.com/en-us/typography/opentype/spec/stat
type TableStat struct {
	baseTable
	bytes []byte
}

func parseTableStat(tag Tag, buf []byte) (Table, error) {
	return &TableStat{
		baseTable: baseTable(tag),
		bytes:     buf,
	}, nil
}

func (t *TableStat) Bytes() []byte {
	return t.bytes
}
