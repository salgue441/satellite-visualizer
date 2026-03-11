package renderer

import "fmt"

// OceanShade returns the character and ANSI color for an ocean cell.
// normalZ controls brightness (closer to camera = brighter).
func OceanShade(normalZ float64) (rune, string) {
	// Map normalZ [0,1] to shade levels
	var ch rune
	var colorCode int

	switch {
	case normalZ > 0.8:
		ch = '█'
		colorCode = 27 // bright blue
	case normalZ > 0.5:
		ch = '▓'
		colorCode = 24
	case normalZ > 0.25:
		ch = '▒'
		colorCode = 21
	case normalZ > 0.1:
		ch = '░'
		colorCode = 19
	default:
		ch = '░'
		colorCode = 17 // dark blue
	}

	color := fmt.Sprintf("\033[38;5;%dm", colorCode)
	return ch, color
}
