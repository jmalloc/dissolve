package responder

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/jmalloc/dissolve/src/dissolve/mdns/transport"
	"github.com/jmalloc/dissolve/src/dissolve/names"

	"github.com/jmalloc/twelf/src/twelf"
	"github.com/miekg/dns"
	"golang.org/x/sync/errgroup"
)

// command is a unit-of-work performed within the server's main loop.
type command interface {
	Execute(ctx context.Context, r *Responder) error
}

// Responder is an implementation of a multicast DNS responder for a single network interface.
type Responder struct {
	answerer    Answerer
	iface       *net.Interface
	disableIPv4 bool
	disableIPv6 bool
	logger      twelf.Logger

	done     chan struct{}
	commands chan command
}

// New returns a new mDNS server.
func New(
	answerer Answerer,
	options ...Option,
) (*Responder, error) {
	r := &Responder{
		answerer: answerer,
		done:     make(chan struct{}),
		commands: make(chan command),
	}

	for _, opt := range options {
		if err := opt(r); err != nil {
			return nil, err
		}
	}

	if r.iface == nil {
		iface, err := internetInterface()
		if err != nil {
			return nil, err
		}
		r.iface = &iface
	}

	if r.logger == nil {
		r.logger = twelf.DefaultLogger
	}

	return r, nil
}

// Probe attempts to "take control" of the given DNS name by probing the
// network to see if any other mDNS responder is already responding for that
// name.
//
// Once acquired, the name is "defended" against other mDNS responders taking
// control of the name. See https://tools.ietf.org/html/rfc6762#section-8.1 for
// information on mDNS probing.
func (r *Responder) Probe(ctx context.Context, name names.FQDN) error {
	ch := make(chan error, 1)

	if err := r.execute(ctx, &acquire{name, ch}); err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-r.done:
		return errors.New("server is no longer running")
	case err := <-ch:
		return err
	}
}

// Release stops the server from "defending" a name that was previously acquired.
func (r *Responder) Release(ctx context.Context, name names.FQDN) error {
	return r.execute(ctx, &release{name})
}

// execute executes a server command immediately and blocks until it is complete.
func (r *Responder) execute(ctx context.Context, c command) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-r.done:
		return errors.New("server is no longer running")
	case r.commands <- c:
		return nil
	}
}

// schedule executes a server command at some point in the future.
func (r *Responder) schedule(ctx context.Context, d time.Duration, c command) {
	go func() {
		if err := sleep(ctx, d); err == nil {
			r.execute(ctx, c)
		}
	}()
}

// Run response to mDNS messages until ctx is canceled or an error occurs.
func (r *Responder) Run(ctx context.Context) error {
	if r.disableIPv4 && r.disableIPv6 {
		return errors.New("both IPv4 and IPv6 are disabled")
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	if !r.disableIPv4 {
		t := &transport.IPv4Transport{
			Logger: r.logger,
		}
		// // r.transports = append(r.transports, t)

		g.Go(func() error {
			return r.receive(ctx, t)
		})
	}

	if !r.disableIPv6 {
		t := &transport.IPv6Transport{
			Logger: r.logger,
		}
		// // r.transports = append(r.transports, t)

		g.Go(func() error {
			return r.receive(ctx, t)
		})
	}

	g.Go(func() error {
		return r.run(ctx)
	})

	err := g.Wait()

	if err == context.Canceled {
		return nil
	}

	return err
}

// run is the server's main loop.
func (r *Responder) run(ctx context.Context) error {
	defer close(r.done)

	// When ready to send its Multicast DNS probe packet(s) the host should
	// first wait for a short random delay time, uniformly distributed in
	// the range 0-250 ms.  This random delay is to guard against the case
	// where several devices are powered on simultaneously, or several
	// devices are connected to an Ethernet hub, which is then powered on,
	// or some other external event happens that might cause a group of
	// hosts to all send synchronized probes.
	if err := sleep(ctx, randT(250*time.Millisecond)); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case c := <-r.commands:
			if err := c.Execute(ctx, r); err != nil {
				return err
			}
		}
	}
}

// receive pipes packets received from t to s.packets
func (r *Responder) receive(ctx context.Context, t transport.Transport) error {
	if err := t.Listen(r.iface); err != nil {
		return err
	}
	defer t.Close()

	go func() {
		<-ctx.Done()
		_ = t.Close() // break out of t.Read() when the context is canceled
	}()

	for {
		in, err := t.Read()
		if err != nil {
			// TODO(jmalloc): check for "closed" error and return ctx.Err() instead
			return err
		}

		m, err := in.Message()

		if err == dns.ErrTruncated {
			// https://tools.ietf.org/html/rfc6762#section-18.5
			//
			// In query messages, if the TC bit is set, it means that additional
			// Known-Answer records may be following shortly.  A responder SHOULD
			// record this fact, and wait for those additional Known-Answer records,
			// before deciding whether to respond.  If the TC bit is clear, it means
			// that the querying host has no additional Known Answers.
			//
			// TODO(jmalloc): This "error" is not actually an error in the case of mDNS.
			// See https://github.com/miekg/dns/issues/423. We attempt to serve the
			// request anyway, without many guarantees as to the validity of the
			// message. We also do not currently support the behavior specified above.
			//
			// Because our DNS responder will not be the only multicast responder on the
			// machine (ie the host OS provides its own) this may not even be possible
			// to implement correctly. See https://tools.ietf.org/html/rfc6762#section-15.2
			// for more information.
			r.logger.DebugString("received mDNS message with non-zero TC flag")
		} else if err != nil {
			r.logger.Log("error parsing mDNS message: %s", err)
			continue
		}

		var c command
		if m.Response {
			c = &handleResponse{in, m}
		} else {
			c = &handleQuery{in, m}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case r.commands <- c:
		}
	}
}
