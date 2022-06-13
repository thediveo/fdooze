//go:build linux

package fdooze

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("IgnoringFds matcher", func() {

	It("correctly handles an invalid actual value", func() {
		m := IgnoringFiledescriptors(nil)
		Expect(m.Match(nil)).Error().To(HaveOccurred())
		Expect(m.Match(42)).Error().To(HaveOccurred())
	})

	It("returns correct failure messages", func() {
		fds := Filedescriptors()
		expected := []FileDescriptor{fds[0], fds[2]}
		actual := []FileDescriptor{fds[1]}
		m := IgnoringFiledescriptors(expected)
		Expect(m.FailureMessage(actual)).To(MatchRegexp(
			`(?s)Expected
\s+<.*>: \[.*\]
to be contained in the list of expected file descriptors
\s+fd \d+, .*
\s+fd \d+, .*`))
		Expect(m.NegatedFailureMessage(actual)).To(MatchRegexp(
			`(?s)Expected
\s+<.*>: \[.*\]
not to be contained in the list of expected file descriptors
\s+fd \d+, .*
\s+fd \d+, .*`))
	})

})
