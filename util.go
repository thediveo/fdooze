//go:build linux

package fdooze

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/onsi/gomega/format"
)

var fdsT = reflect.TypeOf([]FileDescriptor{})

// toFds returns actual as a slice of FileDescriptors, or an error if actual
// isn't a slice of FileDescriptors. matchername specifies the name of the
// matcher to be included in the error message in case of an invalid actual
// type.
func toFds(actual interface{}, matchername string) ([]FileDescriptor, error) {
	val := reflect.ValueOf(actual)
	switch val.Kind() {
	case reflect.Array, reflect.Slice:
		if !val.Type().AssignableTo(fdsT) {
			return nil, fmt.Errorf(
				"%s matcher expects an array or slice of file descriptors.  Got:\n%s",
				matchername, format.Object(actual, 1))
		}
	default:
		return nil, fmt.Errorf(
			"%s matcher expects an array or slice of file descriptors.  Got:\n%s",
			matchername, format.Object(actual, 1))
	}
	return val.Convert(fdsT).Interface().([]FileDescriptor), nil
}

// dumpFds returns detailed textual information about the specified (leaked)
// fds. The fds are numerically sorted in the dump by their file descriptor
// numbers.
func dumpFds(fds []FileDescriptor, indentation uint) string {
	sort.Slice(fds, func(a, b int) bool { return fds[a].Fd() < fds[b].Fd() })
	var out strings.Builder
	for idx, fd := range fds {
		if idx > 0 {
			out.WriteRune('\n')
		}
		out.WriteString(fd.Description(indentation))
	}
	return out.String()
}
