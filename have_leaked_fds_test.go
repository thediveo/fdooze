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

package fdooze

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("HaveLeakedFds matcher", func() {

	It("fails for invalid actual", func() {
		m := HaveLeakedFds(nil)
		Expect(m.Match(nil)).Error().To(HaveOccurred())
		Expect(m.Match(42)).Error().To(HaveOccurred())
	})

	It("fails when filter fails", func() {
		m := HaveLeakedFds(nil, HaveField("Foo", 42))
		Expect(m.Match(Filedescriptors())).Error().To(HaveOccurred())
	})

	It("doesn't trigger a false positive", func() {
		goods := Filedescriptors()
		Expect(goods).NotTo(BeEmpty())
		oozed, err := HaveLeakedFds(goods).Match(goods)
		Expect(err).NotTo(HaveOccurred())
		Expect(oozed).To(BeFalse())
	})

	It("detects and details a leaked fd", func() {
		goods := Filedescriptors()
		Expect(goods).NotTo(BeEmpty())

		f, err := os.Open("have_leaked_fds_test.go")
		Expect(err).NotTo(HaveOccurred())
		defer f.Close()

		m := HaveLeakedFds(goods)
		oozed, err := m.Match(Filedescriptors())
		Expect(err).NotTo(HaveOccurred())
		Expect(oozed).To(BeTrue())
		Expect(m.FailureMessage(nil)).To(MatchRegexp(
			`(?m)Expected to leak \d+ file descriptors:
\s+fd \d+, flags 0x.* \(O_RDONLY,O_CLOEXEC\)
\s+path: ".*/have_leaked_fds_test.go"`))
		Expect(m.NegatedFailureMessage(nil)).To(MatchRegexp(
			`(?m)Expected not to leak \d+ file descriptors:
\s+fd \d+, flags 0x.* \(O_RDONLY,O_CLOEXEC\)
\s+path: ".*/have_leaked_fds_test.go"`))
	})

})
