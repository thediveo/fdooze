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
	"fmt"
	"strings"

	"golang.org/x/sys/unix"
)

// SocketDomain specifies a socket communication domain and implements a
// Stringer returns the symbolic constant name for the domain value. See also:
// https://man7.org/linux/man-pages/man2/socket.2.html
type SocketDomain int

// socketDomainNames maps the address family/domain constants to their
// corresponding textual representations.
var socketDomainNames = map[int]string{
	unix.AF_ALG:        "AF_ALG",
	unix.AF_APPLETALK:  "AF_APPLETALK",
	unix.AF_ASH:        "AF_ASH",
	unix.AF_ATMPVC:     "AF_ATMPVC",
	unix.AF_ATMSVC:     "AF_ATMSVC",
	unix.AF_AX25:       "AF_AX25",
	unix.AF_BLUETOOTH:  "AF_BLUETOOTH",
	unix.AF_BRIDGE:     "AF_BRIDGE",
	unix.AF_CAIF:       "AF_CAIF",
	unix.AF_CAN:        "AF_CAN",
	unix.AF_DECnet:     "AF_DECnet",
	unix.AF_ECONET:     "AF_ECONET",
	unix.AF_IB:         "AF_IB",
	unix.AF_IEEE802154: "AF_IEEE802154",
	unix.AF_INET:       "AF_INET",
	unix.AF_INET6:      "AF_INET6",
	unix.AF_IPX:        "AF_IPX",
	unix.AF_IRDA:       "AF_IRDA",
	unix.AF_ISDN:       "AF_ISDN",
	unix.AF_IUCV:       "AF_IUCV",
	unix.AF_KCM:        "AF_KCM",
	unix.AF_KEY:        "AF_KEY",
	unix.AF_LLC:        "AF_LLC",
	unix.AF_MAX:        "AF_MAX",
	unix.AF_MPLS:       "AF_MPLS",
	unix.AF_NETBEUI:    "AF_NETBEUI",
	unix.AF_NETLINK:    "AF_NETLINK",
	unix.AF_NETROM:     "AF_NETROM",
	unix.AF_NFC:        "AF_NFC",
	unix.AF_PACKET:     "AF_PACKET",
	unix.AF_PHONET:     "AF_PHONET",
	unix.AF_PPPOX:      "AF_PPPOX",
	unix.AF_QIPCRTR:    "AF_QIPCRTR",
	unix.AF_RDS:        "AF_RDS",
	unix.AF_ROSE:       "AF_ROSE",
	unix.AF_RXRPC:      "AF_RXRPC",
	unix.AF_SECURITY:   "AF_SECURITY",
	unix.AF_SMC:        "AF_SMC",
	unix.AF_SNA:        "AF_SNA",
	unix.AF_TIPC:       "AF_TIPC",
	unix.AF_UNIX:       "AF_UNIX",
	unix.AF_UNSPEC:     "AF_UNSPEC",
	unix.AF_VSOCK:      "AF_VSOCK",
	unix.AF_WANPIPE:    "AF_WANPIPE",
	unix.AF_X25:        "AF_X25",
	unix.AF_XDP:        "AF_XDP",
}

// String returns a textual representation for a given SocketDomain value.
func (d SocketDomain) String() string {
	n, ok := socketDomainNames[int(d)]
	if !ok {
		return fmt.Sprintf("domain %d", int(d))
	}
	return n
}

// SocketType indicates the communication semantics of socket and additionally
// returns a textual representation. The term “type” is historically founded in
// the [socket(2)] call parameter names.
//
// [socket(2)]: https://man7.org/linux/man-pages/man2/socket.2.html
type SocketType int

// socketTypeNames maps the socket type constants to their corresponding textual
// representations.
var socketTypeNames = map[int]string{
	unix.SOCK_DGRAM:     "SOCK_DGRAM",
	unix.SOCK_PACKET:    "SOCK_PACKET",
	unix.SOCK_RAW:       "SOCK_RAW",
	unix.SOCK_RDM:       "SOCK_RDM",
	unix.SOCK_SEQPACKET: "SOCK_SEQPACKET",
	unix.SOCK_STREAM:    "SOCK_STREAM",
}

// String returns a textual representation for a given SocketType value.
func (t SocketType) String() string {
	n, ok := socketTypeNames[int(t)]
	if !ok {
		return fmt.Sprintf("type %d", int(t))
	}
	return n
}

// SocketProtocol specifies a particular communication [protocol(5)]. A
// SocketProtocol always must be interpreted in the context of a specific
// [SocketDomain].
//
// [protocol(5)]: https://man7.org/linux/man-pages/man5/protocols.5.html
type SocketProtocol int

