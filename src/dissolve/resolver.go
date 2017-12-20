package dissolve

import (
	"context"
	"net"
)

// Resolver is an interface for making DNS queries.
// It is compatible with the net.Resolver struct.
type Resolver interface {
	// LookupAddr performs a reverse lookup for the given address, returning a
	// list of names mapping to that address.
	LookupAddr(ctx context.Context, addr string) (names []string, err error)

	// LookupCNAME returns the canonical name for the given host. Callers that
	// do not care about the canonical name can call LookupHost or LookupIP
	// directly; both take care of resolving the canonical name as part of the
	// lookup.
	//
	// A canonical name is the final name after following zero or more CNAME
	// records. LookupCNAME does not return an error if host does not contain
	// DNS "CNAME" records, as long as host resolves to address records.
	LookupCNAME(ctx context.Context, host string) (cname string, err error)

	// LookupHost looks up the given host using the local resolver. It returns a
	// slice of that host's addresses.
	LookupHost(ctx context.Context, host string) (addrs []string, err error)

	// LookupIPAddr looks up host using the local resolver. It returns a slice of
	// that host's IPv4 and IPv6 addresses.
	LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error)

	// LookupMX returns the DNS MX records for the given domain name sorted by
	// preference.
	LookupMX(ctx context.Context, name string) ([]*net.MX, error)

	// LookupNS returns the DNS NS records for the given domain name.
	LookupNS(ctx context.Context, name string) ([]*net.NS, error)

	// LookupPort looks up the port for the given network and service.
	LookupPort(ctx context.Context, network, service string) (port int, err error)

	// LookupSRV tries to resolve an SRV query of the given service, protocol,
	// and domain name. The proto is "tcp" or "udp". The returned records are
	// sorted by priority and randomized by weight within a priority.
	//
	// LookupSRV constructs the DNS name to look up following RFC 2782. That is,
	// it looks up _service._proto.name. To accommodate services publishing SRV
	// records under non-standard names, if both service and proto are empty
	// strings, LookupSRV looks up name directly.
	LookupSRV(ctx context.Context, service, proto, name string) (cname string, addrs []*net.SRV, err error)

	// LookupTXT returns the DNS TXT records for the given domain name.
	LookupTXT(ctx context.Context, name string) ([]string, error)
}
