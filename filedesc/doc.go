/*

Package filedesc implements file descriptor (fd) discovery beyond plain fd
numbers. The discovery includes further details, such as the fd flags, socket
local and peer addresses, et cetera.

The filedesc package is not designed as a (generalized) communication diagnosis
package. Instead, it is especially designed to give useful fd details in order
to help identify the origins of leaked file descriptors.

Thus, emphasis is put on providing clear and helpful "dumps" of fd properties;
for instance, by outputting symbolic constant names where known for address
families, socket types, domain-specific protocols and not just obscure octal or
hex numbers (sorry, PDP-11 & co). FileDescriptor.Details returns a detailed
textual representation of a FileDescriptor, in some respect resembling Gomega's
format.Object. But in contrast, FileDescriptor.Details isn't a generic struct
type dump, but instead is hand-crafted to aid quickly understanding a file
descriptor's properties, using symbolic constant names wherever applicable.

In order to better support use case-specific custom matchers, the different
struct types implementing the FileDescriptor interface provide public accessors
(value receivers) to their fd type-specific properties. These can be easily,
erm, accessed using Gomega's HaveField matcher (after checking for the correct
fd type to avoid HaveField errors for non-existing fields and receivers).

*/
package filedesc
