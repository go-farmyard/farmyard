package termcolor

import "strconv"

const EscapeChar = '\033'

// ColorAttribute defines a single SGR Code
type ColorAttribute int

// Base text attributes
const (
	Reset ColorAttribute = iota
	Bold
	Faint
	Italic
	Underline
	BlinkSlow
	BlinkRapid
	ReverseVideo
	Concealed
	CrossedOut
)

// Foreground text colors
const (
	FgBlack ColorAttribute = iota + 30
	FgRed
	FgGreen
	FgYellow
	FgBlue
	FgMagenta
	FgCyan
	FgWhite
)

// Foreground Hi-Intensity text colors
const (
	FgHiBlack ColorAttribute = iota + 90
	FgHiRed
	FgHiGreen
	FgHiYellow
	FgHiBlue
	FgHiMagenta
	FgHiCyan
	FgHiWhite
)

// Background text colors
const (
	BgBlack ColorAttribute = iota + 40
	BgRed
	BgGreen
	BgYellow
	BgBlue
	BgMagenta
	BgCyan
	BgWhite
)

// Background Hi-Intensity text colors
const (
	BgHiBlack ColorAttribute = iota + 100
	BgHiRed
	BgHiGreen
	BgHiYellow
	BgHiBlue
	BgHiMagenta
	BgHiCyan
	BgHiWhite
)

// ColorBytes converts a list of ColorAttributes to a byte array
func ColorBytes(attrs ...ColorAttribute) []byte {
	bytes := make([]byte, 0, 10)
	bytes = append(bytes, EscapeChar, '[')
	if len(attrs) > 0 {
		bytes = append(bytes, strconv.Itoa(int(attrs[0]))...)
		for _, a := range attrs[1:] {
			bytes = append(bytes, ';')
			bytes = append(bytes, strconv.Itoa(int(a))...)
		}
	} else {
		bytes = append(bytes, strconv.Itoa(int(Bold))...)
	}
	bytes = append(bytes, 'm')
	return bytes
}
