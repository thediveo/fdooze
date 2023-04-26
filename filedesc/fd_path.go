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

import "fmt"

// PathFd implements FileDescriptor for an fd with a path to a regular file,
// directory, device, ... in the VFS.
type PathFd struct {
	filedesc
	path string // just a plain and simple absolute path.
}

// NewPathFd returns a new FileDescriptor for an fd with an ordinary file system
// path. The link argument specifies the (absolute) file system path.
func NewPathFd(fdNo int, base string, linkDest string) (FileDescriptor, error) {
	filedesc, err := newFiledesc(fdNo, base)
	if err != nil {
		return nil, err
	}
	return &PathFd{
		filedesc: filedesc,
		path:     linkDest,
	}, nil
}

// Path returns the path name this fd references.
func (p PathFd) Path() string { return p.path }

// Description returns a pretty formatted multi-line textual description
// detailing the fd number, flags, and path.
func (p PathFd) Description(indentation uint) string {
	indent := Indentation(indentation + 1) // further details are always indented further
	return p.filedesc.Description(indentation) +
		fmt.Sprintf("\n%spath: %q", indent, p.path)
}

// Equal returns true, if other is a pathFd with the same fd number and mount
// ID, as well as the same filename/path.
func (p PathFd) Equal(other FileDescriptor) bool {
	o, ok := other.(*PathFd)
	if !ok {
		return false
	}
	return p.filedesc.Equal(&o.filedesc) &&
		p.path == o.path
}
