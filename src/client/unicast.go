package client

import (
	"context"

	"github.com/miekg/dns"
)

// Unicast is an interface for performing unicast DNS queries.
type Unicast interface {
	// Query performs a synchronous, unicast DNS query.
	Query(ctx context.Context, req *dns.Msg, ns string) (res *dns.Msg, err error)
}

// DefaultUnicast is the default unicast DNS client.
var DefaultUnicast Unicast = &StandardUnicast{
	&dns.Client{},
}

// StandardUnicast is Dissolve's standad unicast DNS client implementation.
// It is a thin wrapper around dns.Client
type StandardUnicast struct {
	// Client is the underlying client to use. If it is nil, a zero-value
	// client is used.
	Client *dns.Client
}

// Query performs a synchronous, unicast DNS query.
func (c *StandardUnicast) Query(ctx context.Context, req *dns.Msg, ns string) (res *dns.Msg, err error) {
	cli := c.Client
	if cli == nil {
		cli = &dns.Client{}
	}

	res, _, err = cli.ExchangeContext(ctx, req, ns)
	return
}
