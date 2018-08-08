package bonjour

import (
	"context"
	"net"

	"github.com/jmalloc/dissolve/src/dissolve/resolver"

	"github.com/jmalloc/dissolve/src/dissolve/dnssd"
	"github.com/miekg/dns"
)

// Answerer answers mDNS questions DNS-SD services on the "local." domain.
type Answerer struct {
	answerer dnssd.Answerer
}

// AddInstance adds a service instance to the answerer.
// It panics if i is invalid, or if i.Domain is not "local."
func (a *Answerer) AddInstance(i *dnssd.Instance) {
	if i.Domain.DNSString() != "local." {
		panic("bonjour service must use the 'local' domain")
	}

	a.answerer.AddInstance(i)
}

// RemoveInstance removes a service instance from the answerer.
// It panics if n is invalid.
func (a *Answerer) RemoveInstance(n dnssd.InstanceName) {
	a.answerer.RemoveInstance(n)
}

// Answer populates m with the answer to q.
//
// r is the a resolver that should be used by the answerer if it needs to make
// DNS queries. s is the "source" address of the DNS query being answered.
func (a *Answerer) Answer(
	ctx context.Context,
	r resolver.Resolver,
	s net.Addr,
	q dns.Question,
	m *dns.Msg,
) error {
	return a.answerer.Answer(
		ctx,
		&LocalResolver{
			Resolver: r,
			Source:   s,
		},
		s,
		q,
		m,
	)
}
