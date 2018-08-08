package dnssd

import (
	"github.com/jmalloc/dissolve/src/dissolve/names"
	"github.com/miekg/dns"
)

// ServiceCollection is the map of service name (such as "_http._tcp") to the
// service.
type ServiceCollection map[names.Rel]*Service

// InstanceEnumerationDomain returns the DNS name that is queried to perform
// "service instance enumeration" (aka "browse") on a service within a given
// domain.
//
// See https://tools.ietf.org/html/rfc6763#section-4.
func InstanceEnumerationDomain(service names.Rel, domain names.FQDN) names.FQDN {
	return service.Qualify(domain)
}

// Service represents a DNS-SD service.
type Service struct {
	// Name is the DNS-SD service name, including the protocol, such as
	// "_http._tcp".
	Name names.Rel

	// Domain is the fully-qualified name of the domain that the service is
	// advertised within. For example, Bonjour typically uses "local."
	Domain names.FQDN

	// Instance is the set of instances of this services, keyed by the instance
	// name.
	Instances InstanceCollection
}

// InstanceEnumerationDomain returns the DNS name that is queried to perform
// "service instance enumeration" (aka "browse") on a service within this
// domain.
//
// See https://tools.ietf.org/html/rfc6763#section-4.
func (s *Service) InstanceEnumerationDomain() names.FQDN {
	return InstanceEnumerationDomain(s.Name, s.Domain)
}

// PTR returns the service's PTR record as queried when performing "service type
// enumeration".
//
// It returns false if the service does not have any instances.
//
// See https://tools.ietf.org/html/rfc6763#section-9.
func (s *Service) PTR() (*dns.PTR, bool) {
	if len(s.Instances) == 0 {
		return nil, false
	}

	var ttl uint32

	for _, i := range s.Instances {
		v := i.TTLInSeconds()
		if v > ttl {
			ttl = v
		}
	}

	return &dns.PTR{
		Hdr: dns.RR_Header{
			Name:   ServiceTypeEnumerationDomain(s.Domain).DNSString(),
			Rrtype: dns.TypePTR,
			Class:  dns.ClassINET,
			Ttl:    ttl,
		},
		Ptr: s.InstanceEnumerationDomain().DNSString(),
	}, true
}

// Validate returns an error if the service is configured incorrectly.
func (s *Service) Validate() error {
	if err := s.Name.Validate(); err != nil {
		return err
	}

	return s.Domain.Validate()
}
