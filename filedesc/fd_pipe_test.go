//go:build linux

package filedesc

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/sys/unix"
)

var _ = Describe("pipe fd", func() {

	It("correctly fails for invalid fd number or pipe inode number", func() {
		Expect(NewPipeFd(-1, "foobar")).Error().To(HaveOccurred())
		Expect(NewPipeFd(-1, "pipe:[123456]")).Error().To(HaveOccurred())
	})

	When("given pipe ends", Ordered, func() {

		var pipefds [2]int

		BeforeAll(func() {
			By("creating a pair of unnamed pipe ends")
			Expect(unix.Pipe(pipefds[:])).Error().NotTo(HaveOccurred())
			Expect(pipefds).To(HaveEach(Not(BeZero())))
			DeferCleanup(func() {
				unix.Close(pipefds[0])
				unix.Close(pipefds[1])
			})
		})

		It("returns correct pipe details", func() {
			rfdesc, err := New(pipefds[0])
			Expect(err).NotTo(HaveOccurred())
			Expect(rfdesc.(*PipeFd)).NotTo(BeNil())
			Expect(rfdesc.Description(0)).To(MatchRegexp(
				"(?m)fd %d, flags 0x0 \\(O_RDONLY\\)\n\\s+pipe inode number: \\d+",
				pipefds[0]))

			wfdesc, err := New(pipefds[1])
			Expect(err).NotTo(HaveOccurred())
			Expect(wfdesc.(*PipeFd)).NotTo(BeNil())
			Expect(wfdesc.Description(0)).To(MatchRegexp(
				"(?m)fd %d, flags 0x1 \\(O_WRONLY\\)\n\\s+pipe inode number: \\d+",
				pipefds[1]))

			Expect(rfdesc.(*PipeFd).Ino()).To(Equal(wfdesc.(*PipeFd).Ino()))
		})

		It("determines equality correctly", func() {
			rfdesc, err := New(pipefds[0])
			Expect(err).NotTo(HaveOccurred())
			Expect(rfdesc.(*PipeFd)).NotTo(BeNil())
			wfdesc, err := New(pipefds[1])
			Expect(err).NotTo(HaveOccurred())
			Expect(wfdesc.(*PipeFd)).NotTo(BeNil())

			Expect(rfdesc.Equal(nil)).To(BeFalse())
			Expect(rfdesc.Equal(wfdesc)).To(BeFalse())
			Expect(rfdesc.Equal(rfdesc)).To(BeTrue())
		})

	})

})
