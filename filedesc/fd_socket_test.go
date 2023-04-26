// Copyright 2022 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy
// of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

//go:build linux

package filedesc

import (
	"errors"
	"fmt"
	"os"

	"golang.org/x/sys/unix"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/success"
)

var _ = Describe("socket descriptors", func() {

	const procFdBase = "/proc/self/fd"

	It("handles invalid fd number or socket inode number", func() {
		Expect(NewSocketFd(0, procFdBase, "socket:[abc]")).Error().To(HaveOccurred())
		Expect(NewSocketFd(-1, procFdBase, "socket:[123456]")).Error().To(HaveOccurred())
		Expect(NewSocketFd(0, procFdBase, "socket:[123456]")).Error().To(HaveOccurred())
	})

	Context("invalid base", func() {

		BeforeEach(func() {
			cwd := Successful(os.Getwd())
			Expect(os.Chdir("test")).To(Succeed())
			DeferCleanup(func() {
				os.Chdir(cwd)
			})
		})

		It("reports invalid base", func() {
			Expect(NewSocketFd(0, "proc/bar/fd", "socket:[123456]")).Error().To(
				MatchError(ContainSubstring("invalid fd base")))
		})

		It("reports invalid PID in base", func() {
			Expect(NewSocketFd(0, "./proc/bar/fd", "socket:[123456]")).Error().To(
				MatchError(ContainSubstring("invalid syntax")))
			Expect(NewSocketFd(0, "./proc/0/fd", "socket:[123456]")).Error().To(
				MatchError(ContainSubstring("invalid argument")))
		})

		It("reports when not able to get fd of other process", func() {
			if os.Getuid() == 0 {
				Skip("needs non-root")
			}
			Expect(NewSocketFd(0, "./proc/1/fd", "socket:[123456]")).Error().To(
				MatchError(ContainSubstring("operation not permitted")))
		})

	})

	When("mocking failing socket syscalls", Serial, func() {

		var sockfd int

		BeforeEach(func() {
			sockfd = Successful(unix.Socket(unix.AF_UNIX, unix.SOCK_STREAM, 0))
			DeferCleanup(func() {
				unix.Close(sockfd)
			})
		})

		DescribeTable("reporting when failing to get some socket options", Ordered,
			func(failingOpt int, fail bool) {
				// mock getsockoptInt to fail only for the specified failingOpt,
				// but otherwise pass the call successfully on to the stdlib
				// implementation.
				oldgetsockoptInt := getsockoptInt
				defer func() { getsockoptInt = oldgetsockoptInt }()

				getsockoptInt = func(fd, level, opt int) (int, error) {
					if level != unix.SOL_SOCKET || opt != failingOpt {
						return oldgetsockoptInt(fd, level, opt)
					}
					return 0, fmt.Errorf("failing option %d", opt)
				}

				if fail {
					Expect(NewSocketFd(sockfd, procFdBase, "socket:[123456]")).Error().To(
						MatchError(fmt.Errorf("failing option %d", failingOpt)))
				} else {
					Expect(NewSocketFd(sockfd, procFdBase, "socket:[123456]")).Error().To(Succeed())
				}
			},
			Entry("failing SO_DOMAIN", unix.SO_DOMAIN, true),
			Entry("failing SO_TYPE", unix.SO_TYPE, true),
			Entry("failing SO_PROTOCOL", unix.SO_PROTOCOL, true),
			Entry("failing SO_ACCEPTCONN", unix.SO_ACCEPTCONN, false),
		)

		It("accepts Getsockname to fail", func() {
			oldgetsockname := getsockname
			defer func() { getsockname = oldgetsockname }()

			getsockname = func(fd int) (unix.Sockaddr, error) {
				return nil, errors.New("failing getsockname")
			}

			Expect(NewSocketFd(sockfd, procFdBase, "socket:[123456]")).Error().
				NotTo(HaveOccurred())
		})

		It("accepts Getpeername to fail", func() {
			oldgetpeername := getpeername
			defer func() { getpeername = oldgetpeername }()

			getpeername = func(fd int) (unix.Sockaddr, error) {
				return nil, errors.New("failing getpeername")
			}

			Expect(NewSocketFd(sockfd, procFdBase, "socket:[123456]")).Error().
				NotTo(HaveOccurred())
		})

	})

	It("returns correct socket inode number, domain, type, protocol", func() {
		fd := Successful(unix.Socket(unix.AF_UNIX, unix.SOCK_STREAM, 0))
		defer unix.Close(fd)
		fdesc := Successful(NewSocketFd(
			fd, "/proc/self/fd", "socket:[123456]"))
		sockfd := fdesc.(*SocketFd)
		Expect(sockfd).To(HaveField("Ino()", uint64(123456)))
		Expect(sockfd).To(HaveField("Domain()", unix.AF_UNIX))
		Expect(sockfd).To(HaveField("Type()", unix.SOCK_STREAM))
		Expect(sockfd).To(HaveField("Protocol()", 0))
	})

	Context("various address families", func() {

		It("understands a unix socket", func() {
			By("creating a unix socket the hard way")
			fd := Successful(unix.Socket(unix.AF_UNIX, unix.SOCK_STREAM, 0))
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

})
