package dnssd

import (
	"github.com/jmalloc/dissolve/src/dissolve/names"
)

// ServiceTypeEnumerationDomain returns the DNS name that is queried to perform
// "service type enumeration" for a given domain.
//
// See https://tools.ietf.org/html/rfc6763#section-9
func ServiceTypeEnumerationDomain(domain names.FQDN) names.FQDN {
	return serviceTypeEnum.Qualify(domain)
}

var serviceTypeEnum = names.MustParseRel("_services._dns-sd._udp")

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

// ServiceTypeEnumerationDomain returns DNS name that is queried perform
// "service type enumeration" for this domain.
//
// See https://tools.ietf.org/html/rfc6763#section-4.
func (d *Domain) ServiceTypeEnumerationDomain() names.FQDN {
	return ServiceTypeEnumerationDomain(d.Name)
}

// Validate returns an error if the service is configured incorrectly.
func (d *Domain) Validate() error {
	return d.Name.Validate()
}
