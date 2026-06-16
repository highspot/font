package sfnt

// TableAvar contains the axis variations table, which provides
// optional fine-tuning of the variation space by remapping axis values.
// See https://learn.microsoft.com/en-us/typography/opentype/spec/avar
type TableAvar struct {
	baseTable
	bytes []byte
}

func parseTableAvar(tag Tag, buf []byte) (Table, error) {
	return &TableAvar{
		baseTable: baseTable(tag),
		bytes:     buf,
	}, nil
}

func (t *TableAvar) Bytes() []byte {
	return t.bytes
}
