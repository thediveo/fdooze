//go:build linux

package filedesc

import "golang.org/x/sys/unix"

// To 100+% coverage and beyond...!!!

// So, who is mocking whom?

var getsockoptInt func(int, int, int) (int, error) = unix.GetsockoptInt
var getsockname func(int) (unix.Sockaddr, error) = unix.Getsockname
var getpeername func(int) (unix.Sockaddr, error) = unix.Getpeername
