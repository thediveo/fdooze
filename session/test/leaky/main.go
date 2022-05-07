package main

import (
	"bufio"
	"fmt"
	"os"
)

// Opening for the first time a file triggers creation of additional file
// descriptors that are used by Go's stdlib I/O multiplexing on Linux. This
// so-called epoll-based "netpoller" is also responsible for file I/O, despite
// its slightly misleading name. The epoll-based netpoller on Linux opens an
// anonymous "epoll" inode fd, as well as a (non-blocking) pipe for internal
// purposes in order to be able to break out of a blocking epoll_wait syscall.
//
// We simply immediately close the file we just opened so that we don't leak the
// "priming" fd, yet the I/O multiplexing fds will be left open.
func primeIO() {
	f, err := os.Open("/dev/null")
	if err != nil {
		panic(err)
	}
	defer f.Close()
}

func main() {
	primeIO()
	r := bufio.NewReader(os.Stdin)
	fmt.Println("READY")
	_, _ = r.ReadString('\n')

	f, err := os.Open("./test/leaky/main.go")
	if err != nil {
		panic(err)
	}
	fmt.Println("LEAK")
	_, _ = r.ReadString('\n')

	f.Close()
	fmt.Println("PLUMBED")
	_, _ = r.ReadString('\n')
}
