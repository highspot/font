package sfnt

import (
	"bytes"
	"encoding/binary"
	"go/types"
	"unsafe"
)

type FvarHeader struct {
	MajorVersion   uint16
	MinorVersion   uint16
	OffsetToData   uint16
	CountSizePairs uint16
	AxisCount      uint16
	AxisSize       uint16
	InstanceCount  uint16
	InstanceSize   uint16
}

type VariationAxis struct {
	AxisTag      uint32
	MinValue     fixed
	DefaultValue fixed
	MaxValue     fixed
	Flags        uint16
	NameID       NameID
}

type Instance struct {
	NameID   NameID
	Flags    uint16
	Coord    []*fixed
	PsNameID *NameID
}

type InstanceWithoutPSName struct {
	NameID NameID
	Flags  uint16
	Coord  []*fixed
}

// TableFvar represents the OpenType 'fvar' table.
// https://developer.apple.com/fonts/TrueType-Reference-Manual/RM06/Chap6fvar.html
type TableFvar struct {
	baseTable
	Header   FvarHeader
	Axis     []*VariationAxis
	Instance []*Instance

	bytes []byte
}

func (t *TableFvar) Bytes() []byte {
	return t.bytes
}

func parseTableFvar(tag Tag, buf []byte) (Table, error) {
	r := bytes.NewBuffer(buf)

	var header FvarHeader
	if err := binary.Read(r, binary.BigEndian, &header); err != nil {
		return nil, err
	}

	table := &TableFvar{
		baseTable: baseTable(tag),
		bytes:     buf,
		Header:    header,
		Axis:      make([]*VariationAxis, 0, header.AxisCount),
		Instance:  make([]*Instance, 0, header.InstanceCount),
	}

	for i := 0; i < int(header.AxisCount); i++ {
		var axis VariationAxis
		if err := binary.Read(r, binary.BigEndian, &axis); err != nil {
			return nil, err
		}
		table.Axis = append(table.Axis, &VariationAxis{
			axis.AxisTag,
			axis.MinValue,
			axis.DefaultValue,
			axis.MaxValue,
			axis.Flags,
			axis.NameID,
		})
	}

	noPsNameIdVariant := 2 * uint16(unsafe.Sizeof(types.Uint16))
	invariant := uint16(unsafe.Sizeof(types.Int16)) + uint16(unsafe.Sizeof(types.Uint16))

	for i := 0; i < int(header.InstanceCount); i++ {

		if header.InstanceSize == (noPsNameIdVariant + invariant) {
			var instance InstanceWithoutPSName
			if err := binary.Read(r, binary.BigEndian, &instance); err != nil {
				return nil, err
			}

			table.Instance = append(table.Instance, &Instance{
				instance.NameID,
				instance.Flags,
				instance.Coord,
				nil,
			})

		} else {
			var instance Instance
			if err := binary.Read(r, binary.BigEndian, &instance); err != nil {
				return nil, err
			}

			table.Instance = append(table.Instance, &Instance{
				instance.NameID,
				instance.Flags,
				instance.Coord,
				instance.PsNameID,
			})
		}
	}

	return table, nil
}
