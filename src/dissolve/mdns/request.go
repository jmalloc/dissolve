package mdns

import (
	"errors"
	"net"

	"github.com/miekg/dns"
)

// Source is the origin of an mDNS request.
type Source struct {
	Interface int
	Address   *net.UDPAddr
}

// Request is an mDNS request.
type Request struct {
	Source  Source
	Message *dns.Msg
}

// NewRequest constructs a new mDNS from a DNS message.
func NewRequest(src Source, m *dns.Msg) (*Request, error) {
	if m.Response {
		panic("DNS message is a response")
	}

	// https://tools.ietf.org/html/rfc6762#section-18.3
	//
	// "In both multicast query and multicast response messages, the OPCODE MUST
	// be zero on transmission (only standard queries are currently supported
	// over multicast).  Multicast DNS messages received with an OPCODE other
	// than zero MUST be silently ignored."  Note: OpcodeQuery == 0
	if m.Opcode != dns.OpcodeQuery {
		return nil, errors.New("OPCODE must be zero (query) in mDNS requests")
	}

	// https://tools.ietf.org/html/rfc6762#section-18.11
	//
	// "In both multicast query and multicast response messages, the Response
	// Code MUST be zero on transmission.  Multicast DNS messages received with
	// non-zero Response Codes MUST be silently ignored."
	if m.Rcode != 0 {
		return nil, errors.New("RCODE must be zero in mDNS requests")
	}

	return &Request{src, m}, nil
}

// IsLegacy returns true if this request is a "legacy" request.
//
// A legacy request is a request sent from a DNS querier that does not implement
// the full mDNS specification, and is expecting a "standard" unicast response.
//
// See https://tools.ietf.org/html/rfc6762#section-6.7.
func (r *Request) IsLegacy() bool {
	// If the source UDP port in a received Multicast DNS query is not port 5353,
	// this indicates that the querier originating the query is a simple resolver
	// such as described in Section 5.1, "One-Shot Multicast DNS Queries", which
	// does not fully implement all of Multicast DNS.
	return r.Source.Address.Port != 5353
}

// WantsUnicastResponse returns true if the given question requested a unicast
// response.
//
// It returns a copy of the question with the "unicast response bit" cleared, to
// reflect the actual question class.
//
// See https://tools.ietf.org/html/rfc6762#section-18.12.
func WantsUnicastResponse(q dns.Question) (bool, dns.Question) {
	// In the Question Section of a Multicast DNS query, the top bit of the
	// qclass field is used to indicate that unicast responses are preferred
	// for this particular question.  (See Section 5.4.)
	const unicastResponseBit = 1 << 15

	u := q.Qclass & unicastResponseBit // read top-bit
	q.Qclass &^= unicastResponseBit    // clear top-bit

	return u != 0, q
}
