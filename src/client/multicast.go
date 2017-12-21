package client

import (
	"context"
	"time"

	"github.com/miekg/dns"
)

// Multicast is an interface for performing multicast DNS queries.
type Multicast interface {
	// Query performs a synchronous, multicast DNS query.
	Query(ctx context.Context, req *dns.Msg, wait time.Duration) (res *dns.Msg, err error)
}

// DefaultMulticast is the default multicast DNS client.
var (
	DefaultMulticast     Multicast = &StandardMulticast{}
	DefaultMulticastWait           = 20 * time.Millisecond // TODO
)

// StandardMulticast is Dissolve's standard multicast DNS client implementation.
type StandardMulticast struct {
}

// Query performs a synchronous, multicast DNS query.
func (c *StandardMulticast) Query(ctx context.Context, req *dns.Msg, wait time.Duration) (res *dns.Msg, err error) {
	panic("not impl")
}
