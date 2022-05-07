//go:build linux

package filedesc

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"math/bits"
	"os"
	"strconv"
	"strings"
)

// FileDescriptor describes a Linux "fd" file descriptor in more detail than
// just its fd int number. It describes the type of file descriptor and then
// type-specific properties.
type FileDescriptor interface {
	Fd() int                             // file descriptor number
	Description(indentation uint) string // pretty multi-line description
	Equal(other FileDescriptor) bool     // compare this file descriptor with another one
}

// Filedescriptors returns the list of currently open file descriptors for this
// process in form of FileDescriptor objects.
//
// Note: it is not possible to atomically read both the fd link itself as well
// as the associated fd information, as these are two separate procfs nodes,
// there's always the potential for a race condition when the fd state hasn't
// settled (yet).
func Filedescriptors() []FileDescriptor {
	fds, _ := filedescriptors("/proc/self/fd")
	return fds
}

// Filedescriptors returns the list of currently open file descriptors in form
// of FileDescriptor objects for the process identified by pid. If the calling
// process does not possess the necessary access rights to the process
// identified by pid an error is returned instead.
func ProcessFiledescriptors(pid int) ([]FileDescriptor, error) {
	return filedescriptors(fmt.Sprintf("/proc/%d/fd", pid))
}

// internal implementation to discovery file descriptors that can be tested
// using fake proc file systems.
func filedescriptors(fdDirPath string) ([]FileDescriptor, error) {
	// Don't use ioutil.ReadDir as it will **incorrectly sort** the fd numbers!
	fdfilesdir, err := os.Open(fdDirPath)
	if err != nil {
		return nil, err
	}
	defer fdfilesdir.Close()
	// As we now read the open fds from our process's fd directory, we cannot
	// avoid but to include this directory read fd also, so we need to skip and
	// drop it later when fetching fd details.
	fdfiles, err := fdfilesdir.ReadDir(-1)
	if err != nil {
		return nil, err
	}
	fds := make([]FileDescriptor, 0, len(fdfiles)-1)
	skipDirectoryFd := int(fdfilesdir.Fd())
	for _, fdfile := range fdfiles {
		fd, err := strconv.Atoi(fdfile.Name())
		if err != nil || fd == skipDirectoryFd {
			continue
		}
		fdesc, err := newWithBase(fd, fdDirPath)
		if err != nil {
			continue
		}
		fds = append(fds, fdesc)
	}
	return fds, nil
}

// New returns a FileDescriptor for the fd number specified. The information
// about the specified fd is gathered from the procfs filesystem mounted on
// /proc.
func New(fd int) (FileDescriptor, error) {
	return NewForPID(fd, os.Getpid())
}

// NewForPID returns a FileDescriptor for the process identified by pid and the
// particular fd number.
func NewForPID(fd int, pid int) (FileDescriptor, error) {
	return newWithBase(fd, fmt.Sprintf("/proc/%d/fd", pid))
}

// newWithBase returns a FileDescriptor for the fd of the process in the procfs
// with the base path.
func newWithBase(fd int, base string) (FileDescriptor, error) {
	link, err := os.Readlink(fmt.Sprintf("%s/%d", base, fd))
	if err != nil {
		return nil, err
	}
	return new(fd, link)
}

// new returns a new FileDescriptor for the specified fd number, corresponding
// with the specified link.
func new(fd int, link string) (FileDescriptor, error) {
	// Is this one of the various anonymous inode fd types?
	if strings.HasPrefix(link, anonInodePrefix) {
		return NewAnonInodeFd(fd, link)
	}
	// Is this one of the links with an embedded file type and inode number?
	if delim := strings.Index(link, ":["); delim > 1 {
		factory, ok := fdTypeFactories[link[:delim]]
		if ok {
			return factory(fd, link)
		}
	}
	// Fall back onto the plain file system path fd type.
	return NewPathFd(fd, link)
}

