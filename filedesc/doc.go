/*
Package filedesc implements file descriptor (“fd”) discovery beyond just the
plain fd numbers. The discovery includes further details, such as the fd flags,
socket local and peer addresses, et cetera.

Please note that this package is not designed as a (generalized) communication
diagnosis package. Instead, it is especially designed to give useful fd details
in order to help identify the origins (“source sites”) of leaked file descriptors.

Thus, emphasis is put on providing clear and helpful “dumps” of fd properties;
for instance, by outputting symbolic constant names where known for address
families, socket types, domain-specific protocols – and not obscure octal or hex
numbers (sorry, PDP-11 & co).

[FileDescriptor.Description] returns a detailed textual representation of a
FileDescriptor, in some respect resembling Gomega's [format.Object]. But in
contrast to what Gomega has on offer, FileDescriptor.Details isn't a generic
struct type dump, but instead is hand-crafted to aid quickly understanding a
file descriptor's properties, using symbolic constant names wherever applicable.

In order to better support use case-specific custom matchers, the different
struct types implementing the FileDescriptor interface provide public accessors
(value receivers) to their fd type-specific properties. These properties then
can be easily accessed using Gomega's [HaveField] matcher (preferably only after
first checking for the correct fd type to avoid HaveField errors for
non-existing fields and receivers using the [HaveExistingField] matcher).

# Usage

The most common use case probably is to simply discover the list of open file
descriptor (details) by either calling [Filedescriptors] or
[ProcessFiledescriptors]. While [Filedescriptors] returns the open file
descriptor details of the caller's process, [ProcessFiledescriptors] returns the
open file descriptor details of the process with the specified PID. For this,
the process must be either belonging to the same user or the caller must possess
sufficient capabilities to access arbitrary processes.

[HaveField]: https://onsi.github.io/gomega/#havefieldfield-interface-value-interface
[HaveExistingField]: https://onsi.github.io/gomega/#havefieldfield-interface-value-interface
*/
package filedesc
