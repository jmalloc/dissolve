package mdns

import (
	"errors"

	"github.com/miekg/dns"
)

// validateQuery returns an error if m is not a valid mDNS query.
func validateQuery(m *dns.Msg) error {
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
		return errors.New("OPCODE must be zero (query) in mDNS queries")
	}

	// https://tools.ietf.org/html/rfc6762#section-18.11
	//
	// "In both multicast query and multicast response messages, the Response
	// Code MUST be zero on transmission.  Multicast DNS messages received with
	// non-zero Response Codes MUST be silently ignored."
	if m.Rcode != 0 {
		return errors.New("RCODE must be zero in mDNS queries")
	}

	return nil
}

// wantsUnicastResponse returns true if the given question requested a unicast
// response.
//
// It returns a copy of the question with the "unicast response bit" cleared, to
// reflect the actual question class.
//
// See https://tools.ietf.org/html/rfc6762#section-18.12.
func wantsUnicastResponse(q dns.Question) (bool, dns.Question) {
	// In the Question Section of a Multicast DNS query, the top bit of the
	// qclass field is used to indicate that unicast responses are preferred
	// for this particular question.  (See Section 5.4.)
	const unicastResponseBit = 1 << 15

	u := q.Qclass & unicastResponseBit // read top-bit
	q.Qclass &^= unicastResponseBit    // clear top-bit

	return u != 0, q
}
