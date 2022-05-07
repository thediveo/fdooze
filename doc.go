/*

Package fdooze complements the Gingko/Gomega testing and matchers framework with
matchers for file descriptor leakage detection.

Please note that due to technical restrictions this (experimental) package is
available only for Linux.

Basic Usage

In your project (with a go.mod) run "go get github.com/thediveo/fdooze" to get
and install the latest stable release.

A typical usage in your tests and using Ginkgo then is:

    BeforeEach(func() {
        goodfds := Filedescriptors()
        DeferCleanup(func() {
            Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
        })
    })

This takes a snapshot of "good" file descriptors before each test and then after
each test it checks to see if there are any leftover file descriptors that
weren't already in use before a test. The fdooze package does not blindly just
compare fd numbers, but takes as much additional detail information as possible
into account: like file paths, socket domains, types, protocols and addresses,
et cetera.

On finding leaked file descriptors, fdooze dumps these leaked fds in the failure
message of the HaveLeakedFds matcher. For instance:

    Expected not to leak 1 file descriptors:
        fd 7, flags 0xa0000 (O_RDONLY,O_CLOEXEC)
            path: "/home/leaky/module/oozing_test.go"

For other types of file descriptors, such as pipes and sockets, several details
will differ: instead of a path, other parameters will be shown, like pipe inode
numbers or socket addresses. Due to the limitations of the existing fd discovery
API, it is not possible to see where the file descriptor was opened (which might
be deep inside some 3rd party package anyway).

Expect or Eventually

In case you are already familiar with Gomega's gleak goroutine leak detection
package, then please note that typical fdooze usage doesn't require Eventually,
so Expect is fine most of the time. However, in situations where goroutines open
file descriptors it might be a good idea to first wait for goroutines to
terminate and not leak and only then test for any file descriptor leaks.

When using Eventually() make sure to pass the Filedescriptors function itself to
it, not the result of calling Filedescriptors.

    // Correct
    Eventually(Filedescriptors).ShouldNot(HaveLeakedFds(...))

    // WRONG WRONG WRONG
    Eventually(Filedescriptors()).ShouldNot(HaveLeakedFds(...))

*/
package fdooze
