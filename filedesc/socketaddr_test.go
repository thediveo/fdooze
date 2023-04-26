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
	"fmt"
	"net"

	"golang.org/x/sys/unix"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("socket address", func() {

	It("returns an empty textual representation for empty wrapped socket address", func() {
		Expect(Sockaddr{}.String()).To(BeEmpty())
	})

	It("defaults to struct dumping", func() {
		a := Sockaddr{Sockaddr: &unix.SockaddrL2{}}
		Expect(a.String()).To(Equal(fmt.Sprintf("%#v", a.Sockaddr)))
	})

	It("textifies IP socket addresses", func() {
		a := Sockaddr{Sockaddr: &unix.SockaddrInet4{
			// yep, absolutely obvious expression ... :p
			// https://tip.golang.org/ref/spec#Conversions_from_slice_to_array_pointer
			//
			// ... oh, and beware of Go's net.IP habit to parse and store IPv4
			// addresses as IPv4-Mapped IPv6 addresses, see also:
			// https://datatracker.ietf.org/doc/html/rfc4291#section-2.5.5.2
			Addr: *(*[4]byte)(([]byte)(net.ParseIP("192.0.0.1").To4())),
			Port: 1234,
		}}
		Expect(a.String()).To(Equal("192.0.0.1:1234"))

		a = Sockaddr{Sockaddr: &unix.SockaddrInet6{
			// yep, absolutely obvious expression ... :p
			// https://tip.golang.org/ref/spec#Conversions_from_slice_to_array_pointer
			Addr:   *(*[16]byte)(([]byte)(net.ParseIP("fe80::dead:beef"))),
			Port:   1234,
			ZoneId: 666,
		}}
		Expect(a.String()).To(Equal("[fe80::dead:beef%666]:1234"))
	})

	It("textifies unix (domain) socket addresses", func() {
		a := Sockaddr{Sockaddr: &unix.SockaddrUnix{Name: "@foobar"}}
		Expect(a.String()).To(Equal(a.Sockaddr.(*unix.SockaddrUnix).Name))
	})

	DescribeTable("textifies data link-layer addresses",
		func(protocol int, packettype int, expected string) {
			a := Sockaddr{Sockaddr: &unix.SockaddrLinklayer{
				Ifindex:  1,
				Addr:     [8]byte{0xde, 0xad, 0xbe, 0xef, 0xca, 0xfe},
				Halen:    6,
				Protocol: uint16(protocol),
				Pkttype:  uint8(packettype),
			}}
			Expect(a.String()).To(Equal(expected))
		},
		Entry("unknown protocol and packet type", 0, 42, "DE:AD:BE:EF:CA:FE (HW address type 0x0)\nprotocol 0x0, interface index 1, packet type 42"),
		Entry("unknown protocol, known packet type", 0, unix.PACKET_HOST, "DE:AD:BE:EF:CA:FE (HW address type 0x0)\nprotocol 0x0, interface index 1, packet type PACKET_HOST"),
		Entry("known protocol, known packet type", unix.ETH_P_TSN, unix.PACKET_HOST, "DE:AD:BE:EF:CA:FE (HW address type 0x0)\nprotocol ETH_P_TSN, interface index 1, packet type PACKET_HOST"),
	)

	DescribeTable("textifies VM socket addresses with different CIDs",
		func(cid int, expected string) {
			a := Sockaddr{Sockaddr: &unix.SockaddrVM{
				Port:  12345678,
				Flags: 42,
				CID:   uint32(cid),
			}}
			Expect(a.String()).To(Equal(expected))
		},
		Entry("some CID", 1234, "port 12345678, CID 1234, flags 42"),
		Entry("hypervisor CID", unix.VMADDR_CID_HYPERVISOR, "port 12345678, VMADDR_CID_HYPERVISOR, flags 42"),
		Entry("local CID", unix.VMADDR_CID_LOCAL, "port 12345678, VMADDR_CID_LOCAL, flags 42"),
		Entry("host CID", unix.VMADDR_CID_HOST, "port 12345678, VMADDR_CID_HOST, flags 42"),
		Entry("any CID", unix.VMADDR_CID_ANY, "port 12345678, VMADDR_CID_ANY, flags 42"),
	)

	DescribeTable("textifies NETLINK socket addresses",
		func(pid int, expected string) {
			a := Sockaddr{Sockaddr: &unix.SockaddrNetlink{
				Pid:    uint32(pid),
				Groups: 0x123,
			}}
			Expect(a.String()).To(Equal(expected))
		},
		Entry("kernel", 0, "kernel, multicast groups mask 0x123"),
		Entry("kernel", 42, "(p)id 42, multicast groups mask 0x123"),
	)

	It("textifies XDP socket addresses", func() {
		a := Sockaddr{Sockaddr: &unix.SockaddrXDP{
			Flags:        05, // what ... octal ... is this a PDP 11 or what?!!
			Ifindex:      42,
			QueueID:      1,
			SharedUmemFD: 666,
		}}
		Expect(a.String()).To(Equal(
			"flags: 0x5 (XDP_SHARED_UMEM,XDP_ZEROCOPY), ifindex: 42, queue ID: 1, shared umem fd: 666"))

		a = Sockaddr{Sockaddr: &unix.SockaddrXDP{
			Flags:        0,
			Ifindex:      42,
			QueueID:      1,
			SharedUmemFD: 666,
		}}
		Expect(a.String()).To(Equal(
			"flags: 0x0, ifindex: 42, queue ID: 1, shared umem fd: 666"))
	})

})
