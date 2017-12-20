package dissolve

import (
	"context"
	"net"

	"github.com/miekg/dns"
)

// UnicastResolver is an implementation of Resolver that uses conventional
// "unicast" DNS queries.
//
// It is a drop-in replacement for net.Resolver, with changes that make it
// suitable for RFC-6763 DNS-SD (https://tools.ietf.org/html/rfc6763#section-6.8),
// one of the key components of zerconf.
type UnicastResolver struct {
	// Client is the underlying DNS client used to perform the queries.
	Client *dns.Client

	// Config defines the nameservers and other information used to perform
	// queries.
	Config *dns.ClientConfig
}

// LookupAddr performs a reverse lookup for the given address, returning a
// list of names mapping to that address.
func (r *UnicastResolver) LookupAddr(ctx context.Context, addr string) (names []string, err error) {
	panic("not impl")
}

// LookupCNAME returns the canonical name for the given host. Callers that
// do not care about the canonical name can call LookupHost or LookupIP
// directly; both take care of resolving the canonical name as part of the
// lookup.
//
// A canonical name is the final name after following zero or more CNAME
// records. LookupCNAME does not return an error if host does not contain
// DNS "CNAME" records, as long as host resolves to address records.
func (r *UnicastResolver) LookupCNAME(ctx context.Context, host string) (cname string, err error) {
	panic("not impl")
}

// LookupHost looks up the given host using the local resolver. It returns a
// slice of that host's addresses.
func (r *UnicastResolver) LookupHost(ctx context.Context, host string) (addrs []string, err error) {
	panic("not impl")
}

// LookupIPAddr looks up host using the local resolver. It returns a slice of
// that host's IPv4 and IPv6 addresses.
func (r *UnicastResolver) LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error) {
	panic("not impl")
}

// LookupMX returns the DNS MX records for the given domain name sorted by
// preference.
func (r *UnicastResolver) LookupMX(ctx context.Context, name string) ([]*net.MX, error) {
	panic("not impl")
}

// LookupNS returns the DNS NS records for the given domain name.
func (r *UnicastResolver) LookupNS(ctx context.Context, name string) ([]*net.NS, error) {
	panic("not impl")
}

// LookupPort looks up the port for the given network and service.
func (r *UnicastResolver) LookupPort(ctx context.Context, network, service string) (port int, err error) {
	panic("not impl")
}

// LookupSRV tries to resolve an SRV query of the given service, protocol,
// and domain name. The proto is "tcp" or "udp". The returned records are
// sorted by priority and randomized by weight within a priority.
//
// LookupSRV constructs the DNS name to look up following RFC 2782. That is,
// it looks up _service._proto.name. To accommodate services publishing SRV
// records under non-standard names, if both service and proto are empty
// strings, LookupSRV looks up name directly.
func (r *UnicastResolver) LookupSRV(ctx context.Context, service, proto, name string) (cname string, addrs []*net.SRV, err error) {
	panic("not impl")
}

// LookupTXT returns the DNS TXT records for the given domain name.
func (r *UnicastResolver) LookupTXT(ctx context.Context, name string) ([]string, error) {
	panic("not impl")
}
