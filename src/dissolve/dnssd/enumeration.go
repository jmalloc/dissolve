package dnssd

import "github.com/jmalloc/dissolve/src/dissolve/names"

// TypeEnumDomain returns the DNS name that is queried to perform
// "service type enumeration" for a single domain.
//
// See https://tools.ietf.org/html/rfc6763#section-9
func TypeEnumDomain(domain names.FQDN) names.FQDN {
	return names.UDN("_services._dns-sd._udp").Qualify(domain)
}

// SubTypeEnumDomain returns the DNS name that is queried to perform
// "selective instance enumeration" for a specific service sub-type within a
// single domain.
//
// See https://tools.ietf.org/html/rfc6763#section-7.1
func SubTypeEnumDomain(
	subtype names.Label,
	service names.UDN,
	domain names.FQDN,
) names.FQDN {
	return subtype.
		Join(names.Label("_sub")).
		Join(service).
		Qualify(domain)
}

// InstanceEnumDomain returns the DNS name that is queried to perform
// "service instance enumeration" (aka "browse") on a service within a given
// domain.
//
// See https://tools.ietf.org/html/rfc6763#section-4.
func InstanceEnumDomain(t ServiceType, domain names.FQDN) names.FQDN {
	return t.Qualify(domain)
}
