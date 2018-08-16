package dnssd

import (
	"errors"
	"net"
	"time"

	"github.com/jmalloc/dissolve/src/dissolve/names"
	"github.com/miekg/dns"
)

// DefaultTTL is the default TTL for all DNS records.
const DefaultTTL = 120 * time.Second

// InstanceCollection is the map of the unqualified service instance name to the
// instance.
type InstanceCollection map[InstanceName]*Instance

// Add adds an instance to the collection.
func (c InstanceCollection) Add(i *Instance) {
	c[i.Name] = i
}

// Instance is a DNS-SD service instance.
type Instance struct {
	// Name is the instance's unique name.
	Name InstanceName

	// ServiceType is the type of service that this instance is.
	ServiceType ServiceType

	// Domain is the domain under which the instance is advertised.
	Domain names.FQDN

	// TargetHost is the hostname of the service. This is not necessarily in the
	// same domain as the DNS-SD records.
	//
	// If TargetHost is unqualified, it is assumed to be relative to Domain.
	TargetHost names.Name

	// TargetPort is TCP/UDP port that the service instance listens on.
	TargetPort uint16

	// Text contains a set of key/value pairs that are encoded in the instance's
	// TXT record, as per https://tools.ietf.org/html/rfc6763#section-6.3.
	Text Text

	// Priority is the "service priority" value to use for the service's SRV
	// record. It controls which servers are contacted first. Lower values have a
	// higher priority.
	//
	// See https://tools.ietf.org/html/rfc2782.
	Priority uint16

	// Weight is the "service weight" value to use for the service's SRV record.
	// It controls the likelihood that a server will be chosen from a pool of SRV
	// records with the same priority. Higher values are more likely to be chosen.
	//
	// See https://tools.ietf.org/html/rfc2782.
	Weight uint16

	// TTL is the TTL of the instance's DNS records.
	TTL time.Duration
}

// FQDN returns the instance's fully-qualified domain name.
func (i *Instance) FQDN() names.FQDN {
	return i.Name.Join(i.ServiceType).Qualify(i.Domain)
}

// TargetFQDN returns the FQDN of the target host.
func (i *Instance) TargetFQDN() names.FQDN {
	return i.TargetHost.Qualify(i.Domain)
}

// PTR returns the instance's PTR record.
func (i *Instance) PTR() *dns.PTR {
	return &dns.PTR{
		Hdr: dns.RR_Header{
			Name:   InstanceEnumDomain(i.ServiceType, i.Domain).String(),
			Rrtype: dns.TypePTR,
			Class:  dns.ClassINET,
			Ttl:    i.TTLInSeconds(),
		},
		Ptr: i.FQDN().String(),
	}
}

// SRV returns the instance's SRV record.
func (i *Instance) SRV() *dns.SRV {
	return &dns.SRV{
		Hdr: dns.RR_Header{
			Name:   i.FQDN().String(),
			Rrtype: dns.TypeSRV,
			Class:  dns.ClassINET,
			Ttl:    i.TTLInSeconds(),
		},
		Priority: i.Priority,
		Weight:   i.Weight,
		Target:   i.TargetHost.Qualify(i.Domain).String(),
		Port:     i.TargetPort,
	}
}

// TXT returns the instance's TXT record.
func (i *Instance) TXT() *dns.TXT {
	return &dns.TXT{
		Hdr: dns.RR_Header{
			Name:   i.FQDN().String(),
			Rrtype: dns.TypeTXT,
			Class:  dns.ClassINET,
			Ttl:    i.TTLInSeconds(),
		},
		Txt: i.Text.Pairs(),
	}
}

// A returns an instance's A record for the instance.
func (i *Instance) A(ip net.IP) *dns.A {
	return &dns.A{
		Hdr: dns.RR_Header{
			Name:   i.TargetHost.Qualify(i.Domain).String(),
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    i.TTLInSeconds(),
		},
		A: ip,
	}
}

// AAAA returns an instance's AAAA record for the instance.
func (i *Instance) AAAA(ip net.IP) *dns.AAAA {
	return &dns.AAAA{
		Hdr: dns.RR_Header{
			Name:   i.TargetHost.Qualify(i.Domain).String(),
			Rrtype: dns.TypeAAAA,
			Class:  dns.ClassINET,
			Ttl:    i.TTLInSeconds(),
		},
		AAAA: ip,
	}
}

// TTLInSeconds returns the instance's DNS record TTL in seconds.
// If i.TTL is 0, it uses DefaultTTL.
func (i *Instance) TTLInSeconds() uint32 {
	ttl := i.TTL
	if ttl == 0 {
		ttl = DefaultTTL
	}

	return uint32(ttl.Seconds())
}

// Validate returns an error if the instance is configured incorrectly.
func (i *Instance) Validate() error {
	if err := i.Name.Validate(); err != nil {
		return err
	}

	if err := i.ServiceType.Validate(); err != nil {
		return err
	}

	if err := i.Domain.Validate(); err != nil {
		return err
	}

	if err := i.TargetHost.Validate(); err != nil {
		return err
	}

	if i.TargetPort == 0 {
		return errors.New("target port must not be zero")
	}

	return nil
}
