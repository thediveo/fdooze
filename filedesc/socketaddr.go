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
	"net"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

// Sockaddr wraps unix.Sockaddr (the interface) in order to enhance it by
// implementing the Stringer interface (on this wrapper). The wrapped socket
// address is allowed to be nil, as this is nil-inclusive software.
type Sockaddr struct {
	unix.Sockaddr // any supported socket address (interface)
}

// String returns a textual (single-line) representation of the wrapped kind of
// unix.Sockaddr. For instance, "/path/name" for a unix domain socket address,
// "192.0.0.1:1234" for an IPv4 socket address, "[fe80::dead:beef]:1234" for an
// IPv6 socket address, et cetera. When wrapping a nil unix.Sockaddr, String
// returns an empty string "".
func (a Sockaddr) String() string {
	if a.Sockaddr == nil {
		return ""
	}
	switch sockaddr := a.Sockaddr.(type) {
	case *unix.SockaddrInet4:
		return ipv4AddrFormat(sockaddr)
	case *unix.SockaddrInet6:
		return ipv6AddrFormat(sockaddr)
	case *unix.SockaddrLinklayer:
		return linklayerAddrFormat(sockaddr)
	case *unix.SockaddrNetlink:
		return netlinkAddrString(sockaddr)
	case *unix.SockaddrUnix:
		// https://man7.org/linux/man-pages/man7/unix.7.html#DESCRIPTION
		return sockaddr.Name
	case *unix.SockaddrVM:
		return vmAddrString(sockaddr)
	case *unix.SockaddrXDP:
		return xdpAddrString(sockaddr)
	}
	// fall back to the Go-syntax representation of the socket address value.
	return fmt.Sprintf("%#v", a.Sockaddr)
}

// ipv6AddrFormat returns the single-line textual representation of an IPv6
// socket address (which includes the port number, as well as optionally the
// zone ID if not zero).
//
// See also: https://man7.org/linux/man-pages/man7/ipv6.7.html#DESCRIPTION
func ipv6AddrFormat(sockaddr *unix.SockaddrInet6) string {
	ip := net.IP(sockaddr.Addr[:])
	// \o/ Google knows its way around IPv6! They actually understand what a
	// zone is, as opposed to all the broken sockaddr_in6 C header files out
	// there, that are totally befuddled when it comes to scopes and zones
	// (once out they couldn't fix it anymore, oh well).
	//
	// Says RFC 4007: the scope is already part of the addressing
	// architecture, and thus known beforehand. Zones are "runtime", not
	// known beforehand, and depend on system configuration. They typically
	// pop up only when dealing with link-local fe80:: addresses to identify
	// the link (well, interface) they are associated with.
	if sockaddr.ZoneId == 0 {
		return fmt.Sprintf("[%s]:%d", ip.String(), sockaddr.Port)
	}
	return fmt.Sprintf("[%s%%%d]:%d", ip.String(), sockaddr.ZoneId, sockaddr.Port)
}

// ipv4AddrFormat returns the single-line textual representation of an IPv4
// socket address (which includes the port number).
//
// See also: https://man7.org/linux/man-pages/man7/ip.7.html#DESCRIPTION
func ipv4AddrFormat(sockaddr *unix.SockaddrInet4) string {
	ip := net.IP(sockaddr.Addr[:])
	return fmt.Sprintf("%s:%d", ip.String(), sockaddr.Port)
}

