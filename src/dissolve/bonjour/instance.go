package bonjour

import (
	"context"

	"github.com/jmalloc/dissolve/src/dissolve/dnssd"

	"github.com/jmalloc/dissolve/src/dissolve/mdns/responder"
	"github.com/jmalloc/dissolve/src/dissolve/resolver"
	"github.com/miekg/dns"
)

// instanceAnswerer is a responder.Answerer that provides answers about the
// DNS-SD records associated with a service instance.
type instanceAnswerer struct {
	Resolver resolver.Resolver
	Instance *dnssd.Instance
}

func (an *instanceAnswerer) Answer(
	ctx context.Context,
	q *responder.Question,
	a *responder.Answer,
) error {
	hasSRV := false

	switch q.Qtype {
	case dns.TypeANY:
		hasSRV = true
		a.Unique.Answer(
			an.Instance.SRV(),
			an.Instance.TXT(),
		)

	case dns.TypeSRV:
		hasSRV = true
		a.Unique.Answer(an.Instance.SRV())

	case dns.TypeTXT:
		a.Unique.Answer(an.Instance.TXT())
	}

	// https://tools.ietf.org/html/rfc6763#section-12.2
	//
	// When including an SRV record in a response packet, the
	// server/responder SHOULD include the following additional records:
	//
	// o  All address records (type "A" and "AAAA") named in the SRV rdata.
	if hasSRV {
		// attempt to resolve the A/AAAA records, ignore on failure
		if v4, v6, err := addressRecords(ctx, an.Resolver, q.Interface, an.Instance); err == nil {
			a.Unique.Additional(v4...)
			a.Unique.Additional(v6...)
		}
	}

	return nil
}
