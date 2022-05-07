//go:build linux

package filedesc

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/format"
)

var _ = Describe("filedesc utilities", func() {

	It("returns an indentation string given an level of indentation", func() {
		Expect(Indentation(2)).To(Equal(strings.Repeat(format.Indent, 2)))
	})

	It("hangs the indentation", func() {
		Expect(HangingIndent("", 1)).To(Equal(Indentation(1)))
		Expect(HangingIndent("foo", 1)).To(Equal(
			Indentation(1) + "foo"))
		Expect(HangingIndent("foo\nbar", 1)).To(Equal(
			Indentation(1) + "foo\n" +
				Indentation(2) + "bar"))
		Expect(HangingIndent("foo\nbar\nbaz", 1)).To(Equal(
			Indentation(1) + "foo\n" +
				Indentation(2) + "bar\n" +
				Indentation(2) + "baz"))
	})

})
