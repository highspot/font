package sfnt

import (
	"bytes"
	"encoding/binary"
	"math"
	"strconv"
)

type Panose struct {
	BFamilyType      uint8
	BSerifStyle      uint8
	BWeight          uint8
	BProportion      uint8
	BContrast        uint8
	BStrokeVariation uint8
	BArmStyle        uint8
	BLetterform      uint8
	BMidline         uint8
	BXHeight         uint8
}

type TableOs2Original struct {
	Version             uint16
	XAvgCharWidth       int16
	USWeightClass       uint16
	USWidthClass        uint16
	FSType              int16
	YSubscriptXSize     int16
	YSubscriptYSize     int16
	YSubscriptXOffset   int16
	YSubscriptYOffset   int16
	YSuperscriptXSize   int16
	YSuperscriptYSize   int16
	YSuperscriptXOffset int16
	YSuperscriptYOffset int16
	YStrikeoutSize      int16
	YStrikeoutPosition  int16
	SFamilyClass        int16
	Panose              Panose
	UlUnicodeRange      [4]uint32
	AchVendID           Tag
	FsSelection         uint16
	FsFirstCharIndex    uint16
	FsLastCharIndex     uint16
}

type TableOs2AdditionalFields struct {
	STypoAscender    int16
	STypoDescender   int16
	STypoLineGap     int16
	UsWinAscent      uint16
	UsWinDescent     uint16
	UlCodePageRange1 uint32
	UlCodePageRange2 uint32
	SxHeigh          int16
	SCapHeight       int16
	UsDefaultChar    uint16
	UsBreakChar      uint16
	UsMaxContext     uint16
	UsLowerPointSize uint16
	UsUpperPointSize uint16
}

type TableOS2 struct {
	baseTable
	TableOs2Original
	TableOs2AdditionalFields
	bytes []byte
}

func parseTableOS2(tag Tag, buf []byte) (Table, error) {
	r := bytes.NewBuffer(buf)

	var originalTable TableOs2Original
	if err := binary.Read(r, binary.BigEndian, &originalTable); err != nil {
		return nil, err
	}

	tableOs2 := &TableOS2{
		baseTable:        baseTable(tag),
		TableOs2Original: originalTable,
		bytes:            buf,
	}

	// TODO: There may be additional fields
	// The number of remaining fields varies by
	// font type and version.

	return tableOs2, nil
}

func (t *TableOS2) Bytes() []byte {
	return t.bytes
}

func (t *TableOS2) WeightClass() string {
	switch t.USWeightClass {
	case 1:
		return "Ultra-light"
	case 2:
		return "Extra-light"
	case 3:
		return "Light"
	case 4:
		return "Semi-light"
	case 5:
		return "Medium"
	case 6:
		return "Semi-bold"
	case 7:
		return "Bold"
	case 8:
		return "Extra-bold"
	case 9:
		return "Ultra-bold"
	default:
		// This is the actual weight value
		return strconv.Itoa(int(t.USWeightClass))
	}
}

func (t *TableOS2) WidthClass() string {
	switch t.USWidthClass {
	case 1:
		return "Ultra-condensed"
	case 2:
		return "Extra-condensed"
	case 3:
		return "Condensed"
	case 4:
		return "Semi-condensed"
	case 5:
		return "Medium"
	case 6:
		return "Semi-expanded"
	case 7:
		return "Expanded"
	case 8:
		return "Extra-expanded"
	case 9:
		return "Ultra-expanded"
	default:
		return "Unknown"
	}
}

func (t *TableOS2) IsBold() bool {
	const i = 0b00100000
	mask := i
	return (uint16(mask) & t.FsSelection) == uint16(32)
}

func (t *TableOS2) IsStrikeout() bool {
	mask := 0b00010000
	return (uint16(mask) & t.FsSelection) == uint16(16)
}

func (t *TableOS2) IsOutlined() bool {
	mask := 0b00001000
	return (uint16(mask) & t.FsSelection) == uint16(8)
}

func (t *TableOS2) IsNegative() bool {
	mask := 0b00000100
	return (uint16(mask) & t.FsSelection) == uint16(4)
}

func (t *TableOS2) IsUnderscore() bool {
	mask := 0b00000010
	return (uint16(mask) & t.FsSelection) == uint16(2)
}

func (t *TableOS2) IsItalic() bool {
	mask := 0b00000001
	return (uint16(mask) & t.FsSelection) == uint16(1)
}

