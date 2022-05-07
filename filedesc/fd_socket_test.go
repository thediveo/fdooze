//go:build linux

package filedesc

import (
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/sys/unix"
)

var _ = Describe("socket descriptors", func() {

	It("correctly handles invalid fd number or socket inode number", func() {
		Expect(NewSocketFd(0, "socket:[abc]")).Error().To(HaveOccurred())
		Expect(NewSocketFd(-1, "socket:[123456]")).Error().To(HaveOccurred())
		Expect(NewSocketFd(0, "socket:[123456]")).Error().To(HaveOccurred())
	})

	When("mocking socket syscalls", Serial, func() {

		Context("correctly handles failing socket syscalls", Ordered, func() {

			var fd int

			BeforeAll(func() {
				var err error
				fd, err = unix.Socket(unix.AF_UNIX, unix.SOCK_STREAM, 0)
				Expect(err).NotTo(HaveOccurred())
				DeferCleanup(func() {
					unix.Close(fd)
				})
			})

			DescribeTable("failing to get socket option", Ordered,
				func(blockopt int) {
					oldgetsockoptInt := getsockoptInt
					defer func() { getsockoptInt = oldgetsockoptInt }()

					getsockoptInt = func(fd, level, opt int) (int, error) {
						if level != unix.SOL_SOCKET || opt != blockopt {
							return oldgetsockoptInt(fd, level, opt)
						}
						return 0, fmt.Errorf("failing option %d", opt)
					}

					Expect(NewSocketFd(fd, "socket:[123456]")).Error().To(
						MatchError(fmt.Errorf("failing option %d", blockopt)))
				},
				Entry("failing SO_DOMAIN", unix.SO_DOMAIN),
				Entry("failing SO_TYPE", unix.SO_TYPE),
				Entry("failing SO_PROTOCOL", unix.SO_PROTOCOL),
				Entry("failing SO_ACCEPTCONN", unix.SO_ACCEPTCONN),
			)

			It("correctly handles failing Getsockname", func() {
				oldgetsockname := getsockname
				defer func() { getsockname = oldgetsockname }()

				getsockname = func(fd int) (unix.Sockaddr, error) {
					return nil, errors.New("failing getsockname")
				}

				Expect(NewSocketFd(fd, "socket:[123456]")).Error().To(
					MatchError("failing getsockname"))
			})

			It("correctly handles failing Getpeername", func() {
				oldgetpeername := getpeername
				defer func() { getpeername = oldgetpeername }()

				getpeername = func(fd int) (unix.Sockaddr, error) {
					return nil, errors.New("failing getpeername")
				}

				Expect(NewSocketFd(fd, "socket:[123456]")).Error().To(
					MatchError("failing getpeername"))
			})

		})

	})

	It("returns correct socket inode number, domain, type, protocol", func() {
		fd, err := unix.Socket(unix.AF_UNIX, unix.SOCK_STREAM, 0)
		Expect(err).NotTo(HaveOccurred())
		defer unix.Close(fd)
		fdesc, err := NewSocketFd(fd, "socket:[123456]")
		Expect(err).NotTo(HaveOccurred())
		sockfd := fdesc.(*SocketFd)
		Expect(sockfd).To(HaveField("Ino()", uint64(123456)))
		Expect(sockfd).To(HaveField("Domain()", unix.AF_UNIX))
		Expect(sockfd).To(HaveField("Type()", unix.SOCK_STREAM))
		Expect(sockfd).To(HaveField("Protocol()", 0))
	})

	It("understands a unix socket", func() {
		By("creating a unix socket the hard way")
		fd, err := unix.Socket(unix.AF_UNIX, unix.SOCK_STREAM, 0)
		Expect(err).NotTo(HaveOccurred())
		defer unix.Close(fd)

		By("discovering the unbound (\"unnamed\") unix socket given only the fd")
		fdesc, err := New(fd)
		Expect(err).NotTo(HaveOccurred())
		sockfd := fdesc.(*SocketFd)
		Expect(sockfd.Listening()).To(BeFalse())
		Expect(sockfd.Name()).To(Equal("@")) // erm, sic!
		Expect(sockfd.Addr()).To(HaveField("Name", "@"))
		Expect(sockfd.Peer()).To(Equal(""))
		Expect(sockfd.PeerAddr()).To(BeNil())

		By("discovering the bound (\"named\") unix socket given only the fd")
		const abstractName = "@gfdleak/filedesc/fd_socket_test"
		Expect(unix.Bind(fd, &unix.SockaddrUnix{Name: abstractName})).
			NotTo(HaveOccurred())
		fdesc, err = New(fd)
		Expect(err).NotTo(HaveOccurred())
		Expect(fdesc.(*SocketFd).Name()).To(Equal(abstractName))

		By("listening, ...")
		Expect(unix.Listen(fd, 1)).NotTo(HaveOccurred())
		fdesc, err = New(fd)
		Expect(err).NotTo(HaveOccurred())
		Expect(fdesc.(*SocketFd).Listening()).To(BeTrue())
		Expect(fdesc.(*SocketFd).Description(0)).To(ContainSubstring(" listening "))

		By("...connecting, and accepting")
		fd2, err := unix.Socket(unix.AF_UNIX, unix.SOCK_STREAM, 0)
		Expect(err).NotTo(HaveOccurred())
		defer unix.Close(fd2)
		done := make(chan struct{})
		accepted := make(chan struct{})
		defer close(done)
		go func() {
			defer GinkgoRecover()
			connfd, _, err := unix.Accept(fd)
			close(accepted)
			Expect(err).NotTo(HaveOccurred())
			defer unix.Close(connfd)
			<-done
		}()
		Expect(unix.Connect(fd2, &unix.SockaddrUnix{Name: abstractName})).NotTo(HaveOccurred())
		<-accepted
		connfdesc, err := New(fd2)
		Expect(err).NotTo(HaveOccurred())
		connfd := connfdesc.(*SocketFd)
		Expect(connfd.Name()).To(Equal("@"))
		Expect(connfd.Peer()).To(Equal(abstractName))
		Expect(connfd.PeerAddr()).NotTo(BeNil())
		Expect(connfd.Description(0)).To(MatchRegexp(
			`(?m)fd \d+, flags 0x.* \(O_RDWR\)\n\s+socket\(AF_UNIX, SOCK_STREAM, protocol 0\), ino \d+\n\s+local "@"\n\s+peer "` + abstractName + `"`))

		By("checking (non-) equality")
		Expect(fdesc.Equal(fdesc)).To(BeTrue())
		Expect(fdesc.Equal(connfd)).To(BeFalse())
		Expect(fdesc.Equal(nil)).To(BeFalse())
	})

	It("understands an AF_INET socket", func() {
		By("creating an AF_INET socket the hard way")
		fd, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, 0)
		Expect(err).NotTo(HaveOccurred())
		defer unix.Close(fd)

		By("discovering the unbound (\"unnamed\") AF_INET socket given only the fd")
		fdesc, err := New(fd)
		Expect(err).NotTo(HaveOccurred())
		sfd := fdesc.(*SocketFd)
		Expect(sfd.Name()).To(Equal("0.0.0.0:0"))
		Expect(sfd.Peer()).To(Equal(""))
		Expect(sfd.Description(0)).To(MatchRegexp(
			`fd \d+, flags 0x.* \(O_RDWR\)\n\s+socket\(AF_INET, SOCK_DGRAM, IPPROTO_UDP\), ino \d+\n\s+local "0.0.0.0:0"`))
	})

	It("understands an AF_INET6 socket", func() {
		By("creating an AF_INET6 socket the hard way")
		fd, err := unix.Socket(unix.AF_INET6, unix.SOCK_DGRAM, 0)
		Expect(err).NotTo(HaveOccurred())
		defer unix.Close(fd)

		By("discovering the unbound (\"unnamed\") AF_INET6 socket given only the fd")
		fdesc, err := New(fd)
		Expect(err).NotTo(HaveOccurred())
		sfd := fdesc.(*SocketFd)
		Expect(sfd.Name()).To(Equal("[::]:0"))
		Expect(sfd.Description(0)).To(MatchRegexp(
			`fd \d+, flags 0x.* \(O_RDWR\)\n\s+socket\(AF_INET6, SOCK_DGRAM, IPPROTO_UDP\), ino \d+\n\s+local "\[::\]:0"`))
	})

})
