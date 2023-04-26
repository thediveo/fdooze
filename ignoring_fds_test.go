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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("IgnoringFds matcher", func() {

	It("correctly handles an invalid actual value", func() {
		m := IgnoringFiledescriptors(nil)
		Expect(m.Match(nil)).Error().To(HaveOccurred())
		Expect(m.Match(42)).Error().To(HaveOccurred())
	})

	It("returns correct failure messages", func() {
		fds := Filedescriptors()
		expected := []FileDescriptor{fds[0], fds[2]}
		actual := []FileDescriptor{fds[1]}
		m := IgnoringFiledescriptors(expected)
		Expect(m.FailureMessage(actual)).To(MatchRegexp(
			`(?s)Expected
\s+<.*>: \[.*\]
to be contained in the list of expected file descriptors
\s+fd \d+, .*
\s+fd \d+, .*`))
		Expect(m.NegatedFailureMessage(actual)).To(MatchRegexp(
			`(?s)Expected
\s+<.*>: \[.*\]
not to be contained in the list of expected file descriptors
\s+fd \d+, .*
\s+fd \d+, .*`))
	})

})
