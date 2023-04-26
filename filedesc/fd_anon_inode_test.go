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
	"golang.org/x/sys/unix"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/success"
)

var _ = Describe("anonymous inode fd", func() {

	const fakeBase = "/proc/fake/fd"

	It("correctly fails for invalid fd number", func() {
		Expect(NewAnonInodeFd(-1, fakeBase, "anon_inode:[foobar]")).Error().
			To(HaveOccurred())
	})

	It("returns the correct anonymous inode file type and description", func() {
		fd := Successful(unix.Eventfd(42, unix.EFD_CLOEXEC))
		defer unix.Close(fd)

		fdesc := Successful(New(fd))
		anonfd := fdesc.(*AnonInodeFd)
		Expect(anonfd.FileType()).To(Equal("eventfd"))
		Expect(anonfd.Description(0)).To(MatchRegexp(
			`fd \d+, flags 0x.* \(O_RDWR,O_CLOEXEC\)\n\s+anonymous inode file type: "eventfd"`))
	})

	It("determines equality correctly", func() {
		fd := Successful(unix.Eventfd(42, unix.EFD_CLOEXEC))
		defer unix.Close(fd)

		fdesc := Successful(New(fd))
		Expect(fdesc.Equal(nil)).To(BeFalse())
		Expect(fdesc.Equal(fdesc)).To(BeTrue())

		fd0 := Successful(New(0))
		Expect(fdesc.Equal(fd0)).To(BeFalse())
	})

})
