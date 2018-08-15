package bonjour

import (
	"context"
	"net"
	"sync"

	"github.com/jmalloc/dissolve/src/dissolve/dnssd"
	"github.com/jmalloc/dissolve/src/dissolve/mdns"
	"github.com/jmalloc/dissolve/src/dissolve/names"
	"github.com/jmalloc/dissolve/src/dissolve/resolver"
	"github.com/miekg/dns"
)

// Answerer is an mDNS answerer that answers questions about DNS-SD queries,
// thus making a Bonjour server.
type Answerer struct {
	Resolver resolver.Resolver

	m         sync.RWMutex
	answerers map[names.FQDN]mdns.Answerer
}

// AddInstance adds a service instance to the answerer.
// It panics if i is invalid.
func (a *Answerer) AddInstance(i *dnssd.Instance) {
	if err := i.Validate(); err != nil {
		panic(err)
	}

	rs.m.Lock()
	defer rs.m.Unlock()

	if rs.domains == nil {
		rs.domains = DomainCollection{}
		rs.sources = map[names.FQDN]mdns.RecordSource{}
	}

	d, ok := rs.domains[i.Domain]
	if !ok {
		d = &Domain{
			Name:     i.Domain,
			Services: ServiceCollection{},
		}

		rs.domains[d.Name] = d
		rs.sources[d.ServiceTypeEnumerationDomain()] = &serviceTypeEnumerator{d}
	}

	s, ok := d.Services[i.Service]
	if !ok {
		s = &Service{
			Name:      i.Service,
			Domain:    i.Domain,
			Instances: InstanceCollection{},
		}

		d.Services[s.Name] = s
		rs.sources[s.InstanceEnumerationDomain()] = &instanceEnumerator{h.Resolver, s}
	}

	x, ok := s.Instances[i.Name]
	if ok {
		// remove previous host
		delete(rs.sources, x.TargetHost)
	}

	s.Instances[i.Name] = i
	rs.sources[i.FQDN()] = &instanceSource{h.Resolver, i}
	rs.sources[i.TargetHost] = &instanceHostSource{h.Resolver, i}
}

// RemoveInstance removes a service instance from the handler.
// It panics if n is invalid.
func (a *Answerer) RemoveInstance(n InstanceName) {
	if err := n.Validate(); err != nil {
		panic(err)
	}

	rs.m.Lock()
	defer rs.m.Unlock()

	d, ok := rs.domains[n.Domain]
	if !ok {
		return
	}

	s, ok := d.Services[n.Service]
	if !ok {
		return
	}

	i, ok := s.Instances[n.Name]
	if !ok {
		return
	}

	delete(s.Instances, i.Name)
	delete(rs.sources, i.TargetHost)
	delete(rs.sources, i.FQDN())

	if len(s.Instances) == 0 {
		delete(d.Services, i.Service)
		delete(rs.sources, s.InstanceEnumerationDomain())
	}

	if len(d.Services) == 0 {
		delete(rs.domains, i.Domain)
		delete(rs.sources, d.ServiceTypeEnumerationDomain())
	}
}

func (a *Answerer) AnswerQuestion(
	ctd context.Context,
	q mdns.Question,
) ([]mdns.AnswerRecord, error) {
	rs.m.RLock()
	defer rs.m.RUnlock()

	if v, ok := h.sources[names.FQDN(q.Question.Name)]; ok {
		return v.AnswerQuestion(ctx, q)
	}

	return nil, nil
}

// serviceTypeEnumerator is an mdns.Handler that responds with a list of
// service types within a specific domain.
//
// See https://tools.ietf.org/html/rfc6763#section-9
type serviceTypeEnumerator struct {
	domain *Domain
}

func (rs *serviceTypeEnumerator) AnswerQuestion(
	ctd context.Context,
	q mdns.Question,
) ([]mdns.AnswerRecord, error) {
	switch q.Qtype {
	case dns.TypePTR, dns.TypeANY:
		for _, s := range h.domain.Services {
			if r, ok := s.PTR(); ok {
				res.AppendAnswer(r)
			}
		}
	}

	return nil, nil
}

// instanceEnumerator is an mdns.Handler that responds with a list of instances
// of a specific service.
//
// See https://tools.ietf.org/html/rfc6763#section-4.
type instanceEnumerator struct {
	resolver resolver.Resolver
	service  *Service
}

