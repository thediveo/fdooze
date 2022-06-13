//go:build linux

package fdooze

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/thediveo/fdooze/filedesc"
)

var _ = Describe("util", func() {

	It("checks an actual to be a slice of file descriptors", func() {
		Expect(toFds(nil, "Foo")).Error().To(MatchError(MatchRegexp(
			`Foo matcher expects an array or slice of file descriptors.  Got:\n\s+<nil>: nil`)))
		Expect(toFds([]int{42}, "Foo")).Error().To(MatchError(MatchRegexp(
			`Foo matcher expects an array or slice of file descriptors.  Got:\n\s+<\[\]int | len:1, cap:1>: \[42\]`)))
	})

	It("sorts oozing fds", func() {
		n := func(fd int, link string) FileDescriptor {
			fdesc, err := filedesc.NewPathFd(fd, link)
			Expect(err).WithOffset(1).NotTo(HaveOccurred())
			return fdesc
		}
		fds := []FileDescriptor{
			n(1, "/bar1/baz"),
			n(0, "/foo0/bar"),
		}
		Expect(dumpFds(fds, 0)).To(MatchRegexp(
			`(?m)^fd 0, flags 0x.* \(.*\)\n\s+path: "/foo0/bar"\nfd 1, flags 0x.* \(.*\)\n\s+path: "/bar1/baz"$`))
	})

})
