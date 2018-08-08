package dnssd

import (
	"context"
	"net"
	"sync"

	"github.com/jmalloc/dissolve/src/dissolve/names"
	"github.com/jmalloc/dissolve/src/dissolve/resolver"
	"github.com/jmalloc/dissolve/src/dissolve/server"
	"github.com/miekg/dns"
)

// Answerer answers DNS questions about DNS-SD services.
type Answerer struct {
	m         sync.RWMutex
	domains   DomainCollection
	answerers map[names.FQDN]server.Answerer
}

// AddInstance adds a service instance to the answerer.
// It panics if i is invalid.
func (a *Answerer) AddInstance(i *Instance) {
	if err := i.Validate(); err != nil {
		panic(err)
	}

	a.m.Lock()
	defer a.m.Unlock()

	if a.domains == nil {
		a.domains = DomainCollection{}
		a.answerers = map[names.FQDN]server.Answerer{}
	}

	d, ok := a.domains[i.Domain]
	if !ok {
		d = &Domain{
			Name:     i.Domain,
			Services: ServiceCollection{},
		}

		a.domains[d.Name] = d
		a.answerers[d.ServiceTypeEnumerationDomain()] = &serviceTypeEnumerator{d}
	}

	s, ok := d.Services[i.Service]
	if !ok {
		s = &Service{
			Name:      i.Service,
			Domain:    i.Domain,
			Instances: InstanceCollection{},
		}

		d.Services[s.Name] = s
		a.answerers[s.InstanceEnumerationDomain()] = &instanceEnumerator{s}
	}

	x, ok := s.Instances[i.Name]
	if ok {
		// remove previous host
		delete(a.answerers, x.TargetHost)
	}

	s.Instances[i.Name] = i
	a.answerers[i.FQDN()] = &instanceAnswerer{i}
	a.answerers[i.TargetHost] = &instanceHostAnswerer{i}
}

// RemoveInstance removes a service instance from the answerer.
// It panics if n is invalid.
func (a *Answerer) RemoveInstance(n InstanceName) {
	if err := n.Validate(); err != nil {
		panic(err)
	}

	a.m.Lock()
	defer a.m.Unlock()

	d, ok := a.domains[n.Domain]
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
	delete(a.answerers, i.TargetHost)
	delete(a.answerers, i.FQDN())

	if len(s.Instances) == 0 {
		delete(d.Services, i.Service)
		delete(a.answerers, s.InstanceEnumerationDomain())
	}

	if len(d.Services) == 0 {
		delete(a.domains, i.Domain)
		delete(a.answerers, d.ServiceTypeEnumerationDomain())
	}
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
	a.m.RLock()
	defer a.m.RUnlock()

	if v, ok := a.answerers[names.FQDN(q.Name)]; ok {
		return v.Answer(ctx, r, s, q, m)
	}

	return nil
}

// serviceTypeEnumerator is an answerer that responds with a list of services
// types within a specific domain.
//
// See https://tools.ietf.org/html/rfc6763#section-9
type serviceTypeEnumerator struct {
	domain *Domain
}

func (a *serviceTypeEnumerator) Answer(
	ctx context.Context,
	r resolver.Resolver,
	s net.Addr,
	q dns.Question,
	m *dns.Msg,
) error {
	switch q.Qtype {
	case dns.TypePTR, dns.TypeANY:
		for _, s := range a.domain.Services {
			if r, ok := s.PTR(); ok {
				m.Answer = append(m.Answer, r)
			}
		}
	}

	return nil
}

// instanceEnumerator is an answerer that responds with a list of instances
// of a specific service.
//
// See https://tools.ietf.org/html/rfc6763#section-4.
type instanceEnumerator struct {
	service *Service
}

func (a *instanceEnumerator) Answer(
	ctx context.Context,
	r resolver.Resolver,
	s net.Addr,
	q dns.Question,
	m *dns.Msg,
) error {
	switch q.Qtype {
	case dns.TypePTR, dns.TypeANY:
		for _, i := range a.service.Instances {
			m.Answer = append(
				m.Answer,
				i.PTR(),
			)

			// https://tools.ietf.org/html/rfc6763#section-12.1
			//
			// When including a DNS-SD Service Instance Enumeration or Selective
			// Instance Enumeration (subtype) PTR record in a response packet, the
			// server/responder SHOULD include the following additional records:
			//
			// o  The SRV record(s) named in the PTR rdata.
			// o  The TXT record(s) named in the PTR rdata.
			// o  All address records (type "A" and "AAAA") named in the SRV rdata.
			m.Extra = append(m.Extra, i.SRV(), i.TXT())

			// attempt to resolve the A/AAAA records, ignore on failure
			if v4, v6, err := resolveAddressRecords(ctx, r, i); err == nil {
				m.Extra = append(m.Extra, v4...)
				m.Extra = append(m.Extra, v6...)
			}
		}
	}

	return nil
}

// instanceAnswerer is an answerer that responds with DNS-SD records for a
// specific instance.
type instanceAnswerer struct {
	instance *Instance
}

func (a *instanceAnswerer) Answer(
	ctx context.Context,
	r resolver.Resolver,
	s net.Addr,
	q dns.Question,
	m *dns.Msg,
) error {
	hasSRV := false

	switch q.Qtype {
	case dns.TypeANY:
		hasSRV = true
		m.Answer = append(
			m.Answer,
			a.instance.SRV(),
			a.instance.TXT(),
		)

	case dns.TypeSRV:
		hasSRV = true
		m.Answer = append(m.Answer, a.instance.SRV())

	case dns.TypeTXT:
		m.Answer = append(m.Answer, a.instance.TXT())
	}

	// https://tools.ietf.org/html/rfc6763#section-12.2
	//
	// When including an SRV record in a response packet, the
	// server/responder SHOULD include the following additional records:
	//
	// o  All address records (type "A" and "AAAA") named in the SRV rdata.
	if hasSRV {
		// attempt to resolve the A/AAAA records, ignore on failure
		if v4, v6, err := resolveAddressRecords(ctx, r, a.instance); err == nil {
			m.Extra = append(m.Extra, v4...)
			m.Extra = append(m.Extra, v6...)
		}
	}

	return nil
}

// instanceAnswerer is an answerer that responds with the "A" records for a
// specific instance.
type instanceHostAnswerer struct {
	instance *Instance
}

func (a *instanceHostAnswerer) Answer(
	ctx context.Context,
	r resolver.Resolver,
	s net.Addr,
	q dns.Question,
	m *dns.Msg,
) error {
	switch q.Qtype {
	case dns.TypeA, dns.TypeAAAA, dns.TypeANY:
	default:
		return nil
	}

	v4, v6, err := resolveAddressRecords(ctx, r, a.instance)
	if err != nil {
		return err
	}

	switch q.Qtype {
	case dns.TypeANY:
		m.Answer = append(m.Answer, v4...)
		m.Answer = append(m.Answer, v6...)

	case dns.TypeA:
		m.Answer = append(m.Answer, v4...)
		m.Extra = append(m.Extra, v6...)

	case dns.TypeAAAA:
		m.Answer = append(m.Answer, v6...)
		m.Extra = append(m.Extra, v4...)
	}

	return nil
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
