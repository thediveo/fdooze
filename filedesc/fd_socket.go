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
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

// SocketFd implements the FileDescriptor interface for an fd representing a
// socket from various domains, not least the unix and various networking
// domains.
type SocketFd struct {
	filedesc
	ino       uint64       // socket's inode number.
	domain    SocketDomain // the socket's address/protocol family ("domain")
	typ       SocketType   // type of socket, that is, type parameter to socket()
	protocol  SocketProtocol
	local     Sockaddr
	peer      Sockaddr
	listening bool
}

// NewSocketFd returns a new FileDescriptor for a pipe fd. If there is any
// problem with determining the plethora of socket parameters and binding, then
// a nil FileDescriptor is returned instead with the error indication.
func NewSocketFd(fdNo int, base string, linkDest string) (FileDescriptor, error) {
	inoArg := strings.TrimSuffix(strings.TrimPrefix(linkDest, "socket:["), "]")
	ino, err := strconv.ParseUint(inoArg, 10, 64)
	if err != nil {
		return nil, err
	}
	filedesc, err := newFiledesc(fdNo, base)
	if err != nil {
		return nil, err
	}

	// turn the fdNo into a useable fd (number): for one of our own fd numbers
	// we simply can use it as-is, as we're the same process; but if it is from
	// a different process, we first need to clone the other process's fd into
	// our own fd.
	useableFd := fdNo
	if !strings.HasPrefix(base, "/proc/self/") {
		fields := strings.SplitN(base, "/", 4)
		if len(fields) < 4 {
			return nil, errors.New("invalid fd base \"" + base + "\"")
		}
		pid, err := strconv.Atoi(fields[2])
		if err != nil {
			return nil, err
		}
		pidFd, err := unix.PidfdOpen(pid, 0)
		if err != nil {
			return nil, err
		}
		defer unix.Close(pidFd)
		useableFd, err /* no ":=" */ = unix.PidfdGetfd(pidFd, fdNo, 0)
		if err != nil {
			return nil, err
		}
		defer unix.Close(useableFd)
	}

	// Get the parameters from the call to socket(domain, type, protocol); we
	// need to successfully retrieve these.
	domain, err := getsockoptInt(useableFd, unix.SOL_SOCKET, unix.SO_DOMAIN)
	if err != nil {
		return nil, err
	}
	typ, err := getsockoptInt(useableFd, unix.SOL_SOCKET, unix.SO_TYPE)
	if err != nil {
		return nil, err
	}
	protocol, err := getsockoptInt(useableFd, unix.SOL_SOCKET, unix.SO_PROTOCOL)
	if err != nil {
		return nil, err
	}

	// ...oh, and check if it is a listening socket. But this time we accept
	// failure as only few socket types might champion the concept of
	// "listening".
	listening, _ := getsockoptInt(useableFd, unix.SOL_SOCKET, unix.SO_ACCEPTCONN)

	// Now get the local and remote addresses, erm, "names"; again, these might
	// not be available for some socket families, sadly.
	local, _ := getsockname(useableFd)
	peer, _ := getpeername(useableFd)

	return &SocketFd{
		filedesc:  filedesc,
		ino:       ino,
		domain:    SocketDomain(domain),
		typ:       SocketType(typ),
		protocol:  SocketProtocol(protocol),
		local:     Sockaddr{local},
		peer:      Sockaddr{peer},
		listening: listening > 0,
	}, nil
}

// Ino returns the socket's inode number.
func (s SocketFd) Ino() uint64 { return s.ino }

// Domain returns the socket's communication domain that selects the address
// family used.
func (s SocketFd) Domain() int { return int(s.domain) }

// Type returns the socket's type, which specifies the communication semantics
// (such as byte stream, datagrams, et cetera).
func (s SocketFd) Type() int { return int(s.typ) }

// Protocol returns the socket's protocol, specific within the socket's domain.
func (s SocketFd) Protocol() int { return int(s.protocol) }

// Listening returns true if the socket is in listening mode.
func (s SocketFd) Listening() bool { return s.listening }

// Description returns a pretty formatted textual description of this socket
// file descriptor.
func (s SocketFd) Description(indentation uint) string {
	newindent := "\n" + Indentation(indentation+1)
	var buff strings.Builder

	buff.WriteString(s.filedesc.Description(indentation))

	buff.WriteString(newindent)
	if s.listening {
		buff.WriteString("listening ")
	}
	buff.WriteString(fmt.Sprintf("socket(%s, %s, %s), ino %d",
		s.domain.String(), s.typ.String(), s.protocol.String(s.domain), s.ino))

	buff.WriteString(newindent)
	buff.WriteString(fmt.Sprintf("local %q", s.local.String()))

	if s.peer.Sockaddr != nil {
		buff.WriteString(newindent)
		buff.WriteString(fmt.Sprintf("peer %q", s.peer.String()))
	}

	return buff.String()
}

// Name returns the socket's name (that is, address) in textual format. Call the
// Addr receiver instead in order to get the socket's unix.Sockaddr.
func (s SocketFd) Name() string { return s.local.String() }

// Addr returns the socket's name (that is, address). To access the individual
// family-specific address elements, cast the returned interface to a pointer to
// the correct underlying socket address type, such as *unix.SockaddrUnix or
// *unix.SockaddrInet, et cetera.
func (s SocketFd) Addr() unix.Sockaddr { return s.local.Sockaddr }

// Peer returns the socket peer's name (that is, address) in textual format,
// returning "" if the socket isn't connected. Call the PeerAddr receiver
// instead in order to get the socket peer's unix.Sockaddr.
func (s SocketFd) Peer() string { return s.peer.String() }

// PeerAddr returns the socket peer's name (that is, address). To access the
// individual family-specific address elements, cast the returned interface to a
// pointer to the correct underlying socket address type, such as
// *unix.SockaddrUnix or *unix.SockaddrInet, et cetera.
func (s SocketFd) PeerAddr() unix.Sockaddr { return s.peer.Sockaddr }

// Equal returns true, if other is a pipeFd with the same fd number and mount
// ID, as well as the same inode number.
func (s SocketFd) Equal(other FileDescriptor) bool {
	o, ok := other.(*SocketFd)
	if !ok {
		return false
	}
	return s.filedesc.Equal(&o.filedesc) &&
		s.ino == o.ino &&
		s.domain == o.domain && s.typ == o.typ && s.protocol == o.protocol &&
		s.listening == o.listening &&
		reflect.DeepEqual(s.local, o.local) && reflect.DeepEqual(s.peer, o.peer)
}
