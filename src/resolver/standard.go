package resolver

import (
	"context"
	"net"
	"sort"
	"strings"
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
	addr, _ = ipToArpa(addr)
	res, err := r.query(ctx, addr, dns.TypePTR)
	if err != nil {
		return
	}

	if res != nil {
		for _, ans := range res.Answer {
			if rec, ok := ans.(*dns.PTR); ok {
				names = append(names, rec.Ptr)
			}
		}
	}

	if len(names) == 0 {
		err = &net.DNSError{
			Err:  "unable to resolve address", // TODO
			Name: addr,
		}
	}
	return
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
	res, err := r.query(ctx, host, dns.TypeCNAME)
	if err != nil {
		return
	}

	if res != nil {
		for _, ans := range res.Answer {
			if rec, ok := ans.(*dns.CNAME); ok {
				cname = rec.Target
				return
			}
		}
	}

	err = &net.DNSError{
		Err:  "unable to resolve address", // TODO
		Name: host,
	}

	return
}

// LookupHost looks up the given host. It returns a slice of that host's
// addresses.
func (r *StandardResolver) LookupHost(ctx context.Context, host string) (addrs []string, err error) {
	res, err := r.query(ctx, host, dns.TypeA) // TODO: IPv6/AAAA
	if err != nil {
		return
	}

	if res != nil {
		for _, ans := range res.Answer {
			if rec, ok := ans.(*dns.A); ok {
				addrs = append(addrs, rec.A.String())
			}
		}
	}

	if len(addrs) == 0 {
		err = &net.DNSError{
			Err:  "unable to resolve address", // TODO
			Name: host,
		}
	}

	return
}

// LookupIPAddr looks up host. It returns a slice of that host's IPv4 and IPv6
// addresses.
func (r *StandardResolver) LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error) {
	panic("not impl")
}

// LookupMX returns the DNS MX records for the given domain name sorted by
// preference.
func (r *StandardResolver) LookupMX(ctx context.Context, name string) (mx []*net.MX, err error) {
	res, err := r.query(ctx, name, dns.TypeMX)
	if err != nil {
		return
	}

	if res != nil {
		for _, ans := range res.Answer {
			if rec, ok := ans.(*dns.MX); ok {
				mx = append(mx, &net.MX{
					Host: rec.Mx,
					Pref: rec.Preference,
				})
			}
		}

		sort.Slice(mx, func(i, j int) bool {
			return mx[i].Pref < mx[j].Pref
		})
	}

	if len(mx) == 0 {
		err = &net.DNSError{
			Err:  "unable to resolve address", // TODO
			Name: name,
		}
	}

	return
}

// LookupNS returns the DNS NS records for the given domain name.
func (r *StandardResolver) LookupNS(ctx context.Context, name string) (ns []*net.NS, err error) {
	res, err := r.query(ctx, name, dns.TypeNS)
	if err != nil {
		return
	}

	if res != nil {
		for _, ans := range res.Answer {
			if rec, ok := ans.(*dns.NS); ok {
				ns = append(ns, &net.NS{
					Host: rec.Ns,
				})
			}
		}
	}

	if len(ns) == 0 {
		err = &net.DNSError{
			Err:  "unable to resolve address", // TODO
			Name: name,
		}
	}

	return
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
func (r *StandardResolver) LookupTXT(ctx context.Context, name string) (txt []string, err error) {
	res, err := r.query(ctx, name, dns.TypeTXT)
	if err != nil {
		return
	}

	if res != nil {
		for _, ans := range res.Answer {
			if rec, ok := ans.(*dns.TXT); ok {
				txt = append(txt, rec.Txt...)
			}
		}
	}

	if len(txt) == 0 {
		err = &net.DNSError{
			Err:  "unable to resolve address", // TODO
			Name: name,
		}
	}

	return
}

func (r *StandardResolver) query(ctx context.Context, n string, t uint16) (res *dns.Msg, err error) {
	cfg := r.Config
	if cfg == nil {
		cfg = DefaultConfig
	}

	req := &dns.Msg{}

	for _, n := range cfg.NameList(n) {
		req.SetQuestion(n, t)

		if r.isMulticast(n) {
			res, err = r.queryMulticast(ctx, req)
		} else {
			res, err = r.queryUnicast(ctx, req)
		}

		if res != nil || err != nil {
			return
		}
	}

	return
}

func (r *StandardResolver) queryUnicast(ctx context.Context, req *dns.Msg) (res *dns.Msg, err error) {
	cfg := r.Config
	if cfg == nil {
		cfg = DefaultConfig
	}

	cli := r.Unicast
	if cli == nil {
		cli = client.DefaultUnicast
	}

	for _, ns := range cfg.Servers {
		ns = net.JoinHostPort(ns, cfg.Port)

		res, err = cli.Query(ctx, req, ns)
		if err != nil {
			return
		} else if res == nil {
			continue // TODO: when can res be nil?
		} else if res.Rcode == dns.RcodeNameError || res.Rcode == dns.RcodeSuccess {
			return
		}
	}

	return nil, nil
}

func (r *StandardResolver) queryMulticast(ctx context.Context, req *dns.Msg) (*dns.Msg, error) {
	return nil, nil
}

func (r *StandardResolver) isMulticast(n string) bool {
	if r.IsMulticast != nil {
		return r.IsMulticast(n)
	}

	return strings.HasSuffix(n, ".local.")
}
