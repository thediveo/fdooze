/*

Package session implements retrieving the open file descriptors from a Gomega
gexec.Session (Linux only). This allows checking processes launched by a test
suite for file descriptor leaks, subject to normal process access control.

It is recommended to dot-import the session package, as this keeps the "session"
identifier free to be used by test writers as they see fit.

    session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
    Expect(err).NotTo(HaveOccurred())

    // Optional, please see note below.
    client.DoWarmupAPIThing()

    // When using Eventually, make sure to pass the function, not its result!
    sessionFds := func ([]FileDescriptor, error) {
        return FiledescriptorsFor(session)
    }

    goodfds := sessionFds()

    client.DoSomeAPIThing()
    Expect(session).Should(gbytes.Say("I did the thing"))

    Eventually(sessionFds).ShouldNot(HaveLeakedFds(goodfds))
    Eventually(session.Interrupt()).Should(gexec.Exit(0))

Launched Go Processes False Positives

In case the launched process is implemented in Go, fd leak tests need to be
carefully designed as to not fail with false positive fd leaks caused by Go's
netpoll runtime (for instance, see https://morsmachine.dk/netpoller for more
background information).

For instance, when opening a file or network socket for the first time, Go's
runtime creates an internal epoll fd as well as a non-blocking pipe fd for use
in its internal asynchronous I/O handling.

Unfortunately, it is not possible to easily filter out the file descriptors
belonging to the Go runtime netpoller: fds in general don't record who created
them and for what purpose. An epoll fd might be used in an application itself
and thus quite often be ambigous. Also, the exact fd number will depend on a Go
application highly specific initialization process.

It is thus mandatory to take a "reference" snapshot of baseline fds only after
the launched process has opened its first file or network socket. In case of
network-facing services this will be when the listening transport port has
become available.

*/
package session
