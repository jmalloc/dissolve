package interceptor

import (
	"context"
	"net"

	"github.com/jmalloc/dissolve/src/dissolve/names"
)

var defaultMulticastDomains = []names.FQDN{
	names.FQDN("local."),
}

// Setup installs the interceptor into net.DefaultResolver.
func Setup(domains ...names.FQDN) {
	d := &Dialer{
		MulticastDomains: domains,
		UnicastDial:      net.DefaultResolver.Dial,
	}

	net.DefaultResolver.Dial = d.Dial
}

// Dialer provides a Dial() method for use with net.Resolver.Dial.
//
// The native Go resolver does not support mDNS. This dialer returns a proxy
// connection that intercepts DNS queries that contain questions about mDNS
// names, and sends them via multicast instead of to the unicast DNS server.
type Dialer struct {
	// MulticastDomains is a set of fully-qualified names that should be queried
	// via multicast. If it is empty, "local." is used.
	MulticastDomains []names.FQDN

	// UnicastDial is the underlying dialer used to establish a connection to
	// the unicast DNS server. It defaults to net.Dialer.DialContext().
	UnicastDial func(ctx context.Context, net, addr string) (net.Conn, error)
}

// Dial returns a net.Conn that acts as a proxy to either a conventional unicast
// DNS server, or multicast servers on the local network(s).
func (d *Dialer) Dial(
	ctx context.Context,
	network, address string,
) (net.Conn, error) {
	cli, svr := net.Pipe()

	domains := d.MulticastDomains
	if len(domains) == 0 {
		domains = defaultMulticastDomains
	}

	dial := d.UnicastDial
	if dial == nil {
		d := &net.Dialer{}
		dial = d.DialContext
	}

	i := &interceptor{
		ctx:  ctx,
		net:  network,
		addr: address,
		dial: dial,

		conn:    svr,
		domains: domains,
	}

	go i.Run()

	return cli, nil
}
