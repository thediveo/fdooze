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
func NewPipeFd(fd int, link string) (FileDescriptor, error) {
	inoArg := strings.TrimSuffix(strings.TrimPrefix(link, "pipe:["), "]")
	ino, err := strconv.ParseUint(inoArg, 10, 64)
	if err != nil {
		return nil, err
	}
	filedesc, err := newFiledesc(fd)
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
