package dnssd

import (
	"context"
	"net"
	"strconv"
	"sync"

	"github.com/jmalloc/dissolve/src/dissolve/mdns/transport"
	"github.com/jmalloc/twelf/src/twelf"
	"golang.org/x/sync/errgroup"

	"github.com/miekg/dns"

	"github.com/jmalloc/dissolve/src/dissolve/names"
)

// Resolver is a specialised DNS resolver that provides a synchronous interface
// for locating DNS-SD service instances.
type Resolver struct {
	Client           *dns.Client
	UnicastConfig    *dns.ClientConfig
	MulticastConfig  *dns.ClientConfig
	MulticastDomains []names.FQDN
	Logger           twelf.Logger
}

// NewResolver returns a new DNS-SD resolver.
func NewResolver() (*Resolver, error) {
	conf, err := dns.ClientConfigFromFile("/etc/resolv.conf")
	if err != nil {
		return nil, err
	}

	return &Resolver{
		&dns.Client{},
		conf,
		&dns.ClientConfig{
			Servers: []string{
				transport.IPv4Group.String(),
				transport.IPv6Group.String(),
			},
			Port:    strconv.Itoa(transport.Port),
			Timeout: 2,
		},
		[]names.FQDN{"local."},
		twelf.DiscardLogger{},
	}, nil
}

func (r *Resolver) EnumerateInstances(
	ctx context.Context,
	t ServiceType,
	d names.FQDN,
) (*Service, error) {
	res, err := r.query(
		ctx,
		InstanceEnumDomain(t, d),
		dns.TypePTR,
	)
	if err != nil {
		return nil, err
	}

	s := &Service{
		Type:      t,
		Domain:    d,
		Instances: InstanceCollection{},
	}

	if res == nil {
		return s, nil
	}

	if res.Rcode != dns.RcodeSuccess {
		return s, nil
	}

	var m sync.Mutex
	g, ctx := errgroup.WithContext(ctx)

	for _, rr := range res.Answer {
		if ptr, ok := rr.(*dns.PTR); ok {
			g.Go(func() error {
				i, ok := r.queryInstance(ctx, t, d, ptr, res)
				if ok {
					m.Lock()
					defer m.Unlock()

					s.Instances.Add(i)
				}

				return nil
			})
		}
	}

	return s, g.Wait()
}

func (r *Resolver) queryInstance(
	ctx context.Context,
	t ServiceType,
	d names.FQDN,
	ptr *dns.PTR,
	res *dns.Msg,
) (*Instance, bool) {
	fqdn := names.FQDN(ptr.Ptr)
	if err := fqdn.Validate(); err != nil {
		r.Logger.Debug("unable to query service instance: ", err)
		return nil, false
	}

	srv, txt := extractRecords(ptr.Ptr, res.Extra)
	qtype := dns.TypeNone

	if srv == nil && txt == nil {
		qtype = dns.TypeANY
		r.Logger.Debug("querying SRV and TXT records for '%s'", fqdn)
	} else if srv == nil {
		qtype = dns.TypeSRV
		r.Logger.Debug("querying SRV record for '%s'", fqdn)
	} else if txt == nil {
		qtype = dns.TypeTXT
		r.Logger.Debug("querying TXT record for '%s'", fqdn)
	}

	// perform an additional query to fetch the DNS-SD SRV or TXT records
	if qtype != dns.TypeNone {
		res, err := r.query(ctx, fqdn, qtype)
		if err != nil || res == nil || res.Rcode != dns.RcodeSuccess {
			return nil, false
		}

		s, t := extractRecords(ptr.Ptr, res.Answer)
		if s != nil {
			srv = s
		}
		if t != nil {
			txt = t
		}

		if srv == nil && txt == nil {
			r.Logger.Debug("could not find SRV and TXT records for '%s'", fqdn)
			return nil, false
		} else if srv == nil {
			r.Logger.Debug("could not find SRV record for '%s'", fqdn)
			return nil, false
		} else if txt == nil {
			r.Logger.Debug("could not find TXT record for '%s'", fqdn)
			return nil, false
		}
	}

	in, _ := SplitInstanceName(fqdn)

	th, err := names.Parse(srv.Target)
	if err != nil {
		r.Logger.Debug("could not parse target host for '%s': %s", fqdn, err)
		return nil, false
	}

	tm, err := ParseTextPairs(txt.Txt)
	if err != nil {
		return nil, false
	}

	i := &Instance{
		Name:        in,
		ServiceType: t,
		Domain:      d,
		TargetHost:  th,
		TargetPort:  srv.Port,
		Text:        tm,
	}

	return i, true
}

func (r *Resolver) query(
	ctx context.Context,
	name names.FQDN,
	qtype uint16,
) (*dns.Msg, error) {
	var (
		conf  *dns.ClientConfig
		query *dns.Msg
	)

	if r.isMulticast(name) {
		conf = r.MulticastConfig
		panic("not implemented")
	} else {
		conf = r.UnicastConfig
		query = &dns.Msg{}
		query.SetQuestion(name.String(), qtype)
	}

	if len(conf.Servers) == 0 {
		return nil, nil
	}

	// create a cancelable context so we can abort queries to other services
	// once we get an authoratative response
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	// result is a channel that stores the first result authoratative result
	result := make(chan *dns.Msg, 1)

	for _, s := range conf.Servers {
		s := s // capture loop variable

		// query each server in a separate goroutine
		g.Go(func() error {
			addr := net.JoinHostPort(s, conf.Port)

			res, _, err := r.Client.ExchangeContext(ctx, query, addr)
			if err != nil {
				// this request was aborted, likely because some a result was
				// found on some other DNS server
				if err == context.Canceled {
					return nil
				}

				return err
			}

			switch res.Rcode {
			case dns.RcodeSuccess, dns.RcodeNameError:
				select {
				case result <- res:
					// if we get an authorative answer, and we're the "winning"
					// result, cancel all other goroutines
					cancel()
				default:
				}
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		r.Logger.Debug("unable to query '%s' (qtype: %d): %s", name, qtype, err)
		return nil, err
	}

	close(result)
	res := <-result

	if res == nil {
		r.Logger.Debug("no result received for query '%s' (qtype: %d)", name, qtype)
	} else if res.Rcode != dns.RcodeSuccess {
		r.Logger.Debug("non-success result (%d) received for query '%s' (qtype: %d)", name, res.Rcode, qtype)
	}

	return res, nil
}

// isMulticast returns true if d is a domain that should be queried via mDNS.
func (r *Resolver) isMulticast(d names.FQDN) bool {
	for _, md := range r.MulticastDomains {
		if d.IsWithin(md) {
			return true
		}
	}

	return false
}

// extractRecords returns the SRV and TXT records for the given name from a set
// of records.
func extractRecords(n string, records []dns.RR) (srv *dns.SRV, txt *dns.TXT) {
	for _, rr := range records {
		if rr.Header().Name == n {
			switch v := rr.(type) {
			case *dns.SRV:
				srv = v
			case *dns.TXT:
				txt = v
			}
		}
	}

	return
}
