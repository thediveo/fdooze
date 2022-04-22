package fdooze

import "github.com/thediveo/fdooze/filedesc"

// Filedescriptors returns the list of currently open file descriptors for this
// process.
func Filedescriptors() []filedesc.FileDescriptor {
	return filedesc.Filedescriptors()
}