// fdConstructor returns a new FileDescriptor for the specified fd number and
// link "destination". These destinations can be "ordinary" file paths, or in
// the formats "type:[inode]" and "anon_inode:<type>".
type fdConstructor func(fd int, link string) (FileDescriptor, error)

// fdTypeFactories maps "type:[inode]" fds to their corresponding type factory.
var fdTypeFactories = map[string]fdConstructor{
	"pipe":   NewPipeFd,
	"socket": NewSocketFd,
}

// filedesc describes the information common to all "types" of file descriptor.
type filedesc struct {
	fd    int   // file descriptor number
	flags Flags // access mode and status flags as used by open(2)
	mntId int   // mount ID; might be present in /proc/self/mountinfo
}

// newFiledesc returns a new filedesc for a specific fd (number), initialized
// with information gathered from the procfs filesystem mounted on /proc.
func newFiledesc(fd int) (filedesc, error) {
	// for some types of file descriptors, we might face a rather lengthy
	// fdinfo, so we don't try to swallow it completely, but only read up to the
	// point we need. As it seems, the generic bits of information always come
	// first.
	file, err := os.Open(fmt.Sprintf("/proc/self/fdinfo/%d", fd))
	if err != nil {
		return filedesc{}, err
	}
	defer file.Close()
	return fdFromReader(fd, file)
}

// fdFromReader returns a filedesc initialized from the fdinfo read from the
// specified reader.
func fdFromReader(fd int, r io.Reader) (filedesc, error) {
	f := filedesc{fd: fd}
	scanner := bufio.NewScanner(r)
	complete := false
scanning:
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "pos:"):
			// ...go on...
		case strings.HasPrefix(line, "flags:"):
			flags, err := strconv.ParseUint(strings.Trim(line[6:], "\t "), 8, bits.UintSize)
			if err != nil {
				return filedesc{}, err
			}
			if flags > math.MaxInt {
				return filedesc{}, fmt.Errorf("fdFromReader: flags outside range: %d", flags)
			}
			f.flags = Flags(flags)
		case strings.HasPrefix(line, "mnt_id:"):
			mntId, err := strconv.ParseInt(strings.Trim(line[7:], "\t "), 10, bits.UintSize)
			if err != nil {
				return filedesc{}, err
			}
			if mntId <= 0 || mntId > math.MaxInt {
				return filedesc{}, fmt.Errorf("fdFromReader: mnt_id outside range: %d", mntId)
			}
			f.mntId = int(mntId)
			complete = true
			break scanning
		}
	}
	if err := scanner.Err(); err != nil {
		return filedesc{}, err
	}
	if !complete {
		return filedesc{}, errors.New("fdFromReader: incomplete fdinfo data")
	}
	return f, nil
}

// Fd returns the fd number.
func (fd filedesc) Fd() int { return fd.fd }

// Flags returns the file descriptor's flags, consisting of the access mode and
// status flags as used by open(2).
func (fd filedesc) Flags() Flags { return fd.flags }

// MountId returns the ID of the mount this fd is on.
func (fd filedesc) MountId() int { return fd.mntId }

// Description returns a pretty formatted textual description of the common
// elements for each fd (filedesc): the fd number and the (current) flags. For
// better use, the flags are shown with their symbolic names, where possible.
func (fd filedesc) Description(indentation uint) string {
	flags := strings.Join(fd.flags.Names(), ",") // sic! bang them names together without space
	if flags != "" {
		flags = " (" + flags + ")"
	}
	return Indentation(indentation) +
		fmt.Sprintf("fd %d, flags 0x%x%s", fd.fd, fd.flags, flags)
}

// Equal returns true if other is a filedesc with the same fd number and mount
// ID, but ignores the flags. This caters for before/after situations where the
// fd flags might have changed in between.
func (fd filedesc) Equal(other *filedesc) bool {
	return fd.fd == other.fd && fd.mntId == other.mntId
}
