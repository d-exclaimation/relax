package f

import "strings"

// Join joins a slice of string with a separator
func Join(lines []string, sep string) string {
	return strings.Join(lines, sep)
}

// Text joins a slice of string with a new line
func Text(lines ...string) string {
	return Join(lines, "\n")
}
