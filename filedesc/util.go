//go:build linux

package filedesc

import (
	"strings"

	"github.com/onsi/gomega/format"
)

// Indentation returns an indentation string for the specified indentation level
// (and 0 meaning no indentation). The indentation parameter terminology has
// been taken over from Gomega's format package, where it refers to the level of
// indentation. The width of an indentation level is Gomega's [format.Indent]
// variable, which defaults to four spaces.
func Indentation(indentation uint) string {
	return strings.Repeat(format.Indent, int(indentation)) // still wondering about Repeat("D'OH", -1)...
}

// HangingIndent indents the first line in s the specified indentation level,
// and then all following lines one level deeper. It should not be confused with
// Gomega's [format.IndentString] which indents all lines in a string the same
// level.
func HangingIndent(s string, indentation uint) string {
	firstIndent := Indentation(indentation)
	indent := firstIndent + Indentation(1)
	lines := strings.Split(s, "\n")
	var out strings.Builder
	for idx, line := range lines {
		if idx == 0 {
			out.WriteString(firstIndent)
			out.WriteString(line)
			continue
		}
		out.WriteRune('\n')
		out.WriteString(indent)
		out.WriteString(line)
	}
	return out.String()
}
