//go:build linux

package filedesc

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strings"
	"testing/iotest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/sys/unix"
)

var _ = Describe("file descriptors", func() {

	When("dealing with a single file descriptor", func() {
		It("returns error when reading errors", func() {
			Expect(fdFromReader(42, iotest.ErrReader(errors.New("foobar")))).Error().To(
				MatchError("foobar"))
		})

		It("returns error when reading incomplete information", func() {
			r := strings.NewReader("pos:\t0\nflags:\t042\n")
			Expect(fdFromReader(42, r)).Error().To(
				MatchError(ContainSubstring("incomplete fdinfo data")))
		})

		It("returns error when reading out-of-range information", func() {
			r := strings.NewReader(fmt.Sprintf(
				"pos:\t0\nflags:\t%o\nmnt_id:\t123\n", uint64(math.MaxInt)+1))
			Expect(fdFromReader(42, r)).Error().To(
				MatchError(ContainSubstring("flags outside range:")))
			r = strings.NewReader("pos:\t0\nflags:\t042\nmnt_id:\t-1\n")
			Expect(fdFromReader(42, r)).Error().To(
				MatchError(ContainSubstring("mnt_id outside range:")))
		})

		It("reads and returns common fd information", func() {
			r := strings.NewReader("pos:\t0\nflags:\t042\nmnt_id:\t123\n")
			fdesc, err := fdFromReader(42, r)
			Expect(err).NotTo(HaveOccurred())
			Expect(fdesc.Fd()).To(Equal(42))
			Expect(fdesc.Flags()).To(Equal(Flags(042)))
			Expect(fdesc.MountId()).To(Equal(123))
		})

		It("returns a correct description", func() {
			fdesc := filedesc{
				fd:    42,
				flags: Flags(os.O_APPEND),
				mntId: 123,
			}
			Expect(fdesc.Description(0)).To(Equal(
				fmt.Sprintf("fd 42, flags 0x%x (O_RDONLY,O_APPEND)", os.O_APPEND)))
		})

		It("doesn't fail to read information about fd 0", func() {
			fdesc, err := newFiledesc(0)
			Expect(err).NotTo(HaveOccurred())
			Expect(fdesc.fd).To(Equal(0))
			Expect(fdesc.mntId).NotTo(BeZero())
		})

		It("fails correctly to read invalid fd information", func() {
			r := strings.NewReader("pos:\t0\nflags:\t099\nmnt_id:\t123\n")
			Expect(fdFromReader(0, r)).Error().To(MatchError(MatchRegexp("invalid syntax")))

			r = strings.NewReader("pos:\t0\nflags:\t042\nmnt_id:\tabc\n")
			Expect(fdFromReader(0, r)).Error().To(MatchError(MatchRegexp("invalid syntax")))
		})

		It("fails correctly to read from fd -1", func() {
			Expect(newFiledesc(-1)).Error().To(MatchError(MatchRegexp("open.*/proc/self/fdinfo/-1")))
		})
	})

	When("discovering fds", func() {
		It("returns error or nothing for missing or invalid procfs", func() {
			Expect(filedescriptors("./test/missing-proc/fd")).Error().To(HaveOccurred())
			Expect(filedescriptors("./test/not-an-fd-directory")).Error().To(HaveOccurred())
			Expect(filedescriptors("./test/fake-proc/fd")).To(BeEmpty())
		})

		It("finds this process's file descriptors", func() {
			fd, err := unix.Socket(unix.AF_UNIX, unix.SOCK_STREAM, 0)
			Expect(err).NotTo(HaveOccurred())
			defer unix.Close(fd)

			fdescs := Filedescriptors()
			Expect(fdescs).NotTo(BeEmpty())
			Expect(fdescs).To(ContainElement(HaveField("Fd()", fd)))
		})
	})

})
