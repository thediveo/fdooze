package fdooze

import (
	"fmt"

	"github.com/onsi/gomega/types"
	"github.com/thediveo/fdooze/filedesc"
)

// HaveLeakedFds succeeds if after filtering out expected file descriptors from
// the list of actual file descriptors the remaining list is non-empty. The file
// descriptors not filtered out are considered to have been leaked.
//
// Optional additional filter matchers can be specified that can filter out use
// case-specific file descriptors based on various fd properties. Please refer
// to the filedesc package for details about the defined FileDescriptor
// implementations for various types of file descriptors.
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
// A typical way to check for leaked ("oozed") file descriptors is as follows:
//
//     BeforeEach(func() {
//         goodfds := Filedescriptors()
//         DeferCleanup(func() {
//             Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
//         })
//     })
//
// HaveLeakedFds accepts optional Gomega matchers that it will repeatedly pass
// FileDescriptor values to: if a matcher succeeds, the particular file
// descriptor is considered not to be leaked and thus filtered out. Especially
// Goemega's HaveField matcher can be quite useful in covering specific use
// cases where the otherwise straightforward before-after fd comparism isn't
// enough.
func HaveLeakedFds(fds []filedesc.FileDescriptor, ignoring ...types.GomegaMatcher) types.GomegaMatcher {
	m := &haveLeakedFdsMatcher{
		filters: append([]types.GomegaMatcher{
			IgnoringFiledescriptors(fds),
		}, ignoring...),
	}
	return m
}

type haveLeakedFdsMatcher struct {
	filters []types.GomegaMatcher
	leaked  []filedesc.FileDescriptor
}

func (matcher *haveLeakedFdsMatcher) Match(actual interface{}) (success bool, err error) {
	actualFds, err := toFds(actual, "HaveLeakedFds")
	if err != nil {
		return false, err
	}
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
