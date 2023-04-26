//go:build linux

package filedesc

import (
	"fmt"
	"os"
	"syscall"
)

// Flags specifies a FileDescriptor's flags (mostly as a bit set, with the
// exception of the access mode 2-bit field). It additionally implements
// Stringer returning the known set flags with their symbolic constant names.
type Flags int

// Names returns the known symbolic constant names for the set bit(s).
//
// Please note that the “oddball” multi-bit fields and combinations are handled
// especially and correctly, such as the access mode bits,
// O_TMPFILE/O_DIRECTORY, and O_DSYNC/O_SYNC.
func (f Flags) Names() []string {
	n := make([]string, 0)
	// O_RDONLY, O_WRONLY, and O_RDWR are not bits, but instead elements of a
	// O_ACCMODE two-bit enumeration field.
	switch int(f) & (syscall.O_ACCMODE) {
	case os.O_RDONLY:
		n = append(n, "O_RDONLY")
	case os.O_WRONLY:
		n = append(n, "O_WRONLY")
	case os.O_RDWR:
		n = append(n, "O_RDWR")
	default:
		n = append(n, fmt.Sprintf("access mode %d", int(f)&(syscall.O_ACCMODE)))
	}
	// The single bit flags.
	for flagbit, name := range flagNames {
		if int(f)&flagbit == flagbit {
			n = append(n, name)
		}
	}
	// O_TMPFILE is a Linux oddball that includes O_DIRECTORY, so we handle this
	// as special cases.
	switch int(f) & O_TMPFILE {
	case syscall.O_DIRECTORY:
		n = append(n, "O_DIRECTORY")
	case O_TMPFILE:
		n = append(n, "O_TMPFILE")
	}
	// O_DSYNC/O_SYNC are the Linux oddballs that need their own special
	// treatment.
	switch int(f) & syscall.O_SYNC {
	case syscall.O_DSYNC:
		n = append(n, "O_DSYNC")
	case syscall.O_SYNC:
		n = append(n, "O_SYNC")
	}
	return n
}

// O_TMPFILE creates an unnamed(!) temporary regular(!) file. See also
// https://man7.org/linux/man-pages/man2/open.2.html.
const O_TMPFILE = 020000000 | syscall.O_DIRECTORY

// flagNames maps O_ flag values (bit(s)) to their textual names. Please note:
//   - O_DSYNC and O_SYNC need to be handled especially due to some history of Linux,
//   - O_FSYNC = O_SYNC = O_RSYNC,
//   - O_NDELAY = O_NONBLOCK.
var flagNames = map[int]string{
	os.O_APPEND:        "O_APPEND",
	syscall.O_ASYNC:    "O_ASYNC",
	syscall.O_CLOEXEC:  "O_CLOEXEC",
	os.O_CREATE:        "O_CREAT(E)",
	syscall.O_DIRECT:   "O_DIRECT",
	os.O_EXCL:          "O_EXCL",
	syscall.O_NOATIME:  "O_NOATIME",
	syscall.O_NOCTTY:   "O_NOCTTY",
	syscall.O_NOFOLLOW: "O_NOFOLLOW",
	syscall.O_NONBLOCK: "O_NONBLOCK",
	os.O_TRUNC:         "O_TRUNC",
}
