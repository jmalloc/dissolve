package responder

import (
	"context"

	"github.com/jmalloc/dissolve/src/dissolve/mdns/transport"
	"github.com/miekg/dns"
)

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
