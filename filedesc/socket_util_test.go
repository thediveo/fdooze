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

	"golang.org/x/sys/unix"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
