//go:build linux

package filedesc

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"golang.org/x/sys/unix"
)

var _ = Describe("socket utilities", func() {

	Context("socket parameters", func() {

		It("renders textual representation of socket (communication) domain", func() {
			Expect(SocketDomain(unix.AF_INET6).String()).To(Equal("AF_INET6"))
			Expect(SocketDomain(-1).String()).To(Equal("domain -1"))
		})

		It("renders textual representation of socket type (communication semantics)", func() {
			Expect(SocketType(unix.SOCK_STREAM).String()).To(Equal("SOCK_STREAM"))
			Expect(SocketType(-1).String()).To(Equal("type -1"))
		})

		It("renders textual representation of socket protocol in a domain", func() {
			Expect(SocketProtocol(unix.IPPROTO_TCP).String(unix.AF_INET)).To(
				Equal("IPPROTO_TCP"))
			Expect(SocketProtocol(unix.NETLINK_ROUTE).String(unix.AF_NETLINK)).To(
				Equal("NETLINK_ROUTE"))
			Expect(SocketProtocol(unix.IPPROTO_TCP).String(0)).To(
				Equal(fmt.Sprintf("protocol %d", unix.IPPROTO_TCP)))
		})

	})

	It("converts l2 addresses into text", func() {
		Expect(hexString(nil, ':')).To(Equal(""))
		Expect(hexString([]byte{0x1a, 0x2b, 0x3c, 0x4d, 0x5e, 0x6f}, ':')).To(Equal("1A:2B:3C:4D:5E:6F"))
	})

})
