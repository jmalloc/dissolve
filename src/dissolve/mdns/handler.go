package mdns

import (
	"context"

	"github.com/miekg/dns"
)

// Handler is an interface for handling DNS questions.
type Handler interface {
	// HandleQuestion handles a single question within a DNS request.
	HandleQuestion(
		ctx context.Context,
		req *Request,
		res *Response,
		q dns.Question,
	) error
}
