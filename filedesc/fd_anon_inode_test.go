//go:build linux

package filedesc

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/sys/unix"
)

var _ = Describe("anonymous inode fd", func() {

	It("correctly fails for invalid fd number", func() {
		Expect(NewAnonInodeFd(-1, "anon_inode:[foobar]")).Error().To(HaveOccurred())
	})

	It("returns the correct anonymous inode file type and description", func() {
		fd, err := unix.Eventfd(42, unix.EFD_CLOEXEC)
		Expect(err).NotTo(HaveOccurred())
		defer unix.Close(fd)

		fdesc, err := New(fd)
		Expect(err).NotTo(HaveOccurred())
		anonfd := fdesc.(*AnonInodeFd)
		Expect(anonfd.FileType()).To(Equal("eventfd"))
		Expect(anonfd.Description(0)).To(MatchRegexp(
			`fd \d+, flags 0x.* \(O_RDWR,O_CLOEXEC\)\n\s+anonymous inode file type: "eventfd"`))
	})

	It("determines equality correctly", func() {
		fd, err := unix.Eventfd(42, unix.EFD_CLOEXEC)
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
