package resolver

import (
	"context"
	"net"
	"time"

	"github.com/jmalloc/dissolve/src/client"
	"github.com/miekg/dns"
)

// StandardResolver is Dissolve's standard Resolver implementation.
//
// It is a drop-in replacement for Go's net.Resolver that supports both unicast
// and multicast DNS, with changes that make it suitable for use with RFC-6763
// DNS-SD (https://tools.ietf.org/html/rfc6763), one of the key components of
// zeroconf (bonjour).
type StandardResolver struct {
	// Unicast is the client used to perform unicast DNS queries. If it is nil,
	// client.DefaultUnicast is used.
	Unicast client.Unicast

	// MulticastClient is the client used perform multicast DNS queries. If it
	// is nil, client.DefaultMulticast is used.
	Multicast client.Multicast

	// MulticastWait is the minimum time to wait for responses to multicast DNS
	// queries if the request's context does not specific a wait time. If it is
	// zero, DefaultMulticastWait is used.
	//
	// The actual "threshold" time is found using ResolveMulticastWait(ctx, MulticastWait).
	MulticastWait time.Duration

	// IsMulticast is a predicate function used to test if a FQDN should be
	// queried via multicast DNS.  If it is nil, any FQDN ending in ".local." is
	// queried via multicast.
	IsMulticast func(string) bool

	// Config defines the unicast nameservers and other information used to
	// perform queries. If it is nil, DefaultConfig is used.
	Config *dns.ClientConfig
}

// LookupAddr performs a reverse lookup for the given address, returning a
// list of names mapping to that address.
func (r *StandardResolver) LookupAddr(ctx context.Context, addr string) (names []string, err error) {
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
func (r *StandardResolver) LookupCNAME(ctx context.Context, host string) (cname string, err error) {
	panic("not impl")
}

// LookupHost looks up the given host using the local resolver. It returns a
// slice of that host's addresses.
func (r *StandardResolver) LookupHost(ctx context.Context, host string) (addrs []string, err error) {
	panic("not impl")
}

// LookupIPAddr looks up host using the local resolver. It returns a slice of
// that host's IPv4 and IPv6 addresses.
func (r *StandardResolver) LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error) {
	panic("not impl")
}

// LookupMX returns the DNS MX records for the given domain name sorted by
// preference.
func (r *StandardResolver) LookupMX(ctx context.Context, name string) ([]*net.MX, error) {
	panic("not impl")
}

// LookupNS returns the DNS NS records for the given domain name.
func (r *StandardResolver) LookupNS(ctx context.Context, name string) ([]*net.NS, error) {
	panic("not impl")
}

// LookupPort looks up the port for the given network and service.
func (r *StandardResolver) LookupPort(ctx context.Context, network, service string) (port int, err error) {
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
func (r *StandardResolver) LookupSRV(ctx context.Context, service, proto, name string) (cname string, addrs []*net.SRV, err error) {
	panic("not impl")
}

// LookupTXT returns the DNS TXT records for the given domain name.
func (r *StandardResolver) LookupTXT(ctx context.Context, name string) ([]string, error) {
	panic("not impl")
}
