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
	BasicLatin                         = UnicodeSupport{BitIndex: 0, Name: "Basic Latin", UnicodeRanges: []string{"0000-007F"}}
	Latin1Supplement                   = UnicodeSupport{BitIndex: 1, Name: "Latin-1 Supplement", UnicodeRanges: []string{"0080-00FF"}}
	LatinExtendedA                     = UnicodeSupport{BitIndex: 2, Name: "Latin Extended-A", UnicodeRanges: []string{"0100-017F"}}
	LatinExtendedB                     = UnicodeSupport{BitIndex: 3, Name: "Latin Extended-B", UnicodeRanges: []string{"0180-024F"}}
	IPAPhonetic                        = UnicodeSupport{BitIndex: 4, Name: "IPA & Phonetic Extensions/Supplements", UnicodeRanges: []string{"0250-02AF", "1D00-1D7F", "1D80-1DBF"}}
	SpacingModifierLetters             = UnicodeSupport{BitIndex: 5, Name: "Spacing Modifier Letters", UnicodeRanges: []string{"02B0-02FF", "A700-A71F"}}
	CombiningDiacriticalMarks          = UnicodeSupport{BitIndex: 6, Name: "Combining Diacritical Marks & Supplements", UnicodeRanges: []string{"0300-036F", "1DC0-1DFF"}}
	GreekCoptic                        = UnicodeSupport{BitIndex: 7, Name: "Greek and Coptic", UnicodeRanges: []string{"0370-03FF"}}
	Coptic                             = UnicodeSupport{BitIndex: 8, Name: "Coptic", UnicodeRanges: []string{"2C80-2CFF"}}
	Cyrillic                           = UnicodeSupport{BitIndex: 9, Name: "Cyrillic", UnicodeRanges: []string{"0400-04FF", "0500-052F", "2DE0-2DFF", "A640-A69F"}}
	Armenian                           = UnicodeSupport{BitIndex: 10, Name: "Armenian", UnicodeRanges: []string{"0530-058F"}}
	Hebrew                             = UnicodeSupport{BitIndex: 11, Name: "Hebrew", UnicodeRanges: []string{"0590-05FF"}}
	Vai                                = UnicodeSupport{BitIndex: 12, Name: "Vai", UnicodeRanges: []string{"A500-A63F"}}
	Arabic                             = UnicodeSupport{BitIndex: 13, Name: "Arabic", UnicodeRanges: []string{"0600-06FF", "0750-077F"}}
	NKo                                = UnicodeSupport{BitIndex: 14, Name: "NKo", UnicodeRanges: []string{"07C0-07FF"}}
	Devanagari                         = UnicodeSupport{BitIndex: 15, Name: "Devanagari", UnicodeRanges: []string{"0900-097F"}}
	Bengali                            = UnicodeSupport{BitIndex: 16, Name: "Bengali", UnicodeRanges: []string{"0980-09FF"}}
	Gurmukhi                           = UnicodeSupport{BitIndex: 17, Name: "Gurmukhi", UnicodeRanges: []string{"0A00-0A7F"}}
	Gujarati                           = UnicodeSupport{BitIndex: 18, Name: "Gujarati", UnicodeRanges: []string{"0A80-0AFF"}}
	Oriya                              = UnicodeSupport{BitIndex: 19, Name: "Oriya", UnicodeRanges: []string{"0B00-0B7F"}}
	Tamil                              = UnicodeSupport{BitIndex: 20, Name: "Tamil", UnicodeRanges: []string{"0B80-0BFF"}}
	Telugu                             = UnicodeSupport{BitIndex: 21, Name: "Telugu", UnicodeRanges: []string{"0C00-0C7F"}}
	Kannada                            = UnicodeSupport{BitIndex: 22, Name: "Kannada", UnicodeRanges: []string{"0C80-0CFF"}}
	Malayalam                          = UnicodeSupport{BitIndex: 23, Name: "Malayalam", UnicodeRanges: []string{"0D00-0D7F"}}
	Thai                               = UnicodeSupport{BitIndex: 24, Name: "Thai", UnicodeRanges: []string{"0E00-0E7F"}}
	Lao                                = UnicodeSupport{BitIndex: 25, Name: "Lao", UnicodeRanges: []string{"0E80-0EFF"}}
	Georgian                           = UnicodeSupport{BitIndex: 26, Name: "Georgian and supplement", UnicodeRanges: []string{"10A0-10FF", "2D00-2D2F"}}
	Balinese                           = UnicodeSupport{BitIndex: 27, Name: "Balinese", UnicodeRanges: []string{"1B00-1B7F"}}
	HangulJamo                         = UnicodeSupport{BitIndex: 28, Name: "Hangul Jamo", UnicodeRanges: []string{"1100-11FF"}}
	LatinExtendedAdditional            = UnicodeSupport{BitIndex: 29, Name: "Latin Extended Additional", UnicodeRanges: []string{"1E00-1EFF", "2C60-2C7F", "A720-A7FF"}}
	GreekExtended                      = UnicodeSupport{BitIndex: 30, Name: "Greek Extended", UnicodeRanges: []string{"1F00-1FFF"}}
	GeneralPunctuation                 = UnicodeSupport{BitIndex: 31, Name: "General & Supplemental Punctuation", UnicodeRanges: []string{"2000-206F", "2E00-2E7F"}}
	SuperAndSubScripts                 = UnicodeSupport{BitIndex: 32, Name: "Superscripts And Subscripts", UnicodeRanges: []string{"2070-209F"}}
	CurrencySymbols                    = UnicodeSupport{BitIndex: 33, Name: "Currency Symbols", UnicodeRanges: []string{"20A0-20CF"}}
	CDMSymbols                         = UnicodeSupport{BitIndex: 34, Name: "Combining Diacritical Marks For Symbols", UnicodeRanges: []string{"20D0-20FF"}}
	LetterlikeSymbols                  = UnicodeSupport{BitIndex: 35, Name: "Letterlike Symbols", UnicodeRanges: []string{"2100-214F"}}
	NumberForms                        = UnicodeSupport{BitIndex: 36, Name: "Number Forms", UnicodeRanges: []string{"2150-218F"}}
	Arrows                             = UnicodeSupport{BitIndex: 37, Name: "Arrows", UnicodeRanges: []string{"2190-21FF", "27F0-27FF", "2900-297F", "2B00-2BFF"}}
	MathematicalOperators              = UnicodeSupport{BitIndex: 38, Name: "Mathematical Operators", UnicodeRanges: []string{"2200-22FF", "2A00-2AFF", "27C0-27EF", "2980-29FF"}}
	MiscellaneousTechnical             = UnicodeSupport{BitIndex: 39, Name: "Miscellaneous Technical", UnicodeRanges: []string{"2300-23FF"}}
	ControlPictures                    = UnicodeSupport{BitIndex: 40, Name: "Control Pictures", UnicodeRanges: []string{"2400-243F"}}
	OpticalCharacterRecognition        = UnicodeSupport{BitIndex: 41, Name: "Optical Character Recognition", UnicodeRanges: []string{"2440-245F"}}
	EnclosedAlphanumerics              = UnicodeSupport{BitIndex: 42, Name: "Enclosed Alphanumerics", UnicodeRanges: []string{"2460-24FF"}}
	BoxDrawing                         = UnicodeSupport{BitIndex: 43, Name: "Box Drawing", UnicodeRanges: []string{"2500-257F"}}
	BlockElements                      = UnicodeSupport{BitIndex: 44, Name: "Block Elements", UnicodeRanges: []string{"2580-259F"}}
	GeometricShapes                    = UnicodeSupport{BitIndex: 45, Name: "Geometric Shapes", UnicodeRanges: []string{"25A0-25FF"}}
	MiscellaneousSymbols               = UnicodeSupport{BitIndex: 46, Name: "Miscellaneous Symbols", UnicodeRanges: []string{"2600-26FF"}}
	Dingbats                           = UnicodeSupport{BitIndex: 47, Name: "Dingbats", UnicodeRanges: []string{"2700-27BF"}}
	CJKSymbolsPunctuation              = UnicodeSupport{BitIndex: 48, Name: "CJK Symbols And Punctuation", UnicodeRanges: []string{"3000-303F"}}
	Hiragana                           = UnicodeSupport{BitIndex: 49, Name: "Hiragana", UnicodeRanges: []string{"3040-309F"}}
	Katakana                           = UnicodeSupport{BitIndex: 50, Name: "Katakana & Phonetic Extensions", UnicodeRanges: []string{"30A0-30FF", "31F0-31FF"}}
	Bopomofo                           = UnicodeSupport{BitIndex: 51, Name: "Bopomofo & Extended", UnicodeRanges: []string{"3100-312F", "31A0-31BF"}}
	HangulCompatibilityJamo            = UnicodeSupport{BitIndex: 52, Name: "Hangul Compatibility Jamo", UnicodeRanges: []string{"3130-318F"}}
	PhagsPa                            = UnicodeSupport{BitIndex: 53, Name: "Phags-pa", UnicodeRanges: []string{"A840-A87F"}}
	EnclosedCJKLettersMonths           = UnicodeSupport{BitIndex: 54, Name: "Enclosed CJK Letters And Months", UnicodeRanges: []string{"3200-32FF"}}
	CJKCompatibility                   = UnicodeSupport{BitIndex: 55, Name: "CJK Compatibility", UnicodeRanges: []string{"3300-33FF"}}
	HangulSyllables                    = UnicodeSupport{BitIndex: 56, Name: "Hangul Syllables", UnicodeRanges: []string{"AC00-D7AF"}}
	NonPlane0                          = UnicodeSupport{BitIndex: 57, Name: "Non-Plane 0", UnicodeRanges: []string{"10000-10FFFF"}}
	Phoenician                         = UnicodeSupport{BitIndex: 58, Name: "Phoenician", UnicodeRanges: []string{"10900-1091F"}}
	CJKIdeographsRadicals              = UnicodeSupport{BitIndex: 59, Name: "CJK Ideographs & Radicals", UnicodeRanges: []string{"4E00-9FFF", "2E80-2EFF", "2F00-2FDF", "2FF0-2FFF", "3400-4DBF", "20000-2A6DF", "3190-319F"}}
	PrivateUseAreaPlane0               = UnicodeSupport{BitIndex: 60, Name: "Private Use Area  Plane 0", UnicodeRanges: []string{"E000-F8FF"}}
	CJKStrokes                         = UnicodeSupport{BitIndex: 61, Name: "CJK Strokes", UnicodeRanges: []string{"31C0-31EF", "F900-FAFF", "2F800-2FA1F"}}
	AlphabeticPresentationForms        = UnicodeSupport{BitIndex: 62, Name: "Alphabetic Presentation Forms", UnicodeRanges: []string{"FB00-FB4F"}}
	ArabicPresentationFormsA           = UnicodeSupport{BitIndex: 63, Name: "Arabic Presentation Forms-A", UnicodeRanges: []string{"FB50-FDFF"}}
	CombiningHalfMarks                 = UnicodeSupport{BitIndex: 64, Name: "Combining Half Marks", UnicodeRanges: []string{"FE20-FE2F"}}
	VerticalForms                      = UnicodeSupport{BitIndex: 65, Name: "Vertical Forms", UnicodeRanges: []string{"FE10-FE1F", "FE30-FE4F"}}
	SmallFormVariants                  = UnicodeSupport{BitIndex: 66, Name: "Small Form Variants", UnicodeRanges: []string{"FE50-FE6F"}}
	ArabicPresentationFormsB           = UnicodeSupport{BitIndex: 67, Name: "Arabic Presentation Forms-B", UnicodeRanges: []string{"FE70-FEFF"}}
	HalfwidthFullwidthForms            = UnicodeSupport{BitIndex: 68, Name: "Halfwidth And Fullwidth Forms", UnicodeRanges: []string{"FF00-FFEF"}}
	Specials                           = UnicodeSupport{BitIndex: 69, Name: "Specials", UnicodeRanges: []string{"FFF0-FFFF"}}
	Tibetan                            = UnicodeSupport{BitIndex: 70, Name: "Tibetan", UnicodeRanges: []string{"0F00-0FFF"}}
	Syriac                             = UnicodeSupport{BitIndex: 71, Name: "Syriac", UnicodeRanges: []string{"0700-074F"}}
	Thaana                             = UnicodeSupport{BitIndex: 72, Name: "Thaana", UnicodeRanges: []string{"0780-07BF"}}
	Sinhala                            = UnicodeSupport{BitIndex: 73, Name: "Sinhala", UnicodeRanges: []string{"0D80-0DFF"}}
	Myanmar                            = UnicodeSupport{BitIndex: 74, Name: "Myanmar", UnicodeRanges: []string{"1000-109F"}}
	Ethiopic                           = UnicodeSupport{BitIndex: 75, Name: "Ethiopic", UnicodeRanges: []string{"1200-137F", "1380-139F", "2D80-2DDF"}}
	Cherokee                           = UnicodeSupport{BitIndex: 76, Name: "Cherokee", UnicodeRanges: []string{"13A0-13FF"}}
	UnifiedCanadianAboriginalSyllabics = UnicodeSupport{BitIndex: 77, Name: "Unified Canadian Aboriginal Syllabics", UnicodeRanges: []string{"1400-167F"}}
	Runic                              = UnicodeSupport{BitIndex: 79, Name: "Runic", UnicodeRanges: []string{"16A0-16FF"}}
	Khmer                              = UnicodeSupport{BitIndex: 80, Name: "Khmer", UnicodeRanges: []string{"1780-17FF", "19E0-19FF"}}
	Mongolian                          = UnicodeSupport{BitIndex: 81, Name: "Mongolian", UnicodeRanges: []string{"1800-18AF"}}
	BraillePatterns                    = UnicodeSupport{BitIndex: 82, Name: "Braille Patterns", UnicodeRanges: []string{"2800-28FF"}}
	YiSyllablesRadicals                = UnicodeSupport{BitIndex: 83, Name: "Yi Syllables & Radicals", UnicodeRanges: []string{"A000-A48F", "A490-A4CF"}}
	TagalogRelated                     = UnicodeSupport{BitIndex: 84, Name: "Tagalog And Related", UnicodeRanges: []string{"1700-171F", "1720-173F", "1740-175F", "1760-177F"}}
	OldItalic                          = UnicodeSupport{BitIndex: 85, Name: "Old Italic", UnicodeRanges: []string{"10300-1032F"}}
	Gothic                             = UnicodeSupport{BitIndex: 86, Name: "Gothic", UnicodeRanges: []string{"10330-1034F"}}
	Deseret                            = UnicodeSupport{BitIndex: 87, Name: "Deseret", UnicodeRanges: []string{"10400-1044F"}}
	MusicalSymbols                     = UnicodeSupport{BitIndex: 88, Name: "Sinhala", UnicodeRanges: []string{"1D000-1D0FF", "1D100-1D1FF", "1D200-1D24F"}}
	MathematicalAlphanumericSymbols    = UnicodeSupport{BitIndex: 89, Name: "Mathematical Alphanumeric Symbols", UnicodeRanges: []string{"1D400-1D7FF"}}
	PrivateUsePlane1516                = UnicodeSupport{BitIndex: 90, Name: "Private Use (plane 15 & 16)", UnicodeRanges: []string{"F0000-FFFFD", "100000-10FFFD"}}
	VariationSelectors                 = UnicodeSupport{BitIndex: 91, Name: "Variation Selectors", UnicodeRanges: []string{"FE00-FE0F", "E0100-E01EF"}}
	Tags                               = UnicodeSupport{BitIndex: 92, Name: "Tags", UnicodeRanges: []string{"E0000-E007F"}}
	Limbu                              = UnicodeSupport{BitIndex: 93, Name: "Limbu", UnicodeRanges: []string{"1900-194F"}}
	TaiLe                              = UnicodeSupport{BitIndex: 94, Name: "Tai Le", UnicodeRanges: []string{"1950-197F"}}
	NewTaiLue                          = UnicodeSupport{BitIndex: 95, Name: "New Tai Lue", UnicodeRanges: []string{"1980-19DF"}}
	Buginese                           = UnicodeSupport{BitIndex: 96, Name: "Buginese", UnicodeRanges: []string{"1A00-1A1F"}}
	Glagolitic                         = UnicodeSupport{BitIndex: 97, Name: "Glagolitic", UnicodeRanges: []string{"2C00-2C5F"}}
	Tifinagh                           = UnicodeSupport{BitIndex: 98, Name: "Tifinagh", UnicodeRanges: []string{"2D30-2D7F"}}
	YijingHexagramSymbols              = UnicodeSupport{BitIndex: 99, Name: "Yijing Hexagram Symbols", UnicodeRanges: []string{"4DC0-4DFF"}}
	SylotiNagri                        = UnicodeSupport{BitIndex: 100, Name: "Syloti Nagri", UnicodeRanges: []string{"A800-A82F"}}
	LinearBAegean                      = UnicodeSupport{BitIndex: 101, Name: "Linear B Syllabary & Aegean", UnicodeRanges: []string{"10000-1007F", "10080-100FF", "10100-1013F"}}
	AncientGreekNumbers                = UnicodeSupport{BitIndex: 102, Name: "Ancient Greek Numbers", UnicodeRanges: []string{"10140-1018F"}}
	Ugaritic                           = UnicodeSupport{BitIndex: 103, Name: "Ugaritic", UnicodeRanges: []string{"10380-1039F"}}
	OldPersian                         = UnicodeSupport{BitIndex: 104, Name: "Old Persian", UnicodeRanges: []string{"103A0-103DF"}}
	Shavian                            = UnicodeSupport{BitIndex: 105, Name: "Shavian", UnicodeRanges: []string{"10450-1047F"}}
	Osmanya                            = UnicodeSupport{BitIndex: 106, Name: "Osmanya", UnicodeRanges: []string{"10480-104AF"}}
	CypriotSyllabary                   = UnicodeSupport{BitIndex: 107, Name: "Cypriot Syllabary", UnicodeRanges: []string{"10800-1083F"}}
	Kharoshthi                         = UnicodeSupport{BitIndex: 108, Name: "Kharoshthi", UnicodeRanges: []string{"10A00-10A5F"}}
	TaiXuanJingSymbols                 = UnicodeSupport{BitIndex: 109, Name: "Tai Xuan Jing Symbols", UnicodeRanges: []string{"1D300-1D35F"}}
	Cuneiform                          = UnicodeSupport{BitIndex: 110, Name: "Cuneiform", UnicodeRanges: []string{"12000-123FF", "12400-1247F"}}
	CountingRodNumerals                = UnicodeSupport{BitIndex: 111, Name: "Counting Rod Numerals", UnicodeRanges: []string{"1D360-1D37F"}}
	Sundanese                          = UnicodeSupport{BitIndex: 112, Name: "Sundanese", UnicodeRanges: []string{"1B80-1BBF"}}
	Lepcha                             = UnicodeSupport{BitIndex: 113, Name: "Lepcha", UnicodeRanges: []string{"1C00-1C4F"}}
	OlChiki                            = UnicodeSupport{BitIndex: 114, Name: "Ol Chiki", UnicodeRanges: []string{"1C50-1C7F"}}
	Saurashtra                         = UnicodeSupport{BitIndex: 115, Name: "Saurashtra", UnicodeRanges: []string{"A880-A8DF"}}
	KayahLi                            = UnicodeSupport{BitIndex: 116, Name: "Kayah Li", UnicodeRanges: []string{"A900-A92F"}}
	Rejang                             = UnicodeSupport{BitIndex: 117, Name: "Rejang", UnicodeRanges: []string{"A930-A95F"}}
	Cham                               = UnicodeSupport{BitIndex: 118, Name: "Cham", UnicodeRanges: []string{"AA00-AA5F"}}
	AncientSymbols                     = UnicodeSupport{BitIndex: 119, Name: "Ancient Symbols", UnicodeRanges: []string{"10190-101CF"}}
	PhaistosDisc                       = UnicodeSupport{BitIndex: 120, Name: "Phaistos Disc", UnicodeRanges: []string{"101D0-101FF"}}
	CarianLycianLydian                 = UnicodeSupport{BitIndex: 121, Name: "Carian Lycian Lydian", UnicodeRanges: []string{"102A0-102DF", "10280-1029F", "10920-1093F"}}
	DominoMahjong                      = UnicodeSupport{BitIndex: 122, Name: "Domino & Mahjong Tiles", UnicodeRanges: []string{"1F030-1F09F", "1F000-1F02F"}}
)

