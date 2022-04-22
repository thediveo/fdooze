<!-- markdownlint-disable-next-line MD022 -->
# `fdooze`
<img title="Goigi the gopher" align="right" width="150" src="images/goigi-small.png">

[![PkgGoDev](https://img.shields.io/badge/-reference-blue?logo=go&logoColor=white&labelColor=505050)](https://pkg.go.dev/github.com/thediveo/fdooze)
[![GitHub](https://img.shields.io/github/license/thediveo/fdooze)](https://img.shields.io/github/license/thediveo/fdooze)
[![Go Report Card](https://goreportcard.com/badge/github.com/thediveo/fdooze)](https://goreportcard.com/report/github.com/thediveo/fdooze)

`fdooze` complements [Gomega](https://github.com/onsi/gomega) with tests to
detect leaked ("oozed") file descriptors.

> **Note:** `fdooze` is available on **Linux only**, as discovering file
> descriptor information requires using highly system-specific APIs and the
> descriptor information varies across different systems (if available at all).

## Basic Usage

In your project (with a `go.mod`) run `go get github.com/thediveo/fdooze` to get
and install the latest stable release.

A typical usage in your tests then is (using
[Ginkgo](https://github.com/onsi/ginkgo)):

```go
BeforeEach(func() {
    goodfds := Filedescriptors()
    DeferCleanup(func() {
        Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
    })
})
```

This takes a snapshot of "good" file descriptors before each test and then after
each test it checks to see if there are any leftover file descriptors that
weren't already in use before a test. `fdooze` does not blindly just compare fd
numbers, but takes as much additional detail information as possible into
account: like file paths, socket domains, types, protocols and addresses, et
cetera.

On finding leaked file descriptors, `fdooze` dumps these leaked fds in the
failure message of the `HaveLeakedFds` matcher. For instance:

```
Expected not to leak 1 file descriptors:
    fd 7, flags 0xa0000 (O_RDONLY,O_CLOEXEC)
        path: "/home/leaky/module/oozing_test.go"
```

For other types of file descriptors, such as pipes and sockets, several details
will differ: instead of a path, other parameters will be shown, like pipe inode
numbers or socket addresses. Due to the limitations of the existing fd discovery
API, it is not possible to see _where_ the file descriptor was opened (which
might be deep inside some 3rd party package anyway).

## `Expect` or `Eventually`?

In case you are already familiar with Gomega's
[gleak](https://onsi.github.io/gomega/#codegleakcode-finding-leaked-goroutines)
goroutine leak detection package, then please note that typical `fdooze` usage
doesn't require `Eventually`, so `Expect` is fine most of the time. However, in
situations where goroutines open file descriptors it might be a good idea to
first wait for goroutines to terminate and not leak and only then test for any
file descriptor leaks.

## Go Version Support

`fdooze` supports versions of Go that are noted by the Go release policy, that
is, major versions _N_ and _N_-1 (where _N_ is the current major version).

## Goigi the Gopher

Goigi the gopher mascot clearly has been inspired by the Go gopher art work of
[Renee French](http://reneefrench.blogspot.com/). It seems as if Goigi has some
issues with plumbing file descriptors properly.

## ⚖️ Copyright and License

`fdooze` is Copyright 2022 Harald Albrecht, and licensed under the Apache
License, Version 2.0.
