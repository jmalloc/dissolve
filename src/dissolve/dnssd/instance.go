package dnssd

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/jmalloc/dissolve/src/dissolve/names"
	"github.com/miekg/dns"
)

// InstanceName is a fully-qualified instance name.
type InstanceName struct {
	Name    names.Host
	Service names.Rel
	Domain  names.FQDN
}

// NewInstanceName returns a new fully-qualified instance name from its
// name, service and domain components.
func NewInstanceName(n, s, d string) (InstanceName, error) {
	v := InstanceName{
		names.Host(n),
		names.Rel(n),
		names.FQDN(n),
	}

	return v, v.Validate()
}

// Validate returns an error if the name is malformed.
func (n InstanceName) Validate() error {
	if err := n.Name.Validate(); err != nil {
		return err
	}

	if err := n.Service.Validate(); err != nil {
		return err
	}

	return n.Domain.Validate()
}

// FQDN returns the FQDN of the instance name.
func (n InstanceName) FQDN() names.FQDN {
	return n.Name.Qualify(
		n.Service.Qualify(n.Domain),
	)
}

// InstanceCollection is the map of the unqualified service instance name to the
// instance.
type InstanceCollection map[names.Host]*Instance

// DefaultTTL is the default TTL for all DNS records.
const DefaultTTL = 120 * time.Second

// Instance is a DNS-SD service instance.
type Instance struct {
	InstanceName

	// TargetHost is the fully-qualified hostname of the service. This is not
	// necessarily in the same domain under which discovery is performed.
	TargetHost names.FQDN

	// TargetPort is TCP/UDP port that the service instance listens on.
	TargetPort uint16

	// Text contains a set of key/value pairs that are encoded in the instance's
	// TXT record, as per https://tools.ietf.org/html/rfc6763#section-6.3.
	Text map[string]string

	// TTL is the TTL of the instance's DNS records.
	TTL time.Duration
}

// NewInstance returns a new service instance.
func NewInstance(
	name, service, domain string,
	host string, port uint16,
) (*Instance, error) {
	i := &Instance{
		InstanceName: InstanceName{
			names.Host(name),
			names.Rel(service),
			names.FQDN(domain),
		},
		TargetHost: names.FQDN(host),
		TargetPort: port,
	}

	if err := i.Validate(); err != nil {
		return nil, err
	}

	return i, nil
}

// PTR returns the instance's PTR record.
func (i *Instance) PTR() *dns.PTR {
	return &dns.PTR{
		Hdr: dns.RR_Header{
			Name:   InstanceEnumerationDomain(i.Service, i.Domain).DNSString(),
			Rrtype: dns.TypePTR,
			Class:  dns.ClassINET,
			Ttl:    i.TTLInSeconds(),
		},
		Ptr: i.InstanceName.FQDN().DNSString(),
	}
}

// SRV returns the instance's SRV record.
func (i *Instance) SRV() *dns.SRV {
	return &dns.SRV{
		Hdr: dns.RR_Header{
			Name:   i.InstanceName.FQDN().DNSString(),
			Rrtype: dns.TypeSRV,
			Class:  dns.ClassINET,
			Ttl:    i.TTLInSeconds(),
		},
		Priority: 10, // TODO(jmalloc): support priority and weight
		Weight:   1,
		Target:   i.TargetHost.DNSString(),
		Port:     i.TargetPort,
	}
}

// TXT returns the instance's TXT record.
func (i *Instance) TXT() *dns.TXT {
	r := &dns.TXT{
		Hdr: dns.RR_Header{
			Name:   i.InstanceName.FQDN().DNSString(),
			Rrtype: dns.TypeTXT,
			Class:  dns.ClassINET,
			Ttl:    i.TTLInSeconds(),
		},
	}

	for k, v := range i.Text {
		r.Txt = append(
			r.Txt,
			fmt.Sprintf("%s=%s", k, v),
		)
	}

	return r
}

// A returns an instance's A record for the instance.
func (i *Instance) A(ip net.IP) *dns.A {
	return &dns.A{
		Hdr: dns.RR_Header{
			Name:   i.TargetHost.DNSString(),
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
			Name:   i.TargetHost.DNSString(),
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

	if err := i.Service.Validate(); err != nil {
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
