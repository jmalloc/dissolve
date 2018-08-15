package responder

import (
	"context"
	"errors"

	"github.com/jmalloc/dissolve/src/dissolve/mdns/transport"
	"github.com/miekg/dns"
)

// newQuery returns a new (empty) DNS query.
func newQuery(legacy bool, q ...dns.Question) *dns.Msg {
	m := &dns.Msg{}

	// https://tools.ietf.org/html/rfc6762#section-18.1
	//
	// In multicast query messages, the Query Identifier SHOULD be set to
	// zero on transmission.
	if !legacy {
		m.Id = dns.Id()
	}

	// https://tools.ietf.org/html/rfc6762#section-18.3
	//
	// "In both multicast query and multicast response messages, the OPCODE MUST
	// be zero on transmission (only standard queries are currently supported
	// over multicast).  Multicast DNS messages received with an OPCODE other
	// than zero MUST be silently ignored."  Note: OpcodeQuery == 0
	m.Opcode = dns.OpcodeQuery

	// https://tools.ietf.org/html/rfc6762#section-18.4
	//
	// In query messages, the Authoritative Answer bit MUST be zero on
	// transmission, and MUST be ignored on reception.
	m.Authoritative = false

	// https://tools.ietf.org/html/rfc6762#section-18.5
	//
	// In query messages, if the TC bit is set, it means that additional
	// Known-Answer records may be following shortly.  A responder SHOULD
	// record this fact, and wait for those additional Known-Answer records,
	// before deciding whether to respond.  If the TC bit is clear, it means
	// that the querying host has no additional Known Answers.
	//
	// TODO(jmalloc): support for truncated known-answer records.
	m.Truncated = false

	// https://tools.ietf.org/html/rfc6762#section-18.6
	//
	// From Section 18.6, the RFC goes on to say that the following bits must all
	// be set to zero:
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

// In the Question Section of a Multicast DNS query, the top bit of the
// qclass field is used to indicate that unicast responses are preferred
// for this particular question.  (See Section 5.4.)
//
// See https://tools.ietf.org/html/rfc6762#section-18.12.
const unicastResponseBit = 1 << 15

// wantsUnicastResponse returns true if the given question requested a unicast
// response.
//
// It returns a copy of the question with the "unicast response bit" cleared, to
// reflect the actual question class.
//
// See https://tools.ietf.org/html/rfc6762#section-18.12.
func wantsUnicastResponse(q dns.Question) (bool, dns.Question) {
	u := q.Qclass & unicastResponseBit // read top-bit
	q.Qclass &^= unicastResponseBit    // clear top-bit

	return u != 0, q
}

// setUnicastResponse adds the "unicast response bit" to the given question.
func setUnicastResponse(q dns.Question) dns.Question {
	q.Qclass |= unicastResponseBit
	return q
}

// handleQuery is a command that handles a DNS query.
type handleQuery struct {
	Packet  *transport.InboundPacket
	Message *dns.Msg
}

func (c *handleQuery) Execute(ctx context.Context, r *Responder) error {
	defer c.Packet.Close()

	err := validateQuery(c.Message)
	if err != nil {
		return err
	}

	var (
		legacy = c.Packet.Source.IsLegacy()
		uRes   = newResponse(c.Message, true)
		mRes   = newResponse(c.Message, false)
	)

	for _, rawQ := range c.Message.Question {
		unicast, dnsQ := wantsUnicastResponse(rawQ)

		var (
			q = Question{
				Question:  dnsQ,
				Query:     c.Message,
				Interface: *r.iface,
			}
			a = Answer{}
		)

		err = r.answerer.Answer(ctx, &q, &a)
		if err != nil {
			return err
		}

		if !a.Unique.IsEmpty() {
			// TODO(jmalloc): probe/announce uniquely-scoped records before
			// providing answers to them.
		}

		if unicast || legacy {
			a.appendToMessage(uRes)
		} else {
			a.appendToMessage(mRes)
		}
	}

	_, err = transport.SendUnicastResponse(c.Packet, uRes)
	if err != nil {
		return err
	}

	_, err = transport.SendMulticastResponse(c.Packet, mRes)
	if err != nil {
		return err
	}

	// this is a no-op unless compiled with the 'debug' build tag
	dumpRequestResponse(c.Packet, c.Message, uRes, mRes)

	return nil
}