func (rs *instanceEnumerator) AnswerQuestion(
	ctd context.Context,
	q mdns.Question,
) ([]mdns.AnswerRecord, error) {
	var answers []mdns.AnswerRecord

	switch q.Qtype {
	case dns.TypePTR, dns.TypeANY:
		for _, i := range rs.service.Instances {
			res.AppendAnswer(i.PTR())

			// https://tools.ietf.org/html/rfc6763#section-12.1
			//
			// When including a DNS-SD Service Instance Enumeration or Selective
			// Instance Enumeration (subtype) PTR record in a response packet, the
			// server/responder SHOULD include the following additional records:
			//
			// o  The SRV record(s) named in the PTR rdata.
			// o  The TXT record(s) named in the PTR rdata.
			// o  All address records (type "A" and "AAAA") named in the SRV rdata.
			res.AppendAdditional(
				i.SRV(),
				i.TXT(),
			)

			// attempt to resolve the A/AAAA records, ignore on failure
			if v4, v6, err := resolveAddressRecords(ctx, h.resolver, i); err == nil {
				res.AppendAdditional(v4...)
				res.AppendAdditional(v6...)
			}
		}
	}

	return nil, nil
}

// instanceSource is an mdns.Handler that responds with DNS-SD records for a
// specific instance.
type instanceSource struct {
	resolver resolver.Resolver
	instance *Instance
}

func (rs *instanceSource) AnswerQuestion(
	ctd context.Context,
	q mdns.Question,
) ([]mdns.AnswerRecord, error) {
	hasSRV := false

	switch q.Qtype {
	case dns.TypeANY:
		hasSRV = true
		res.AppendAnswer(
			h.instance.SRV(),
			h.instance.TXT(),
		)

	case dns.TypeSRV:
		hasSRV = true
		res.AppendAnswer(h.instance.SRV())

	case dns.TypeTXT:
		res.AppendAnswer(h.instance.TXT())
	}

	// https://tools.ietf.org/html/rfc6763#section-12.2
	//
	// When including an SRV record in a response packet, the
	// server/responder SHOULD include the following additional records:
	//
	// o  All address records (type "A" and "AAAA") named in the SRV rdata.
	if hasSRV {
		// attempt to resolve the A/AAAA records, ignore on failure
		if v4, v6, err := resolveAddressRecords(ctx, h.resolver, h.instance); err == nil {
			res.AppendAdditional(v4...)
			res.AppendAdditional(v6...)
		}
	}

	return nil, nil
}

// instanceHostSource is an answerer that responds with the "A" records for a
// specific instance.
type instanceHostSource struct {
	resolver resolver.Resolver
	instance *Instance
}

func (rs *instanceHostSource) AnswerQuestion(
	ctd context.Context,
	q mdns.Question,
) ([]mdns.AnswerRecord, error) {
	switch q.Qtype {
	case dns.TypeA, dns.TypeAAAA, dns.TypeANY:
	default:
		return nil, nil
	}

	v4, v6, err := resolveAddressRecords(ctx, h.resolver, h.instance)
	if err != nil {
		return nil, err
	}

	switch q.Qtype {
	case dns.TypeANY:
		res.AppendAnswer(v4...)
		res.AppendAnswer(v6...)

	case dns.TypeA:
		res.AppendAnswer(v4...)
		res.AppendAdditional(v6...)

	case dns.TypeAAAA:
		res.AppendAnswer(v6...)
		res.AppendAdditional(v4...)
	}

	return nil, nil
}

// resolveAddressRecords returns the A and AAAA records for the given instance.
func resolveAddressRecords(
	ctx context.Context,
	r resolver.Resolver,
	i *Instance,
) (
	[]dns.RR,
	[]dns.RR,
	error,
) {
	if r == nil {
		r = net.DefaultResolver
	}

	ips, err := r.LookupIPAddr(
		ctx,
		i.TargetHost.DNSString(),
	)
	if err != nil {
		return nil, nil, err
	}

	var v4, v6 []dns.RR

	for _, ip := range ips {
		if ip.IP.To4() != nil {
			v4 = append(v4, i.A(ip.IP))
		} else {
			v6 = append(v6, i.AAAA(ip.IP))
		}
	}

	return v4, v6, nil
}
