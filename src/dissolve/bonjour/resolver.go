package bonjour

import (
	"context"
	"net"
	"strings"
	"sync"

	"github.com/jmalloc/dissolve/src/dissolve/resolver"
)

// LocalResolver is an implementation of resolver.Resolver that that resolves IP
// address lookups on the .local domain to this machine.
type LocalResolver struct {
	resolver.Resolver
	Source net.Addr

	addrs []net.IPAddr
	once  sync.Once
}

// LookupIPAddr looks up host. It returns a slice of that host's IPv4 and
// IPv6 addresses.
func (r *LocalResolver) LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error) {
	if !strings.HasSuffix(host, ".local.") {
		return r.Resolver.LookupIPAddr(ctx, host)
	}

	var err error

	r.once.Do(func() {
		var con net.Conn
		con, err = net.Dial("udp", r.Source.String())
		if err != nil {
			return
		}
		defer con.Close()

		addr := con.LocalAddr().(*net.UDPAddr)
		r.addrs = []net.IPAddr{
			{
				IP:   addr.IP,
				Zone: addr.Zone,
			},
		}
	})

	return r.addrs, err
}
