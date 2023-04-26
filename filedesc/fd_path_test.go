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

var _ = Describe("file path fd", func() {

	const fakeBase = "/proc/fake/fd"

	It("fails when given an invalid fd number", func() {
		Expect(NewPathFd(-1, fakeBase, "/foobar")).Error().
			To(HaveOccurred())
	})

	It("returns correct path information", func() {
		fd := Successful(unix.Open("fd_path_test.go", unix.O_RDONLY, 0))
		defer unix.Close(fd)

		fdesc := Successful(New(fd))
		Expect(fdesc).To(HaveField("Path()", MatchRegexp("/filedesc/fd_path_test.go$")))
		Expect(fdesc.Description(0)).To(MatchRegexp(
			"(?m)fd %d, flags .* \\(O_RDONLY\\)\n\\s+path: \".*/fd_path.test.go\"",
			fd))
	})

	It("determines equality correctly", func() {
		fd := Successful(unix.Open("fd_path_test.go", unix.O_RDONLY, 0))
		defer unix.Close(fd)

		fdesc := Successful(New(fd))
		Expect(fdesc.Equal(nil)).To(BeFalse())
		Expect(fdesc.Equal(fdesc)).To(BeTrue())

		fd0 := Successful(New(0))
		Expect(fdesc.Equal(fd0)).To(BeFalse())
	})

})
