//go:build linux

package session

import (
	"errors"

	"github.com/onsi/gomega/gexec"
	"github.com/thediveo/fdooze/filedesc"
)

// Filedescriptors returns the list of currently open file descriptors for the
// process specified by session.
func FiledescriptorsFor(session *gexec.Session) ([]filedesc.FileDescriptor, error) {
	if session == nil || session.Command == nil {
		return nil, errors.New("invalid session or session command")
	}
	if session.Command.Process == nil || session.Command.Process.Pid == -1 {
		return nil, errors.New("invalid session without process")
	}
	if session.Command.ProcessState != nil {
		return nil, errors.New("session has already ended")
	}
	return filedesc.ProcessFiledescriptors(session.Command.Process.Pid)
}
