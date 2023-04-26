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

var _ = Describe("pipe fd", func() {

	const fakeBase = "/proc/fake/fd"

	It("correctly fails for invalid fd number or pipe inode number", func() {
		Expect(NewPipeFd(-1, fakeBase, "foobar")).Error().To(HaveOccurred())
		Expect(NewPipeFd(-1, fakeBase, "pipe:[123456]")).Error().To(HaveOccurred())
	})

	When("given pipe ends", Ordered, func() {

		var pipefds [2]int

		BeforeAll(func() {
			By("creating a pair of unnamed pipe ends")
			Expect(unix.Pipe(pipefds[:])).To(Succeed())
			Expect(pipefds).To(HaveEach(Not(BeZero())))
			DeferCleanup(func() {
				unix.Close(pipefds[0])
				unix.Close(pipefds[1])
			})
		})

		It("returns correct pipe details", func() {
			rfdesc := Successful(New(pipefds[0]))
			Expect(rfdesc.(*PipeFd)).NotTo(BeNil())
			Expect(rfdesc.Description(0)).To(MatchRegexp(
				"(?m)fd %d, flags 0x0 \\(O_RDONLY\\)\n\\s+pipe inode number: \\d+",
				pipefds[0]))

			wfdesc := Successful(New(pipefds[1]))
			Expect(wfdesc.(*PipeFd)).NotTo(BeNil())
			Expect(wfdesc.Description(0)).To(MatchRegexp(
				"(?m)fd %d, flags 0x1 \\(O_WRONLY\\)\n\\s+pipe inode number: \\d+",
				pipefds[1]))

			Expect(rfdesc.(*PipeFd).Ino()).To(Equal(wfdesc.(*PipeFd).Ino()))
		})

		It("determines equality correctly", func() {
			rfdesc := Successful(New(pipefds[0]))
			Expect(rfdesc.(*PipeFd)).NotTo(BeNil())
			wfdesc := Successful(New(pipefds[1]))
			Expect(wfdesc.(*PipeFd)).NotTo(BeNil())

			Expect(rfdesc.Equal(nil)).To(BeFalse())
			Expect(rfdesc.Equal(wfdesc)).To(BeFalse())
			Expect(rfdesc.Equal(rfdesc)).To(BeTrue())
		})

	})

})
