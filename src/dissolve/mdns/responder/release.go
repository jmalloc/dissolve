package responder

import (
	"context"

	"github.com/jmalloc/dissolve/src/dissolve/names"
)

type release struct {
	name names.FQDN
}

func (c *release) Execute(ctx context.Context, r *Responder) error {
	panic("ni")
}
