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
	"fmt"
	"strconv"
	"strings"
)

// PipeFd implements the FileDescriptor interface for an fd representing a pipe,
// as created by the pipe and pipe2 syscalls. See also pipe(2).
//
// Pipes are “unnamed” or “anonymous” and should not be confused with fifos,
// the latter being accessed as part of the file system. While pipes are
// identified by inodes, these inodes come from a special “pipefs” virtual file
// system. The mounted pipefs isn't visible in the VFS and thus cannot be
// viewed. It only serves for managing pipe inodes.
//
// For pipefs, see also:
// https://www.linux.org/threads/pipefs-sockfs-debugfs-and-securityfs.9638/
type PipeFd struct {
	filedesc
	ino uint64 // pipe's inode number from the (single) pipefs instance.
}

// NewPipeFd returns a new FileDescriptor for a pipe fd.
func NewPipeFd(fdNo int, base string, linkDest string) (FileDescriptor, error) {
	inoArg := strings.TrimSuffix(strings.TrimPrefix(linkDest, "pipe:["), "]")
	ino, err := strconv.ParseUint(inoArg, 10, 64)
	if err != nil {
		return nil, err
	}
	filedesc, err := newFiledesc(fdNo, base)
	if err != nil {
		return nil, err
	}
	return &PipeFd{
		filedesc: filedesc,
		ino:      ino,
	}, nil
}

// Ino returns the inode number uniquely identifying this pipe.
func (p PipeFd) Ino() uint64 { return p.ino }

// Description returns a pretty formatted multi-line textual description
// detailing the fd number, flags, and path.
func (p PipeFd) Description(indentation uint) string {
	indent := Indentation(indentation + 1) // further details are always indented further
	desc := p.filedesc.Description(indentation) +
		fmt.Sprintf("\n%spipe inode number: %d", indent, p.ino)
	return desc
}

// Equal returns true, if other is a pipeFd with the same fd number and mount
// ID, as well as the same inode number.
func (p PipeFd) Equal(other FileDescriptor) bool {
	o, ok := other.(*PipeFd)
	if !ok {
		return false
	}
	return p.filedesc.Equal(&o.filedesc) &&
		p.ino == o.ino
}
