package mdns

import "net"

// Port is the mDNS port number.
const Port = 5353

var (
	// IPv4Group is the multicast group used for mDNS over IPv4.
	//
	// See https://tools.ietf.org/html/rfc6762#section-3.
	IPv4Group = net.ParseIP("224.0.0.251")

	// IPv4Address is the address to which mDNS queries are sent when using IPv4.
	//
	// See https://tools.ietf.org/html/rfc6762#section-3.
	IPv4Address = &net.UDPAddr{IP: IPv4Group, Port: Port}

	// IPv6Group is the multicast group used for mDNS over IPv6.
	//
	// See https://tools.ietf.org/html/rfc6762#section-3.
	IPv6Group = net.ParseIP("ff02::fb")

	// IPv6Address is the address to which mDNS queries are sent when using IPv6.
	//
	// See https://tools.ietf.org/html/rfc6762#section-3.
	IPv6Address = &net.UDPAddr{IP: IPv6Group, Port: Port}
)
