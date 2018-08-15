package responder

import (
	"context"

	"github.com/jmalloc/dissolve/src/dissolve/mdns/transport"
	"github.com/miekg/dns"
)

// newResponse returns a new (empty) response to a mDNS query.
//
// See https://tools.ietf.org/html/rfc6762#section-6 and
// https://tools.ietf.org/html/rfc6762#section-18.
func newResponse(query *dns.Msg, unicast bool) *dns.Msg {
	m := &dns.Msg{}
	m.SetReply(query)

	// https://tools.ietf.org/html/rfc6762#section-6
	//
	// Multicast DNS responses MUST NOT contain any questions in the
	// Question Section.  Any questions in the Question Section of a
	// received Multicast DNS response MUST be silently ignored.  Multicast
	// DNS queriers receiving Multicast DNS responses do not care what
	// question elicited the response; they care only that the information
	// in the response is true and accurate.
	m.Question = nil

	// https://tools.ietf.org/html/rfc6762#section-18.1
	//
	// In multicast responses, including unsolicited multicast responses,
	// the Query Identifier MUST be set to zero on transmission, and MUST be
	// ignored on reception.
	//
	// In legacy unicast response messages generated specifically in
	// response to a particular (unicast or multicast) query, the Query
	// Identifier MUST match the ID from the query message.
	if !unicast {
		m.Id = 0
	}

	// https://tools.ietf.org/html/rfc6762#section-18.3
	//
	// In both multicast query and multicast response messages, the OPCODE
	// MUST be zero on transmission (only standard queries are currently
	// supported over multicast).  Multicast DNS messages received with an
	// OPCODE other than zero MUST be silently ignored.
	m.Opcode = dns.OpcodeQuery

	// https://tools.ietf.org/html/rfc6762#section-18.4
	//
	// In response messages for Multicast domains, the Authoritative Answer
	// bit MUST be set to one (not setting this bit would imply there's some
	// other place where "better" information may be found) and MUST be
	// ignored on reception.
	m.Authoritative = true

	// https://tools.ietf.org/html/rfc6762#section-18.5
	//
	// From Section 18.5, the RFC goes on to say that the following bits must all
	// be set to zero:
	m.Truncated = false          // - 18.5: TC (TRUNCATED) Bit
	m.RecursionDesired = false   // - 18.6: RD (Recursion Desired) Bit
	m.RecursionAvailable = false // - 18.7: RA (Recursion Available) Bit
	m.Zero = false               // - 18.8: Z (Zero) Bit
	m.AuthenticatedData = false  // - 18.9: AD (Authentic Data) Bit
	m.CheckingDisabled = false   // - 18.10: CD (Checking Disabled) Bit
	m.Rcode = dns.RcodeSuccess   // - 18.11: RCODE (Response Code)

	// https://tools.ietf.org/html/rfc6762#section-18.14
	//
	// When generating Multicast DNS messages, implementations SHOULD use
	// name compression wherever possible to compress the names of resource
	// records, by replacing some or all of the resource record name with a
	// compact two-byte reference to an appearance of that data somewhere
	// earlier in the message [RFC1035].
	m.Compress = true

	return m
}

// handleResponse is a server command that handles a multicast DNS response.
type handleResponse struct {
	Packet  *transport.InboundPacket
	Message *dns.Msg
}

func (c *handleResponse) Execute(ctx context.Context, r *Responder) error {
	defer c.Packet.Close()
	// TODO(jmalloc): we need to "defend" our records here
	// https://tools.ietf.org/html/rfc6762#section-9
	return nil
}
