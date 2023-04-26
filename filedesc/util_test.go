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
	"strings"

	"github.com/onsi/gomega/format"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("filedesc utilities", func() {

	It("returns an indentation string given an level of indentation", func() {
		Expect(Indentation(2)).To(Equal(strings.Repeat(format.Indent, 2)))
	})

	It("hangs the indentation", func() {
		Expect(HangingIndent("", 1)).To(Equal(Indentation(1)))
		Expect(HangingIndent("foo", 1)).To(Equal(
			Indentation(1) + "foo"))
		Expect(HangingIndent("foo\nbar", 1)).To(Equal(
			Indentation(1) + "foo\n" +
				Indentation(2) + "bar"))
		Expect(HangingIndent("foo\nbar\nbaz", 1)).To(Equal(
			Indentation(1) + "foo\n" +
				Indentation(2) + "bar\n" +
				Indentation(2) + "baz"))
	})

})
