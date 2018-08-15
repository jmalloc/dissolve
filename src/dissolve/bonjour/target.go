package bonjour

import (
	"context"
	"net"

	"github.com/jmalloc/dissolve/src/dissolve/names"

	"github.com/jmalloc/dissolve/src/dissolve/dnssd"
	"github.com/jmalloc/dissolve/src/dissolve/mdns/responder"
	"github.com/jmalloc/dissolve/src/dissolve/resolver"
	"github.com/miekg/dns"
)

// targetAnswerer is a responder.Answerer that provides DNS answers for the
// target hostname of a DNS-SD service instance.
type targetAnswerer struct {
	Resolver resolver.Resolver
	Instance *dnssd.Instance
}

func (an *targetAnswerer) Answer(
	ctx context.Context,
	q *responder.Question,
	a *responder.Answer,
) error {
	switch q.Qtype {
	case dns.TypeA, dns.TypeAAAA, dns.TypeANY:
	default:
		return nil
	}

	v4, v6, err := addressRecords(ctx, an.Resolver, q.Interface, an.Instance)
	if err != nil {
		return err
	}

	switch q.Qtype {
	case dns.TypeANY:
		a.Unique.Answer(v4...)
		a.Unique.Answer(v6...)

	case dns.TypeA:
		a.Unique.Answer(v4...)
		a.Unique.Additional(v6...)

	case dns.TypeAAAA:
		a.Unique.Answer(v6...)
		a.Unique.Additional(v4...)
	}

	return nil
}

// addressRecords returns the A and AAAA records for the given instance.
func addressRecords(
	ctx context.Context,
	r resolver.Resolver,
	f net.Interface,
	i *dnssd.Instance,
) (
	[]dns.RR,
	[]dns.RR,
	error,
) {
	var (
		addresses []net.IP
		err       error
	)

	if i.TargetHost.IsQualified() {
		addresses, err = resolveRemoteAddrs(ctx, r, i.TargetHost)
	} else {
		addresses, err = resolveLocalAddrs(f)
	}

	if err != nil {
		return nil, nil, err
	}

	var v4, v6 []dns.RR

	for _, ip := range addresses {
		if ip.To4() != nil {
			v4 = append(v4, i.A(ip))
		} else {
			v6 = append(v6, i.AAAA(ip))
		}
	}

	return v4, v6, nil
}

// resolveRemoteAddrs uses r to find the IP addresses of name.
func resolveRemoteAddrs(
	ctx context.Context,
	r resolver.Resolver,
	name names.Name,
) ([]net.IP, error) {
	if r == nil {
		r = net.DefaultResolver
	}

	addrs, err := r.LookupIPAddr(ctx, name.String())
	if err != nil {
		return nil, err
	}

	addresses := make([]net.IP, len(addrs))
	for i, addr := range addrs {
		addresses[i] = addr.IP
	}

	return addresses, nil
}

// resolveLocalAddrs returns the IP addresses of the given interface.
func resolveLocalAddrs(
	f net.Interface,
) ([]net.IP, error) {
	addrs, err := f.Addrs()
	if err != nil {
		return nil, err
	}

	addresses := make([]net.IP, 0, len(addrs))
	for _, addr := range addrs {
		if x, ok := addr.(*net.IPNet); ok {
			addresses = append(addresses, x.IP)
		}
	}

	return addresses, nil
}
