package transport

import "net"

// Endpoint is the origin or destination of a packet.
type Endpoint struct {
	InterfaceIndex int
	Address        *net.UDPAddr
}

// IsLegacy returns true if this endpoint is a "legacy" endpoint.
//
// A legacy endpoints is DNS querier that does not implement the full mDNS
// specification, and is expecting a "standard" unicast response.
//
// See https://tools.ietf.org/html/rfc6762#section-6.7.
func (ep *Endpoint) IsLegacy() bool {
	// If the source UDP port in a received Multicast DNS query is not port 5353,
	// this indicates that the querier originating the query is a simple resolver
	// such as described in Section 5.1, "One-Shot Multicast DNS Queries", which
	// does not fully implement all of Multicast DNS.
	return ep.Address.Port != Port
}
