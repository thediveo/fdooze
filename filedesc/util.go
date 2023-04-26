// Copyright 2022 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy
// of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

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
