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
	"os"
	"syscall"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("fd flags", func() {

	It("returns correct flag access mode name", func() {
		Expect(Flags(os.O_RDONLY | syscall.O_CLOEXEC).Names()).To(ContainElement("O_RDONLY"))
		Expect(Flags(os.O_WRONLY | syscall.O_CLOEXEC).Names()).To(ContainElement("O_WRONLY"))
		Expect(Flags(os.O_RDWR | syscall.O_CLOEXEC).Names()).To(ContainElement("O_RDWR"))
		Expect(Flags(syscall.O_ACCMODE | syscall.O_CLOEXEC).Names()).To(ContainElement(MatchRegexp(`access mode \d`)))
	})

	It("returns correct flag names", func() {
		Expect(Flags(os.O_WRONLY | syscall.O_CLOEXEC | syscall.O_NOATIME).Names()).To(ConsistOf("O_WRONLY", "O_CLOEXEC", "O_NOATIME"))
		Expect(Flags(os.O_WRONLY | syscall.O_APPEND).Names()).To(ConsistOf("O_WRONLY", "O_APPEND"))
	})

	It("handles Linux flag oddballs correctly", func() {
		Expect(Flags(os.O_WRONLY | O_TMPFILE).Names()).To(ConsistOf("O_WRONLY", "O_TMPFILE"))
		Expect(Flags(os.O_WRONLY | syscall.O_DIRECTORY).Names()).To(ConsistOf("O_WRONLY", "O_DIRECTORY"))

		Expect(Flags(os.O_WRONLY | syscall.O_DSYNC).Names()).To(ConsistOf("O_WRONLY", "O_DSYNC"))
		Expect(Flags(os.O_WRONLY | syscall.O_SYNC).Names()).To(ConsistOf("O_WRONLY", "O_SYNC"))
	})

})
