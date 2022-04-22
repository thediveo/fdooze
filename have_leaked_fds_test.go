package fdooze

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/thediveo/fdooze/filedesc"
)

var _ = Describe("HaveLeakedFds matcher", func() {

	It("fails for invalid actual", func() {
		m := HaveLeakedFds(nil)
		Expect(m.Match(nil)).Error().To(HaveOccurred())
		Expect(m.Match(42)).Error().To(HaveOccurred())
	})

	It("fails when filter fails", func() {
		m := HaveLeakedFds(nil, HaveField("Foo", 42))
		Expect(m.Match(Filedescriptors())).Error().To(HaveOccurred())
	})

	It("doesn't trigger a false positive", func() {
		goods := Filedescriptors()
		Expect(goods).NotTo(BeEmpty())
		oozed, err := HaveLeakedFds(goods).Match(goods)
		Expect(err).NotTo(HaveOccurred())
		Expect(oozed).To(BeFalse())
	})

	It("detects and details a leaked fd", func() {
		goods := Filedescriptors()
		Expect(goods).NotTo(BeEmpty())

		f, err := os.Open("have_leaked_fds_test.go")
		Expect(err).NotTo(HaveOccurred())
		defer f.Close()

		m := HaveLeakedFds(goods)
		oozed, err := m.Match(Filedescriptors())
		Expect(err).NotTo(HaveOccurred())
		Expect(oozed).To(BeTrue())
		Expect(m.FailureMessage(nil)).To(MatchRegexp(
			`(?m)Expected to leak \d+ file descriptors:
\s+fd \d+, flags 0x.* \(O_RDONLY,O_CLOEXEC\)
\s+path: ".*/have_leaked_fds_test.go"`))
		Expect(m.NegatedFailureMessage(nil)).To(MatchRegexp(
			`(?m)Expected not to leak \d+ file descriptors:
\s+fd \d+, flags 0x.* \(O_RDONLY,O_CLOEXEC\)
\s+path: ".*/have_leaked_fds_test.go"`))
	})

	It("sorts oozing fds", func() {
		n := func(fd int, link string) filedesc.FileDescriptor {
			fdesc, err := filedesc.NewPathFd(fd, link)
			Expect(err).WithOffset(1).NotTo(HaveOccurred())
			return fdesc
		}
		fds := []filedesc.FileDescriptor{
			n(1, "/bar1/baz"),
			n(0, "/foo0/bar"),
		}
		Expect(dumpFds(fds, 0)).To(MatchRegexp(
			`(?m)^fd 0, flags 0x.* \(.*\)\n\s+path: "/foo0/bar"\nfd 1, flags 0x.* \(.*\)\n\s+path: "/bar1/baz"$`))
	})

})
