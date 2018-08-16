package dnssd

import (
	"fmt"

	"github.com/jmalloc/dissolve/src/dissolve/names"
	"github.com/miekg/dns"
)

// ServiceCollection is the map of service type (such as "_http._tcp") to the
// service.
type ServiceCollection map[ServiceType]*Service

// Add adds a domain to the collection.
func (c ServiceCollection) Add(s *Service) {
	c[s.Type] = s
}

// Service represents a DNS-SD service.
type Service struct {
	// Type is the DNS-SD service type, including the protocol, such as
	// "_http._tcp".
	Type ServiceType

	// Domain is the fully-qualified name of the domain that the service is
	// advertised within. For example, Bonjour typically uses "local."
	Domain names.FQDN

	// Instance is the set of instances of this services, keyed by the instance
	// name.
	Instances InstanceCollection
}

// InstanceEnumDomain returns the DNS name that is queried to perform
// "service instance enumeration" (aka "browse") on a service within this
// domain.
//
// See https://tools.ietf.org/html/rfc6763#section-4.
func (s *Service) InstanceEnumDomain() names.FQDN {
	return InstanceEnumDomain(s.Type, s.Domain)
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
			Name:   TypeEnumDomain(s.Domain).String(),
			Rrtype: dns.TypePTR,
			Class:  dns.ClassINET,
			Ttl:    ttl,
		},
		Ptr: s.InstanceEnumDomain().String(),
	}, true
}

// Validate returns an error if the service is configured incorrectly.
func (s *Service) Validate() error {
	if err := s.Type.Validate(); err != nil {
		return err
	}

	if err := s.Domain.Validate(); err != nil {
		return err
	}

	for n, i := range s.Instances {
		if i.Name != n {
			return fmt.Errorf(
				"service instance '%s' is stored under the  '%s' key",
				string(i.Name), // don't use .String()
				string(n),
			)
		}

		if i.ServiceType != s.Type {
			return fmt.Errorf(
				"service instance '%s' has type '%s', expected '%s'",
				string(i.Name), // don't use .String()
				i.ServiceType,
				s.Type,
			)
		}

		if err := i.Validate(); err != nil {
			return err
		}
	}

	return nil
}