var socketIPNames = map[int]string{
	unix.IPPROTO_AH:       "IPPROTO_AH",
	unix.IPPROTO_BEETPH:   "IPPROTO_BEETPH",
	unix.IPPROTO_COMP:     "IPPROTO_COMP",
	unix.IPPROTO_DCCP:     "IPPROTO_DCCP",
	unix.IPPROTO_DSTOPTS:  "IPPROTO_DSTOPTS",
	unix.IPPROTO_EGP:      "IPPROTO_EGP",
	unix.IPPROTO_ENCAP:    "IPPROTO_ENCAP",
	unix.IPPROTO_ESP:      "IPPROTO_ESP",
	unix.IPPROTO_ETHERNET: "IPPROTO_ETHERNET",
	unix.IPPROTO_FRAGMENT: "IPPROTO_FRAGMENT",
	unix.IPPROTO_GRE:      "IPPROTO_GRE",
	unix.IPPROTO_ICMP:     "IPPROTO_ICMP",
	unix.IPPROTO_ICMPV6:   "IPPROTO_ICMPV6",
	unix.IPPROTO_IDP:      "IPPROTO_IDP",
	unix.IPPROTO_IGMP:     "IPPROTO_IGMP",
	unix.IPPROTO_IP:       "IPPROTO_IP",
	unix.IPPROTO_IPIP:     "IPPROTO_IPIP",
	unix.IPPROTO_IPV6:     "IPPROTO_IPV6",
	unix.IPPROTO_L2TP:     "IPPROTO_L2TP",
	unix.IPPROTO_MH:       "IPPROTO_MH",
	unix.IPPROTO_MPLS:     "IPPROTO_MPLS",
	unix.IPPROTO_MPTCP:    "IPPROTO_MPTCP",
	unix.IPPROTO_MTP:      "IPPROTO_MTP",
	unix.IPPROTO_NONE:     "IPPROTO_NONE",
	unix.IPPROTO_PIM:      "IPPROTO_PIM",
	unix.IPPROTO_PUP:      "IPPROTO_PUP",
	unix.IPPROTO_RAW:      "IPPROTO_RAW",
	unix.IPPROTO_ROUTING:  "IPPROTO_ROUTING",
	unix.IPPROTO_RSVP:     "IPPROTO_RSVP",
	unix.IPPROTO_SCTP:     "IPPROTO_SCTP",
	unix.IPPROTO_TCP:      "IPPROTO_TCP",
	unix.IPPROTO_TP:       "IPPROTO_TP",
	unix.IPPROTO_UDP:      "IPPROTO_UDP",
	unix.IPPROTO_UDPLITE:  "IPPROTO_UDPLITE",
}

var socketNlNames = map[int]string{
	unix.NETLINK_ROUTE:          "NETLINK_ROUTE",
	unix.NETLINK_UNUSED:         "NETLINK_UNUSED",
	unix.NETLINK_USERSOCK:       "NETLINK_USERSOCK",
	unix.NETLINK_FIREWALL:       "NETLINK_FIREWALL",
	unix.NETLINK_SOCK_DIAG:      "NETLINK_SOCK_DIAG",
	unix.NETLINK_NFLOG:          "NETLINK_NFLOG",
	unix.NETLINK_XFRM:           "NETLINK_XFRM",
	unix.NETLINK_SELINUX:        "NETLINK_SELINUX",
	unix.NETLINK_ISCSI:          "NETLINK_ISCSI",
	unix.NETLINK_AUDIT:          "NETLINK_AUDIT",
	unix.NETLINK_FIB_LOOKUP:     "NETLINK_FIB_LOOKUP",
	unix.NETLINK_CONNECTOR:      "NETLINK_CONNECTOR",
	unix.NETLINK_NETFILTER:      "NETLINK_NETFILTER",
	unix.NETLINK_IP6_FW:         "NETLINK_IP6_FW",
	unix.NETLINK_DNRTMSG:        "NETLINK_DNRTMSG",
	unix.NETLINK_KOBJECT_UEVENT: "NETLINK_KOBJECT_UEVENT",
	unix.NETLINK_GENERIC:        "NETLINK_GENERIC",
	unix.NETLINK_SCSITRANSPORT:  "NETLINK_SCSITRANSPORT",
	unix.NETLINK_ECRYPTFS:       "NETLINK_ECRYPTFS",
	unix.NETLINK_RDMA:           "NETLINK_RDMA",
	unix.NETLINK_CRYPTO:         "NETLINK_CRYPTO",
	unix.NETLINK_SMC:            "NETLINK_SMC",
}

// String returns the textual representation corresponding to a socket protocol
// from the AF_INET and AF_INET6 domains. For other domains, it returns a
// textual description based on the protocol number. Please note that
// SocketProtocol on purpose does not implement the Stringer interface, as
// protocols are only defined in the contexts of specific domains. A socket
// protocol without known the domain is thus useless and ambiguous. In
// consequence, String strictly requires a domain parameter.
func (p SocketProtocol) String(domain SocketDomain) string {
	switch domain {
	case unix.AF_INET, unix.AF_INET6:
		if pname, ok := socketIPNames[int(p)]; ok {
			return pname
		}
	case unix.AF_NETLINK:
		if nlname, ok := socketNlNames[int(p)]; ok {
			return nlname
		}
	}
	return fmt.Sprintf("protocol %d", int(p))
}

// hexString returns the hexadecimal encoding (using uppercase hex digits A-F)
// of src, separating the every two digits using separator.
func hexString(src []byte, separator rune) string {
	var hex strings.Builder
	for idx, b := range src {
		if idx > 0 && separator != 0 {
			hex.WriteRune(separator)
		}
		hex.WriteByte(hexDigits[b>>4])
		hex.WriteByte(hexDigits[b&0x0f])
	}
	return hex.String()
}

// hexDigits contains all hex digits for easy nibble conversion.
const hexDigits = "0123456789ABCDEF"
