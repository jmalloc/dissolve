package bonjour

import (
	"context"

	"github.com/jmalloc/dissolve/src/dissolve/dnssd"

	"github.com/jmalloc/dissolve/src/dissolve/mdns/responder"
	"github.com/jmalloc/dissolve/src/dissolve/resolver"
	"github.com/miekg/dns"
)

// typeEnumAnswerer is an mdns.Handler that responds with a list of
// service types within a specific domain.
//
// See https://tools.ietf.org/html/rfc6763#section-9
type typeEnumAnswerer struct {
	Domain *dnssd.Domain
}

func (an *typeEnumAnswerer) Answer(
	ctd context.Context,
	q *responder.Question,
	a *responder.Answer,
) error {
	switch q.Qtype {
	case dns.TypePTR, dns.TypeANY:
		for _, s := range an.Domain.Services {
			if r, ok := s.PTR(); ok {
				a.Shared.Answer(r)
			}
		}
	}

	return nil
}

// instanceEnumAnswerer is an mDNS answerer that responds with a list of
// instances of a specific service.
//
// See https://tools.ietf.org/html/rfc6763#section-4.
type instanceEnumAnswerer struct {
	Resolver resolver.Resolver
	Service  *dnssd.Service
}

func (an *instanceEnumAnswerer) Answer(
	ctx context.Context,
	q *responder.Question,
	a *responder.Answer,
) error {
	switch q.Qtype {
	case dns.TypePTR, dns.TypeANY:
		for _, i := range an.Service.Instances {
			a.Unique.Answer(i.PTR())

			// https://tools.ietf.org/html/rfc6763#section-12.1
			//
			// When including a DNS-SD Service Instance Enumeration or Selective
			// Instance Enumeration (subtype) PTR record in a response packet, the
			// server/responder SHOULD include the following additional records:
			//
			// o  The SRV record(s) named in the PTR rdata.
			// o  The TXT record(s) named in the PTR rdata.
			// o  All address records (type "A" and "AAAA") named in the SRV rdata.
			a.Unique.Additional(
				i.SRV(),
				i.TXT(),
			)

			// attempt to resolve the A/AAAA records, ignore on failure
			if v4, v6, err := addressRecords(ctx, an.Resolver, q.Interface, i); err == nil {
				a.Unique.Additional(v4...)
				a.Unique.Additional(v6...)
			}
		}
	}

	return nil
}
