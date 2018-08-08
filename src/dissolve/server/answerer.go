package server

import (
	"context"
	"net"

	"github.com/jmalloc/dissolve/src/dissolve/resolver"
	"github.com/miekg/dns"
)

// Answerer is an interface that provides answers to DNS questions.
type Answerer interface {
	// Answer populates m with the answer to q.
	//
	// r is the a resolver that should be used by the answerer if it needs to make
	// DNS queries. s is the "source" address of the DNS query being answered.
	Answer(
		ctx context.Context,
		r resolver.Resolver,
		s net.Addr,
		q dns.Question,
		m *dns.Msg,
	) error
}
