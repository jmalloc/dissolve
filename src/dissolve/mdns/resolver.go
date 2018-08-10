package mdns

import (
	"context"
	"net"
	"strings"

	"github.com/jmalloc/dissolve/src/dissolve/resolver"
)

// localResolver is an implementation of resolver.Resolver that resolves
// hostnames in the ".local" domain to this machine's IP addresses.
type localResolver struct {
	resolver.Resolver
}

// NewLocalResolver returns a new resolver that resolves hostnames in the
// ".local" domain to this machine's IP addresses.
func NewLocalResolver(next resolver.Resolver) resolver.Resolver {
	if next == nil {
		next = net.DefaultResolver
	}

	return &localResolver{Resolver: next}
}

// LookupIPAddr looks up host. It returns a slice of that host's IPv4 and
// IPv6 addresses.
func (r *localResolver) LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error) {
	if !strings.HasSuffix(host, ".local.") {
		return r.Resolver.LookupIPAddr(ctx, host)
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var addrs []net.IPAddr

	for _, iface := range ifaces {
		ifaceAddrs, _ := iface.Addrs()

		for _, addr := range ifaceAddrs {
			ip, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			if ip.IP.IsLoopback() {
				continue
			}

			addrs = append(
				addrs,
				net.IPAddr{
					IP: ip.IP,
				},
			)
		}
	}

	return addrs, nil
}
