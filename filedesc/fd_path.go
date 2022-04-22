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
func NewPathFd(fd int, link string) (FileDescriptor, error) {
	filedesc, err := newFiledesc(fd)
	if err != nil {
		return nil, err
	}
	return &PathFd{
		filedesc: filedesc,
		path:     link,
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
