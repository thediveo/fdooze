//go:build linux

package filedesc

import (
	"fmt"
	"strings"
)

const anonInodePrefix = "anon_inode:"

// AnonInodeFd implements FileDescriptor for an fd for an anonymous inode of
// some “file” type, such as event fds, timer fds, et cetera. This is a generic,
// catch-all implementation to be used for any file type of anonymous inode
// where we don't define a dedicated type.
type AnonInodeFd struct {
	filedesc
	ftype string // "file" type of anonymous inode, without any enclosing square brackets.
}

// NewAnonInodeFd returns a new FileDescriptor for an fd for an “anonymous
// inode”.
func NewAnonInodeFd(fd int, link string) (FileDescriptor, error) {
	filedesc, err := newFiledesc(fd)
	if err != nil {
		return nil, err
	}
	return &AnonInodeFd{
		filedesc: filedesc,
		ftype:    strings.Trim(link[len(anonInodePrefix):], "[]"),
	}, nil
}

// FileType returns the “file type” of this anonymous inode.
func (a AnonInodeFd) FileType() string { return a.ftype }

// Description returns a pretty formatted multi-line textual description
// detailing the fd number, flags, and “file type” of anonymous node.
func (a AnonInodeFd) Description(indentation uint) string {
	indent := Indentation(indentation + 1) // further details are always indented further
	return a.filedesc.Description(indentation) +
		fmt.Sprintf("\n%sanonymous inode file type: %q", indent, a.ftype)
}

// Equal returns true, if other is also an anonymous inode of the same type and
// with the same fd number (and mount ID).
func (a AnonInodeFd) Equal(other FileDescriptor) bool {
	o, ok := other.(*AnonInodeFd)
	if !ok {
		return false
	}
	return a.filedesc.Equal(&o.filedesc) &&
		a.ftype == o.ftype
}
