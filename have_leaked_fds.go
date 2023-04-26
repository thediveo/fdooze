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
	"fmt"

	"github.com/onsi/gomega/types"
)

// HaveLeakedFds succeeds if after filtering out expected file descriptors from
// the list of actual file descriptors the remaining list is non-empty. The file
// descriptors not filtered out are considered to have been leaked.
//
// Optional additional filter matchers can be specified that can filter out use
// case-specific file descriptors based on various fd properties. Please refer
// to the [github.com/thediveo/fdooze/filedesc] package for details about the
// defined [github.com/thediveo/fdooze/filedesc.FileDescriptor] implementations
// for various types of file descriptors.
//
// File descriptors are identified not only based on the fd number, but also
// additional associated information, such as a file path they link to, type,
// socket inode number, et cetera. As fd numbers tend to quickly get reused this
// allows detecting changed fds in many (if not most) situations with enough
// accuracy.
//
// HaveLeakedFds does not assume any well-known fds, and in particular, it does
// not make any assumptions about fds with numbers 0, 1, 2.
//
// A typical way to check for leaked (“oozed”) file descriptors is as follows,
// after dot-importing the fdooze package:
//
//	 import . "github.com/thediveo/fdooze"
//
//	 var _ = Describe("...", {
//		  BeforeEach(func() {
//		      goodfds := Filedescriptors()
//		      DeferCleanup(func() {
//		          Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
//		      })
//		  })
//	 })
//
// HaveLeakedFds accepts optional Gomega matchers (of type
// [types.GomegaMatcher]) that it will repeatedly pass FileDescriptor values to:
// if a matcher succeeds, the particular file descriptor is considered not to be
// leaked and thus filtered out. Especially Gomega's [HaveField] matcher can be
// quite useful in covering specific use cases where the otherwise
// straightforward before-after fd comparism isn't enough.
//
// [HaveField]: https://onsi.github.io/gomega/#havefieldfield-interface-value-interface
func HaveLeakedFds(fds []FileDescriptor, ignoring ...types.GomegaMatcher) types.GomegaMatcher {
	m := &haveLeakedFdsMatcher{
		filters: append([]types.GomegaMatcher{
			IgnoringFiledescriptors(fds),
		}, ignoring...),
	}
	return m
}

type haveLeakedFdsMatcher struct {
	filters []types.GomegaMatcher
	leaked  []FileDescriptor
}

func (matcher *haveLeakedFdsMatcher) Match(actual interface{}) (success bool, err error) {
	actualFds, err := toFds(actual, "HaveLeakedFds")
	if err != nil {
		return false, err
	}
	matcher.leaked = nil
nextFd:
	for _, actualFd := range actualFds {
		for _, filter := range matcher.filters {
			matches, err := filter.Match(actualFd)
			if err != nil {
				return false, err
			}
			if matches {
				continue nextFd
			}
		}
		matcher.leaked = append(matcher.leaked, actualFd)
	}
	if len(matcher.leaked) == 0 {
		return false, nil
	}
	return true, nil // we have leak(ed)
}

// FailureMessage returns a failure message if there are leaked file
// descriptors, listing the leaked fds with (some) detail information.
func (matcher *haveLeakedFdsMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected to leak %d file descriptors:\n%s",
		len(matcher.leaked), dumpFds(matcher.leaked, 1))
}

// NegatedFailureMessage returns a negated failure message if there aren't any
// leaked file descriptors.
func (matcher *haveLeakedFdsMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected not to leak %d file descriptors:\n%s",
		len(matcher.leaked), dumpFds(matcher.leaked, 1))
}
