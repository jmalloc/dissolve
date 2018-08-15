package bonjour

import (
	"context"
	"sync"

	"github.com/jmalloc/dissolve/src/dissolve/dnssd"
	"github.com/jmalloc/dissolve/src/dissolve/mdns/responder"
	"github.com/jmalloc/dissolve/src/dissolve/names"
	"github.com/jmalloc/dissolve/src/dissolve/resolver"
)

// Answerer is an mDNS answerer that answers questions about DNS-SD services,
// thus making a Bonjour server.
type Answerer struct {
	Resolver resolver.Resolver

	m         sync.RWMutex
	domains   dnssd.DomainCollection
	answerers map[names.FQDN]responder.Answerer
}

// AddInstance adds a service instance to the answerer.
//
// It panics if i is invalid.
func (an *Answerer) AddInstance(i *dnssd.Instance) {
	if err := i.Validate(); err != nil {
		panic(err)
	}

	an.m.Lock()
	defer an.m.Unlock()

	if an.domains == nil {
		an.domains = dnssd.DomainCollection{}
		an.answerers = map[names.FQDN]responder.Answerer{}
	}

	d, ok := an.domains[i.Domain]
	if !ok {
		d = &dnssd.Domain{
			Name:     i.Domain,
			Services: dnssd.ServiceCollection{},
		}

		an.domains[d.Name] = d
		an.answerers[d.TypeEnumDomain()] = &typeEnumAnswerer{d}
	}

	s, ok := d.Services[i.ServiceType]
	if !ok {
		s = &dnssd.Service{
			Type:      i.ServiceType,
			Domain:    i.Domain,
			Instances: dnssd.InstanceCollection{},
		}

		d.Services[s.Type] = s
		an.answerers[s.InstanceEnumDomain()] = &instanceEnumAnswerer{an.Resolver, s}
	}

	x, ok := s.Instances[i.Name]
	if ok {
		// remove previous host
		delete(an.answerers, x.TargetFQDN())
	}

	s.Instances[i.Name] = i
	an.answerers[i.FQDN()] = &instanceAnswerer{an.Resolver, i}
	an.answerers[i.TargetFQDN()] = &targetAnswerer{an.Resolver, i}
}

// RemoveInstance removes a service instance from the handler.
func (an *Answerer) RemoveInstance(
	instance dnssd.InstanceName,
	service dnssd.ServiceType,
	domain names.FQDN,
) {
	an.m.Lock()
	defer an.m.Unlock()

	d, ok := an.domains[domain]
	if !ok {
		return
	}

	s, ok := d.Services[service]
	if !ok {
		return
	}

	i, ok := s.Instances[instance]
	if !ok {
		return
	}

	delete(s.Instances, i.Name)
	delete(an.answerers, i.TargetFQDN())
	delete(an.answerers, i.FQDN())

	if len(s.Instances) == 0 {
		delete(d.Services, i.ServiceType)
		delete(an.answerers, s.InstanceEnumDomain())
	}

	if len(d.Services) == 0 {
		delete(an.domains, i.Domain)
		delete(an.answerers, d.TypeEnumDomain())
	}
}

// Answer populates an answer to a single DNS question.
func (an *Answerer) Answer(
	ctx context.Context,
	q *responder.Question,
	a *responder.Answer,
) error {
	an.m.RLock()
	defer an.m.RUnlock()

	if v, ok := an.answerers[names.FQDN(q.Question.Name)]; ok {
		return v.Answer(ctx, q, a)
	}

	return nil
}
