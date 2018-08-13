package main

import (
	"context"
	"log"
	"net"

	"github.com/jmalloc/dissolve/src/dissolve/mdns"
	"github.com/jmalloc/twelf/src/twelf"
	"github.com/miekg/dns"
)

type answerer struct {
}

func (answerer) Answer(ctx context.Context, q *mdns.Question, a *mdns.Answer) error {
	if q.Name == "foo.bar.local." {
		a.Unique.Answer(&dns.A{
			Hdr: dns.RR_Header{
				Name:   q.Name,
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    120,
			},
			A: net.ParseIP("192.168.60.36"),
		})
	}

	return nil
}

func main() {
	svr, err := mdns.NewServer(
		answerer{},
		mdns.UseLogger(twelf.DebugLogger),
	)
	if err != nil {
		log.Fatal(err)
	}

	err = svr.Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}
