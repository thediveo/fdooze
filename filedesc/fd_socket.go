//go:build linux

package filedesc

import (
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
	stype     SocketType   // type of socket, that is, type parameter to socket()
	protocol  SocketProtocol
	local     Sockaddr
	peer      Sockaddr
	listening bool
}

// NewSocketFd returns a new FileDescriptor for a pipe fd. If there is any
// problem with determining the plethora of socket parameters and binding, then
// a nil FileDescriptor is returned instead with the error indication.
func NewSocketFd(fd int, link string) (FileDescriptor, error) {
	inoArg := strings.TrimSuffix(strings.TrimPrefix(link, "socket:["), "]")
	ino, err := strconv.ParseUint(inoArg, 10, 64)
	if err != nil {
		return nil, err
	}
	filedesc, err := newFiledesc(fd)
	if err != nil {
		return nil, err
	}

	// Get the parameters from the call to socket(domain, type, protocol); we
	// need to successfully retrieve these.
	domain, err := getsockoptInt(fd, unix.SOL_SOCKET, unix.SO_DOMAIN)
	if err != nil {
		return nil, err
	}
	stype, err := getsockoptInt(fd, unix.SOL_SOCKET, unix.SO_TYPE)
	if err != nil {
		return nil, err
	}
	protocol, err := getsockoptInt(fd, unix.SOL_SOCKET, unix.SO_PROTOCOL)
	if err != nil {
		return nil, err
	}
	// Oh, and check if it is a listening socket...
	listening, err := getsockoptInt(fd, unix.SOL_SOCKET, unix.SO_ACCEPTCONN)
	if err != nil {
		return nil, err
	}

	// Now get the local and remote addresses, erm, "names"...
	local, err := getsockname(fd)
	if err != nil {
		return nil, err
	}
	// ...please note that getpeername(2) will fail if the socket isn't
	// connected and then return ENOTCONN. This is expected, but all other error
	// results are considered to be, well, errors.
	peer, err := getpeername(fd)
	if err != nil && err != unix.ENOTCONN {
		return nil, err
	}

	return &SocketFd{
		filedesc:  filedesc,
		ino:       ino,
		domain:    SocketDomain(domain),
		stype:     SocketType(stype),
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
func (s SocketFd) Type() int { return int(s.stype) }

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
		s.domain.String(), s.stype.String(), s.protocol.String(s.domain), s.ino))

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
		s.domain == o.domain && s.stype == o.stype && s.protocol == o.protocol &&
		s.listening == o.listening &&
		reflect.DeepEqual(s.local, o.local) && reflect.DeepEqual(s.peer, o.peer)
}
