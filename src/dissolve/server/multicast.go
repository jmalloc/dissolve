package server

import (
	"context"
	"net"

	"github.com/davecgh/go-spew/spew"
	"github.com/jmalloc/dissolve/src/dissolve/resolver"
	"github.com/jmalloc/twelf/src/twelf"
	"github.com/miekg/dns"
)

var mdnsIPv4Group = &net.UDPAddr{
	IP:   net.ParseIP("224.0.0.251"),
	Port: 5353,
}

// MulticastServer is a mDNS (multicast DNS) server.
//
// See https://tools.ietf.org/html/rfc6762.
type MulticastServer struct {
	Answerer Answerer
	Resolver resolver.Resolver
	Logger   twelf.Logger

	v4con *net.UDPConn
}

// Run answers mDNS requests until ctx is canceled or an error occurs.
func (s *MulticastServer) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	v4, err := net.ListenMulticastUDP("udp4", nil, mdnsIPv4Group)
	if err != nil {
		return err
	}

	// close the connection if the context is canceled.
	go func() {
		<-ctx.Done()
		_ = v4.Close()
	}()

	s.v4con = v4

	return s.recv(ctx)
}

func (s *MulticastServer) recv(ctx context.Context) error {
	buf := make([]byte, 65536)

	for {
		n, src, err := s.v4con.ReadFromUDP(buf)
		if err != nil {
			s.Logger.Log("error reading mDNS request: %s", err)
			// TODO(jmalloc): check for "closed" error and return ctx.Err() instead
			return err
		}

		var req dns.Msg

		if err := req.Unpack(buf[:n]); err != nil {
			s.Logger.Log("error parsing mDNS request: %s", err)
		}

		if err := s.handleQuery(ctx, src, &req); err != nil {
			s.Logger.Log("error handling mDNS request: %s", err)
		}
	}
}

func (s *MulticastServer) handleQuery(ctx context.Context, src *net.UDPAddr, req *dns.Msg) error {
	// https://tools.ietf.org/html/rfc6762#section-18.3
	//
	// "In both multicast query and multicast response messages, the OPCODE MUST
	// be zero on transmission (only standard queries are currently supported
	// over multicast).  Multicast DNS messages received with an OPCODE other
	// than zero MUST be silently ignored."  Note: OpcodeQuery == 0
	if req.Opcode != dns.OpcodeQuery {
		s.Logger.DebugString("ignoring mDNS query with not-zero OPCODE")
		return nil
	}

	// https://tools.ietf.org/html/rfc6762#section-18.11
	//
	// "In both multicast query and multicast response messages, the Response
	// Code MUST be zero on transmission.  Multicast DNS messages received with
	// non-zero Response Codes MUST be silently ignored."
	if req.Rcode != 0 {
		s.Logger.DebugString("ignoring mDNS query with not-zero RCODE")
		return nil
	}

	// https://tools.ietf.org/html/rfc6762#section-18.5
	//
	// In query messages, if the TC bit is set, it means that additional
	// Known-Answer records may be following shortly.  A responder SHOULD
	// record this fact, and wait for those additional Known-Answer records,
	// before deciding whether to respond.  If the TC bit is clear, it means
	// that the querying host has no additional Known Answers.
	//
	// We do not currently support this optional behaviour.
	if req.Truncated {
		s.Logger.DebugString("responding immediately to mDNS query with non-zero TC flag")
	}

	// https://tools.ietf.org/html/rfc6762#section-18.1
	//
	// In legacy unicast response messages generated specifically in
	// response to a particular (unicast or multicast) query, the Query
	// Identifier MUST match the ID from the query message.
	uc := newResponse(req, req.Id)

	// https://tools.ietf.org/html/rfc6762#section-18.1
	//
	// In multicast responses, including unsolicited multicast responses,
	// the Query Identifier MUST be set to zero on transmission, and MUST be
	// ignored on reception.
	mc := newResponse(req, 0)

	r := s.Resolver
	if r == nil {
		r = net.DefaultResolver
	}

	// Ask each answerer for results for each question.
	for _, q := range req.Question {
		var res *dns.Msg
		if wantsUnicastResponse(q) {
			res = uc
		} else {
			res = mc
		}

		spew.Dump(q)
		s.Answerer.Answer(ctx, r, src, q, res)
	}

	// Send a unicast response.
	if len(uc.Answer) != 0 {
		spew.Dump(uc)

		buf, err := uc.Pack()
		if err != nil {
			return err
		}

		if _, err := s.v4con.WriteToUDP(buf, src); err != nil {
			return err
		}
	}

	// Send a multicast response.
	if len(mc.Answer) != 0 {
		spew.Dump(mc)

		buf, err := uc.Pack()
		if err != nil {
			return err
		}

		if _, err = s.v4con.WriteToUDP(buf, mdnsIPv4Group); err != nil {
			return err
		}
	}

	return nil
}

// wantsUnicastResponse returns true if the given question wants a unicast
// response.
func wantsUnicastResponse(q dns.Question) bool {
	// https://tools.ietf.org/html/rfc6762#section-18.12
	//
	// In the Question Section of a Multicast DNS query, the top bit of the
	// qclass field is used to indicate that unicast responses are preferred
	// for this particular question.  (See Section 5.4.)
	return q.Qclass&(1<<15) != 0
}

// newResponse returns a new empty response to the given request.
//
// See https://tools.ietf.org/html/rfc6762#section-18
func newResponse(req *dns.Msg, id uint16) *dns.Msg {
	return &dns.Msg{
		MsgHdr: dns.MsgHdr{
			// https://tools.ietf.org/html/rfc6762#section-18.1
			Id: id,

			// https://tools.ietf.org/html/rfc6762#section-18.2
			//
			// In response messages the QR bit MUST be one.
			Response: true,

			// https://tools.ietf.org/html/rfc6762#section-18.3
			//
			// In both multicast query and multicast response messages, the OPCODE
			// MUST be zero on transmission (only standard queries are currently
			// supported over multicast).  Multicast DNS messages received with an
			// OPCODE other than zero MUST be silently ignored.
			Opcode: dns.OpcodeQuery,

			// https://tools.ietf.org/html/rfc6762#section-18.4
			//
			// In response messages for Multicast domains, the Authoritative Answer
			// bit MUST be set to one (not setting this bit would imply there's some
			// other place where "better" information may be found) and MUST be
			// ignored on reception.
			Authoritative: true,

			// https://tools.ietf.org/html/rfc6762#section-18.5
			//
			// From Section 18.5, the RFC goes on to say that the following bits must all
			// be set to zero:
			//
			//   - 18.5: TC (TRUNCATED) Bit
			//   - 18.6: RD (Recursion Desired) Bit
			//   - 18.7: RA (Recursion Available) Bit
			//   - 18.8: Z (Zero) Bit
			//   - 18.9: AD (Authentic Data) Bit
			//   - 18.10: CD (Checking Disabled) Bit
			//   - 18.11: RCODE (Response Code)
		},

		// https://tools.ietf.org/html/rfc6762#section-18.14
		//
		// When generating Multicast DNS messages, implementations SHOULD use
		// name compression wherever possible to compress the names of resource
		// records, by replacing some or all of the resource record name with a
		// compact two-byte reference to an appearance of that data somewhere
		// earlier in the message [RFC1035].
		Compress: true,
	}
}
