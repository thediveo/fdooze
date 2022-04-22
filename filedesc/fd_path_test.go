package filedesc

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/sys/unix"
)

var _ = Describe("path fd", func() {

	It("correctly fails for invalid fd number", func() {
		Expect(NewPathFd(-1, "/foobar")).Error().To(HaveOccurred())
	})

	It("returns correct path information", func() {
		fd, err := unix.Open("fd_path_test.go", unix.O_RDONLY, 0)
		Expect(err).NotTo(HaveOccurred())
		defer unix.Close(fd)

		fdesc, err := New(fd)
		Expect(err).NotTo(HaveOccurred())

		Expect(fdesc).To(HaveField("Path()", MatchRegexp("/filedesc/fd_path_test.go$")))
		Expect(fdesc.Description(0)).To(MatchRegexp(
			"(?m)fd %d, flags .* \\(O_RDONLY\\)\n\\s+path: \".*/fd_path.test.go\"",
			fd))
	})

	It("determines equality correctly", func() {
		fd, err := unix.Open("fd_path_test.go", unix.O_RDONLY, 0)
		Expect(err).NotTo(HaveOccurred())
		defer unix.Close(fd)

		fdesc, err := New(fd)
		Expect(err).NotTo(HaveOccurred())

		Expect(fdesc.Equal(nil)).To(BeFalse())
		Expect(fdesc.Equal(fdesc)).To(BeTrue())

		fd0, err := New(0)
		Expect(err).NotTo(HaveOccurred())
		Expect(fdesc.Equal(fd0)).To(BeFalse())
	})

})
