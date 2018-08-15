package responder

import (
	"context"

	"github.com/jmalloc/dissolve/src/dissolve/names"
)

type acquire struct {
	Name   names.FQDN
	Result chan error
}

func (c *acquire) Execute(ctx context.Context, r *Responder) error {
	panic("ni")
	// dnsQ := dns.Question{
	// 	Name:   c.Name.String(),
	// 	Qclass: dns.ClassANY,
	// 	Qtype:  dns.TypeANY,
	// }

	// rawQ := setUnicastResponse(dnsQ)

	// query := newQuery(false, rawQ)

	// var (
	// 	q = Question{
	// 		Question:       dnsQ,
	// 		Query:          query,
	// 		Interface: *r.iface,
	// 	}
	// 	a = Answer{}
	// )

	// err := s.answerer.Answer(ctx, &q, &a)
	// if err != nil {
	// 	return err
	// }

	// panic("ni")
}