// linklayerAddrFormat returns the single-line textual representation of a data
// link layer (L2) socket address. This is not to be confused with MAC
// addresses: the MAC address is included in L2 socket addresses, but not the
// whole story.
//
// See also: https://man7.org/linux/man-pages/man7/packet.7.html#DESCRIPTION
func linklayerAddrFormat(sockaddr *unix.SockaddrLinklayer) string {
	pkttypename := packetTypeNames[sockaddr.Pkttype]
	if pkttypename == "" {
		pkttypename = strconv.FormatUint(uint64(sockaddr.Pkttype), 10)
	}
	ethtypename := ethTypeNames[sockaddr.Protocol]
	if ethtypename == "" {
		ethtypename = fmt.Sprintf("0x%x", sockaddr.Protocol)
	}
	return fmt.Sprintf("%s (HW address type 0x%x)\nprotocol %s, interface index %d, packet type %s",
		hexString(sockaddr.Addr[:sockaddr.Halen], ':'), sockaddr.Hatype,
		ethtypename, sockaddr.Ifindex, pkttypename)
}

// netlinkAddrString returns the single-line textual representation of a netlink
// socket address.
//
// See also: https://man7.org/linux/man-pages/man7/netlink.7.html#DESCRIPTION
func netlinkAddrString(sockaddr *unix.SockaddrNetlink) string {
	var dest string
	if sockaddr.Pid == 0 {
		dest = "kernel"
	} else {
		dest = fmt.Sprintf("(p)id %d", sockaddr.Pid)
	}
	return fmt.Sprintf("%s, multicast groups mask 0x%x",
		dest, sockaddr.Groups)

}

// vmAddrString returns the single-line textual representation of a VM socket
// address.
//
// See also: https://man7.org/linux/man-pages/man7/vsock.7.html#DESCRIPTION
func vmAddrString(sockaddr *unix.SockaddrVM) string {
	var cid string
	switch sockaddr.CID {
	case unix.VMADDR_CID_HYPERVISOR:
		cid = "VMADDR_CID_HYPERVISOR"
	case unix.VMADDR_CID_LOCAL:
		cid = "VMADDR_CID_LOCAL"
	case unix.VMADDR_CID_HOST:
		cid = "VMADDR_CID_HOST"
	case unix.VMADDR_CID_ANY:
		cid = "VMADDR_CID_ANY"
	default:
		cid = fmt.Sprintf("CID %d", sockaddr.CID)
	}
	return fmt.Sprintf("port %d, %s, flags %d", sockaddr.Port, cid, sockaddr.Flags)
}

// xdpAddrString returns the single-line textual representation of an XDP socket
// address.
//
// See also: https://www.kernel.org/doc/html/v4.19/networking/af_xdp.html and:
// https://elixir.bootlin.com/linux/v5.17.3/source/include/uapi/linux/if_xdp.h#L32
func xdpAddrString(sockaddr *unix.SockaddrXDP) string {
	flags := xdpFlagNames(sockaddr.Flags)
	if flags != "" {
		flags = " (" + flags + ")"
	}
	return fmt.Sprintf("flags: 0x%x%s, ifindex: %d, queue ID: %d, shared umem fd: %d",
		sockaddr.Flags, flags,
		sockaddr.Ifindex, sockaddr.QueueID, sockaddr.SharedUmemFD)
}

// xdpFlagNames returns a textual representation of the set and known XDP socket
// flags. Unknown flags are ignored.
//
// See:
// https://elixir.bootlin.com/linux/latest/source/include/uapi/linux/if_xdp.h#L15
func xdpFlagNames(flags uint16) string {
	if flags == 0 {
		return "" // quickly bail out if no flags present at all.
	}
	sf := []string{}
	for bitno, name := range xdpFlags {
		if flags&(1<<bitno) != 0 {
			sf = append(sf, name)
		}
	}
	return strings.Join(sf, ",") // sic! no space
}

// xdpFlags names the AF_XDP socket address flags, with the first array element
// naming the flag for bit 0, the next array element naming flag bit 1, and so
// on.
var xdpFlags = [...]string{
	"XDP_SHARED_UMEM",
	"XDP_COPY",
	"XDP_ZEROCOPY",
	"XDP_USE_NEED_WAKEUP",
}

