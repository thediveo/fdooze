//go:build linux

package fdooze

import "github.com/thediveo/fdooze/filedesc"

// FileDescriptor describes a Linux "fd" file descriptor in more detail than
// just its fd int number; it is a type alias of [filedesc.FileDescriptor].
type FileDescriptor = filedesc.FileDescriptor

// Filedescriptors returns the list of currently open file descriptors for this
// process.
func Filedescriptors() []FileDescriptor {
	return filedesc.Filedescriptors()
}
