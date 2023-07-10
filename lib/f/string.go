package f

import (
	"strings"
	"unicode/utf8"
)

// Join joins a slice of string with a separator
func Join(lines []string, sep string) string {
	return strings.Join(lines, sep)
}

// Text joins a slice of string with a new line
func Text(lines ...string) string {
	return Join(lines, "\n")
}

// TailString returns the tail of a string (everything after the first character)
func TailString(s string) string {
	_, i := utf8.DecodeRuneInString(s)
	return s[i:]
}
