package main

import (
	"encoding/json"
	"fmt"
	"github.com/iancoleman/strcase"
	"strconv"

	"github.com/ConradIrwin/font/sfnt"
)

// Info prints the name table (contains metadata).
func Info(font *sfnt.Font) error {

	data := make(map[sfnt.NameID]string)
	output := make(map[string]interface{})

	if font.HasTable(sfnt.TagName) {
		name, err := font.NameTable()
		if err != nil {
			return err
		}
		for _, entry := range name.List() {
			if entry.NameID > 255 {
				data[entry.NameID] = entry.String()
			} else {
				output[strcase.ToSnake(entry.Label())] = entry.String()
			}
		}
	}

	if font.HasTable(sfnt.TagFvar) {
		fvar, err := font.FvarTable()
		if err != nil {
			return err
		}
		for i := 0; i < int(fvar.Header.AxisCount); i++ {
			a := fvar.Axis[i]
			if data[a.NameID] == "wght" {
				output["font_weight"] = strconv.Itoa(int(a.DefaultValue.Major))
			}
		}
	}

	if font.HasTable(sfnt.TagOS2) {
		os2, err := font.OS2Table()
		if err != nil {
			return err
		}
		output["font_weight"] = os2.WeightClass()
		output["font_width"] = os2.WidthClass()
		output["font_style"] = os2.FontStyle()
		output["unicode_range"] = os2.UnicodeRanges()
	}

	marshal, err := json.MarshalIndent(output, " ", " ")
	if err != nil {
		return err
	}

	fmt.Println(string(marshal))
	return nil
}