// packetTypeNames maps SockaddrLinklayer's packet types to their symbolic
// constant names.
var packetTypeNames = map[uint8]string{
	unix.PACKET_HOST:      "PACKET_HOST",
	unix.PACKET_BROADCAST: "PACKET_BROADCAST",
	unix.PACKET_MULTICAST: "PACKET_MULTICAST",
	unix.PACKET_OTHERHOST: "PACKET_OTHERHOST",
	unix.PACKET_OUTGOING:  "PACKET_OUTGOING",
}

// ethTypeNames maps Ethernet protocol numbers to their symbolic constant names.
var ethTypeNames = map[uint16]string{
	unix.ETH_P_1588:       "ETH_P_1588",
	unix.ETH_P_8021AD:     "ETH_P_8021AD",
	unix.ETH_P_8021AH:     "ETH_P_8021AH",
	unix.ETH_P_8021Q:      "ETH_P_8021Q",
	unix.ETH_P_80221:      "ETH_P_80221",
	unix.ETH_P_802_2:      "ETH_P_802_2",
	unix.ETH_P_802_3:      "ETH_P_802_3",
	unix.ETH_P_802_3_MIN:  "ETH_P_802_3_MIN",
	unix.ETH_P_802_EX1:    "ETH_P_802_EX1",
	unix.ETH_P_AARP:       "ETH_P_AARP",
	unix.ETH_P_AF_IUCV:    "ETH_P_AF_IUCV",
	unix.ETH_P_ALL:        "ETH_P_ALL",
	unix.ETH_P_AOE:        "ETH_P_AOE",
	unix.ETH_P_ARCNET:     "ETH_P_ARCNET",
	unix.ETH_P_ARP:        "ETH_P_ARP",
	unix.ETH_P_ATALK:      "ETH_P_ATALK",
	unix.ETH_P_ATMFATE:    "ETH_P_ATMFATE",
	unix.ETH_P_ATMMPOA:    "ETH_P_ATMMPOA",
	unix.ETH_P_AX25:       "ETH_P_AX25",
	unix.ETH_P_BATMAN:     "ETH_P_BATMAN",
	unix.ETH_P_BPQ:        "ETH_P_BPQ",
	unix.ETH_P_CAIF:       "ETH_P_CAIF",
	unix.ETH_P_CAN:        "ETH_P_CAN",
	unix.ETH_P_CANFD:      "ETH_P_CANFD",
	unix.ETH_P_CFM:        "ETH_P_CFM",
	unix.ETH_P_CONTROL:    "ETH_P_CONTROL",
	unix.ETH_P_CUST:       "ETH_P_CUST",
	unix.ETH_P_DDCMP:      "ETH_P_DDCMP",
	unix.ETH_P_DEC:        "ETH_P_DEC",
	unix.ETH_P_DIAG:       "ETH_P_DIAG",
	unix.ETH_P_DNA_DL:     "ETH_P_DNA_DL",
	unix.ETH_P_DNA_RC:     "ETH_P_DNA_RC",
	unix.ETH_P_DNA_RT:     "ETH_P_DNA_RT",
	unix.ETH_P_DSA:        "ETH_P_DSA",
	unix.ETH_P_DSA_8021Q:  "ETH_P_DSA_8021Q",
	unix.ETH_P_ECONET:     "ETH_P_ECONET",
	unix.ETH_P_EDSA:       "ETH_P_EDSA",
	unix.ETH_P_ERSPAN:     "ETH_P_ERSPAN",
	unix.ETH_P_ERSPAN2:    "ETH_P_ERSPAN2",
	unix.ETH_P_FCOE:       "ETH_P_FCOE",
	unix.ETH_P_FIP:        "ETH_P_FIP",
	unix.ETH_P_HDLC:       "ETH_P_HDLC",
	unix.ETH_P_HSR:        "ETH_P_HSR",
	unix.ETH_P_IBOE:       "ETH_P_IBOE",
	unix.ETH_P_IEEE802154: "ETH_P_IEEE802154",
	unix.ETH_P_IEEEPUP:    "ETH_P_IEEEPUP",
	unix.ETH_P_IEEEPUPAT:  "ETH_P_IEEEPUPAT",
	unix.ETH_P_IFE:        "ETH_P_IFE",
	unix.ETH_P_IP:         "ETH_P_IP",
	unix.ETH_P_IPV6:       "ETH_P_IPV6",
	unix.ETH_P_IPX:        "ETH_P_IPX",
	unix.ETH_P_IRDA:       "ETH_P_IRDA",
	unix.ETH_P_LAT:        "ETH_P_LAT",
	unix.ETH_P_LINK_CTL:   "ETH_P_LINK_CTL",
	unix.ETH_P_LLDP:       "ETH_P_LLDP",
	unix.ETH_P_LOCALTALK:  "ETH_P_LOCALTALK",
	unix.ETH_P_LOOP:       "ETH_P_LOOP",
	unix.ETH_P_LOOPBACK:   "ETH_P_LOOPBACK",
	unix.ETH_P_MACSEC:     "ETH_P_MACSEC",
	unix.ETH_P_MAP:        "ETH_P_MAP",
	unix.ETH_P_MCTP:       "ETH_P_MCTP",
	unix.ETH_P_MOBITEX:    "ETH_P_MOBITEX",
	unix.ETH_P_MPLS_MC:    "ETH_P_MPLS_MC",
	unix.ETH_P_MPLS_UC:    "ETH_P_MPLS_UC",
	unix.ETH_P_MRP:        "ETH_P_MRP",
	unix.ETH_P_MVRP:       "ETH_P_MVRP",
	unix.ETH_P_NCSI:       "ETH_P_NCSI",
	unix.ETH_P_NSH:        "ETH_P_NSH",
	unix.ETH_P_PAE:        "ETH_P_PAE",
	unix.ETH_P_PAUSE:      "ETH_P_PAUSE",
	unix.ETH_P_PHONET:     "ETH_P_PHONET",
	unix.ETH_P_PPPTALK:    "ETH_P_PPPTALK",
	unix.ETH_P_PPP_DISC:   "ETH_P_PPP_DISC",
	unix.ETH_P_PPP_MP:     "ETH_P_PPP_MP",
	unix.ETH_P_PPP_SES:    "ETH_P_PPP_SES",
	unix.ETH_P_PREAUTH:    "ETH_P_PREAUTH",
	unix.ETH_P_PRP:        "ETH_P_PRP",
	unix.ETH_P_PUP:        "ETH_P_PUP",
	unix.ETH_P_PUPAT:      "ETH_P_PUPAT",
	unix.ETH_P_QINQ1:      "ETH_P_QINQ1",
	unix.ETH_P_QINQ2:      "ETH_P_QINQ2",
	unix.ETH_P_QINQ3:      "ETH_P_QINQ3",
	unix.ETH_P_RARP:       "ETH_P_RARP",
	unix.ETH_P_SCA:        "ETH_P_SCA",
	unix.ETH_P_SLOW:       "ETH_P_SLOW",
	unix.ETH_P_SNAP:       "ETH_P_SNAP",
	unix.ETH_P_TDLS:       "ETH_P_TDLS",
	unix.ETH_P_TEB:        "ETH_P_TEB",
	unix.ETH_P_TIPC:       "ETH_P_TIPC",
	unix.ETH_P_TRAILER:    "ETH_P_TRAILER",
	unix.ETH_P_TR_802_2:   "ETH_P_TR_802_2",
	unix.ETH_P_TSN:        "ETH_P_TSN",
	unix.ETH_P_WAN_PPP:    "ETH_P_WAN_PPP",
	unix.ETH_P_WCCP:       "ETH_P_WCCP",
	unix.ETH_P_X25:        "ETH_P_X25",
	unix.ETH_P_XDSA:       "ETH_P_XDSA",
}
