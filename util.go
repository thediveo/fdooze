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
	"reflect"
	"strings"

	"golang.org/x/exp/slices"

	"github.com/onsi/gomega/format" // That's fine ... because this is a package used only in tests anyway
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
	slices.SortFunc(fds, func(a, b FileDescriptor) int { return a.FdNo() - b.FdNo() })
	var out strings.Builder
	for idx, fd := range fds {
		if idx > 0 {
			out.WriteRune('\n')
		}
		out.WriteString(fd.Description(indentation))
	}
	return out.String()
}
