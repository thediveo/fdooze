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

package session

import (
	"errors"
	"io/fs"

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
	// We can only try now to get the file descriptors for the process belonging
	// to the session. If that fails and the reason is that we couldn't read the
	// process's file descriptor directory, then return a more meaningful error
	// to the caller that the session already has terminated.
	fds, err := filedesc.ProcessFiledescriptors(session.Command.Process.Pid)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, errors.New("session has already ended")
	}
	return fds, err
}
