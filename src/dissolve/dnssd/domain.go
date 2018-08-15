package dnssd

import (
	"fmt"

	"github.com/jmalloc/dissolve/src/dissolve/names"
)

// DomainCollection is the map of domain name to domain.
type DomainCollection map[names.FQDN]*Domain

// Domain is a representation of an internet domain name that has DNS-SD service
// instances.
type Domain struct {
	// Name is the fully-qualified name of the domain such as "example.org.".
	Name names.FQDN

	// Services is the set of services within the zone.
	Services ServiceCollection
}

// TypeEnumDomain returns DNS name that is queried perform "service type
// enumeration" for this domain.
//
// See https://tools.ietf.org/html/rfc6763#section-4.
func (d *Domain) TypeEnumDomain() names.FQDN {
	return TypeEnumDomain(d.Name)
}

// SubTypeEnumDomain returns the DNS name that is queried to perform
// "selective instance enumeration" for a specific service sub-type within this
// domain.
//
// See https://tools.ietf.org/html/rfc6763#section-7.1
func (d *Domain) SubTypeEnumDomain(subtype names.Label, service names.UDN) names.FQDN {
	return SubTypeEnumDomain(subtype, service, d.Name)
}

// Validate returns an error if the service is configured incorrectly.
func (d *Domain) Validate() error {
	if err := d.Name.Validate(); err != nil {
		return err
	}

	for t, s := range d.Services {
		if s.Type != t {
			return fmt.Errorf(
				"service '%s' is stored under the  '%s' key",
				s.Type,
				t,
			)
		}

		if err := s.Validate(); err != nil {
			return err
		}
	}

	return nil
}
