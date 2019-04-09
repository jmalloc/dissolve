package responder

import (
	"context"

	"github.com/jmalloc/dissolve/src/dissolve/mdns"
	"github.com/jmalloc/dissolve/src/dissolve/mdns/transport"
	"github.com/miekg/dns"
)

// handleQuery is a command that handles a DNS query.
type handleQuery struct {
	Packet  *transport.InboundPacket
	Message *dns.Msg
}

func (c *handleQuery) Execute(ctx context.Context, r *Responder) error {
	if err := c.query(ctx, r); err != nil {
		r.logger.Log("error handling mDNS query: %s", err)
	}

	return nil
}

func (c *handleQuery) query(ctx context.Context, r *Responder) error {
	defer c.Packet.Close()

	err := mdns.ValidateQuery(c.Message)
	if err != nil {
		return err
	}

	var (
		legacy = c.Packet.Source.IsLegacy()
		uRes   = mdns.NewResponse(c.Message, true)
		mRes   = mdns.NewResponse(c.Message, false)
	)

	for _, rawQ := range c.Message.Question {
		unicast, dnsQ := mdns.WantsUnicastResponse(rawQ)

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

		// TODO(jmalloc): probe/announce uniquely-scoped records before
		// providing answers to them.

		if unicast || legacy {
			a.appendToMessage(uRes, legacy)
		} else {
			a.appendToMessage(mRes, false)
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