func (t *TableOS2) FontStyle() string {
	if t.IsItalic() {
		return "italic"
	} else if t.IsBold() {
		return "bold"
	} else if t.IsUnderscore() {
		return "underscore"
	} else if t.IsOutlined() {
		return "outlined"
	} else if t.IsStrikeout() {
		return "strikeout"
	} else if t.IsNegative() {
		return "negative"
	}

	return "normal"
}

func (t *TableOS2) UnicodeRanges() []string {
	ranges := make([]string, 0)
	for _, unicodeRange := range SupportedUnicodeRanges {
		isSupported := unicodeRange.isSupported(t.UlUnicodeRange)
		if isSupported {
			for _, r := range unicodeRange.UnicodeRanges {
				ranges = append(ranges, "U+"+r)
			}
		}
	}
	return ranges
}

type UnicodeSupport struct {
	BitIndex      int
	Name          string
	UnicodeRanges []string
}

func (s UnicodeSupport) isSupported(unicodeRangeMask [4]uint32) bool {
	applicableBitMaskIndex := s.BitIndex / 32
	exponentValue := s.BitIndex % 32
	return (uint32(math.Exp2(float64(exponentValue))) & unicodeRangeMask[applicableBitMaskIndex]) != 0
}

// Definitions based on https://docs.microsoft.com/en-us/typography/opentype/spec/os2#ur
var (
	BasicLatin                = UnicodeSupport{BitIndex: 0, Name: "Basic Latin", UnicodeRanges: []string{"0000-007F"}}
	Latin1Supplement          = UnicodeSupport{BitIndex: 1, Name: "Latin-1 Supplement", UnicodeRanges: []string{"0080-00FF"}}
	LatinExtendedA            = UnicodeSupport{BitIndex: 2, Name: "Latin Extended-A", UnicodeRanges: []string{"0100-017F"}}
	LatinExtendedB            = UnicodeSupport{BitIndex: 3, Name: "Latin Extended-B", UnicodeRanges: []string{"0180-024F"}}
	IPAPhonetic               = UnicodeSupport{BitIndex: 4, Name: "IPA & Phonetic Extensions/Supplements", UnicodeRanges: []string{"0250-02AF", "1D00-1D7F", "1D80-1DBF"}}
	SpacingModifierLetters    = UnicodeSupport{BitIndex: 5, Name: "Spacing Modifier Letters", UnicodeRanges: []string{"02B0-02FF", "A700-A71F"}}
	CombiningDiacriticalMarks = UnicodeSupport{BitIndex: 6, Name: "Combining Diacritical Marks & Supplements", UnicodeRanges: []string{"0300-036F", "1DC0-1DFF"}}
	GreekCoptic               = UnicodeSupport{BitIndex: 7, Name: "Greek and Coptic", UnicodeRanges: []string{"0370-03FF"}}
	Coptic                    = UnicodeSupport{BitIndex: 8, Name: "Coptic", UnicodeRanges: []string{"2C80-2CFF"}}
	Cyrillic                  = UnicodeSupport{BitIndex: 9, Name: "Cyrillic", UnicodeRanges: []string{"0400-04FF", "0500-052F", "2DE0-2DFF", "A640-A69F"}}
	Armenian                  = UnicodeSupport{BitIndex: 10, Name: "Armenian", UnicodeRanges: []string{"0530-058F"}}
	Hebrew                    = UnicodeSupport{BitIndex: 11, Name: "Hebrew", UnicodeRanges: []string{"0590-05FF"}}
	Vai                       = UnicodeSupport{BitIndex: 12, Name: "Vai", UnicodeRanges: []string{"A500-A63F"}}
	Arabic                    = UnicodeSupport{BitIndex: 13, Name: "Arabic", UnicodeRanges: []string{"0600-06FF", "0750-077F"}}
	NKo                       = UnicodeSupport{BitIndex: 14, Name: "NKo", UnicodeRanges: []string{"07C0-07FF"}}
	Devanagari                = UnicodeSupport{BitIndex: 15, Name: "Devanagari", UnicodeRanges: []string{"0900-097F"}}
	Bengali                   = UnicodeSupport{BitIndex: 16, Name: "Bengali", UnicodeRanges: []string{"0980-09FF"}}
	Gurmukhi                  = UnicodeSupport{BitIndex: 17, Name: "Gurmukhi", UnicodeRanges: []string{"0A00-0A7F"}}
	Gujarati                  = UnicodeSupport{BitIndex: 18, Name: "Gujarati", UnicodeRanges: []string{"0A80-0AFF"}}
	Oriya                     = UnicodeSupport{BitIndex: 19, Name: "Oriya", UnicodeRanges: []string{"0B00-0B7F"}}
	Tamil                     = UnicodeSupport{BitIndex: 20, Name: "Tamil", UnicodeRanges: []string{"0B80-0BFF"}}
	Telugu                    = UnicodeSupport{BitIndex: 21, Name: "Telugu", UnicodeRanges: []string{"0C00-0C7F"}}
	Kannada                   = UnicodeSupport{BitIndex: 22, Name: "Kannada", UnicodeRanges: []string{"0C80-0CFF"}}
	Malayalam                 = UnicodeSupport{BitIndex: 23, Name: "Malayalam", UnicodeRanges: []string{"0D00-0D7F"}}
	Thai                      = UnicodeSupport{BitIndex: 24, Name: "Thai", UnicodeRanges: []string{"0E00-0E7F"}}
	Lao                       = UnicodeSupport{BitIndex: 25, Name: "Lao", UnicodeRanges: []string{"0E80-0EFF"}}
	Georgian                  = UnicodeSupport{BitIndex: 26, Name: "Georgian and supplement", UnicodeRanges: []string{"10A0-10FF", "2D00-2D2F"}}
	Balinese                  = UnicodeSupport{BitIndex: 27, Name: "Balinese", UnicodeRanges: []string{"1B00-1B7F"}}
	HangulJamo                = UnicodeSupport{BitIndex: 28, Name: "Hangul Jamo", UnicodeRanges: []string{"1100-11FF"}}
	LatinExtendedAdditional   = UnicodeSupport{BitIndex: 29, Name: "Latin Extended Additional", UnicodeRanges: []string{"1E00-1EFF", "2C60-2C7F", "A720-A7FF"}}
	GreekExtended             = UnicodeSupport{BitIndex: 30, Name: "Greek Extended", UnicodeRanges: []string{"1F00-1FFF"}}
	GeneralPunctuation        = UnicodeSupport{BitIndex: 31, Name: "General & Supplemental Punctuation", UnicodeRanges: []string{"2000-206F", "2E00-2E7F"}}
	SuperAndSubScripts        = UnicodeSupport{BitIndex: 32, Name: "Superscripts And Subscripts", UnicodeRanges: []string{"2070-209F"}}
	CurrencySymbols           = UnicodeSupport{BitIndex: 33, Name: "Currency Symbols", UnicodeRanges: []string{"20A0-20CF"}}
	CDMSymbols                = UnicodeSupport{BitIndex: 34, Name: "Combining Diacritical Marks For Symbols", UnicodeRanges: []string{"20D0-20FF"}}
	LetterlikeSymbols         = UnicodeSupport{BitIndex: 35, Name: "Letterlike Symbols", UnicodeRanges: []string{"2100-214F"}}
	NumberForms               = UnicodeSupport{BitIndex: 36, Name: "Number Forms", UnicodeRanges: []string{"2150-218F"}}
	Arrows                    = UnicodeSupport{BitIndex: 37, Name: "Arrows", UnicodeRanges: []string{"2190-21FF", "27F0-27FF", "2900-297F", "2B00-2BFF"}}
	MathematicalOperators     = UnicodeSupport{BitIndex: 38, Name: "Mathematical Operators", UnicodeRanges: []string{"2200-22FF", "2A00-2AFF", "27C0-27EF", "2980-29FF"}}
	MiscellaneousTechnical    = UnicodeSupport{BitIndex: 39, Name: "Miscellaneous Technical", UnicodeRanges: []string{"2300-23FF"}}

	// TODO: More unicode ranges
)

var SupportedUnicodeRanges = []UnicodeSupport{
	BasicLatin, Latin1Supplement, LatinExtendedA, LatinExtendedB, IPAPhonetic, SpacingModifierLetters, CombiningDiacriticalMarks,
	GreekCoptic, Coptic, Cyrillic, Armenian, Hebrew, Arabic, Vai, NKo,
	Devanagari, Bengali, Gurmukhi, Gujarati, Oriya, Tamil, Telugu, Kannada, Malayalam, Thai, Lao, Georgian,
	Balinese, HangulJamo, LatinExtendedAdditional, GreekExtended, GeneralPunctuation, SuperAndSubScripts,
	CurrencySymbols, CDMSymbols, LetterlikeSymbols, NumberForms, Arrows, MathematicalOperators, MiscellaneousTechnical,
}