var SupportedUnicodeRanges = []UnicodeSupport{
	BasicLatin, Latin1Supplement, LatinExtendedA, LatinExtendedB, IPAPhonetic, SpacingModifierLetters, CombiningDiacriticalMarks,
	GreekCoptic, Coptic, Cyrillic, Armenian, Hebrew, Arabic, Vai, NKo,
	Devanagari, Bengali, Gurmukhi, Gujarati, Oriya, Tamil, Telugu, Kannada, Malayalam, Thai, Lao, Georgian,
	Balinese, HangulJamo, LatinExtendedAdditional, GreekExtended, GeneralPunctuation, SuperAndSubScripts,
	CurrencySymbols, CDMSymbols, LetterlikeSymbols, NumberForms, Arrows, MathematicalOperators, MiscellaneousTechnical,
	ControlPictures, OpticalCharacterRecognition, EnclosedAlphanumerics, BoxDrawing, BlockElements, GeometricShapes, MiscellaneousSymbols,
	Dingbats, CJKSymbolsPunctuation, Hiragana, Katakana, Bopomofo, HangulCompatibilityJamo, PhagsPa, EnclosedCJKLettersMonths, CJKCompatibility,
	HangulSyllables, NonPlane0, Phoenician, CJKIdeographsRadicals, PrivateUseAreaPlane0, CJKStrokes, AlphabeticPresentationForms, ArabicPresentationFormsA,
	CombiningHalfMarks, SmallFormVariants, VerticalForms, ArabicPresentationFormsB, HalfwidthFullwidthForms, Specials, Tibetan, Syriac, Thaana,
	Sinhala, Myanmar, Ethiopic, Cherokee, UnifiedCanadianAboriginalSyllabics, Runic, Khmer, Mongolian, YiSyllablesRadicals, BraillePatterns,
	TagalogRelated, OldItalic, Gothic, MusicalSymbols, Deseret, MathematicalAlphanumericSymbols, PrivateUsePlane1516, VariationSelectors, Tags,
	Limbu, TaiLe, NewTaiLue, Glagolitic, Tifinagh, Buginese, YijingHexagramSymbols, SylotiNagri, LinearBAegean, AncientGreekNumbers, Ugaritic,
	OldPersian, Shavian, Osmanya, CypriotSyllabary, Kharoshthi, TaiXuanJingSymbols, Cuneiform, CountingRodNumerals, Sundanese,
	Lepcha, OlChiki, Saurashtra, KayahLi, Rejang, Cham, AncientSymbols, PhaistosDisc, CarianLycianLydian, DominoMahjong,
}
