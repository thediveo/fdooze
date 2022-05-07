//go:build linux

package fdooze

import (
	"fmt"

	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
	"github.com/thediveo/fdooze/filedesc"
)

// IgnoringFiledescriptors succeeds if an actual FileDescriptor in contained in
// a slice of expected file descriptors. An actual FileDescriptor is considered
// to be contained, if the slice must contains a FileDescriptor with the same fd
// number and FileDescriptor.Equal considers both file descriptors to be equal.
//
// Please note that fd flags and file offsets are ignored when testing for
// equality, in order to avoid spurious false positives.
func IgnoringFiledescriptors(fds []filedesc.FileDescriptor) types.GomegaMatcher {
	m := &ignoringFds{
		ignoreFds: map[int]filedesc.FileDescriptor{},
	}
	for _, fd := range fds {
		m.ignoreFds[fd.Fd()] = fd
	}
	return m
}

type ignoringFds struct {
	ignoreFds map[int]filedesc.FileDescriptor
}

// Match succeeds if actual is a FileDescriptor contained in the set of expected
// file descriptors. Containment uses FileDescriptor.Equal to test for file
// descriptor equality.
func (matcher *ignoringFds) Match(actual interface{}) (success bool, err error) {
	actualFd, ok := actual.(filedesc.FileDescriptor)
	if !ok {
		return false, fmt.Errorf(
			"IgnoringFiledescriptor matcher expects a filedesc.FileDescriptor.  Got:\n%s",
			format.Object(actual, 1))
	}
	fd, ok := matcher.ignoreFds[actualFd.Fd()]
	if !ok {
		return false, nil
	}
	return actualFd.Equal(fd), nil
}

// FailureMessage returns a failure message if the actual file descriptor isn't
// in the set of file descriptors to be ignored.
func (matcher *ignoringFds) FailureMessage(actual interface{}) (message string) {
	expected := make([]filedesc.FileDescriptor, 0, len(matcher.ignoreFds))
	for _, fd := range matcher.ignoreFds {
		expected = append(expected, fd)
	}
	return fmt.Sprintf("Expected\n%s\nto be contained in the list of expected file descriptors\n%s",
		format.Object(actual, 1),
		dumpFds(expected, 1))
}

// NegatedFailureMessage returns a failure message if the actual file descriptor
// actually is in the set of file descriptors to be ignored.
func (matcher *ignoringFds) NegatedFailureMessage(actual interface{}) (message string) {
	expected := make([]filedesc.FileDescriptor, 0, len(matcher.ignoreFds))
	for _, fd := range matcher.ignoreFds {
		expected = append(expected, fd)
	}
	return fmt.Sprintf("Expected\n%s\nnot to be contained in the list of expected file descriptors\n%s",
		format.Object(actual, 1),
		dumpFds(expected, 1))
}
