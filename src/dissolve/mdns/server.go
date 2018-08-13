package mdns

import (
	"context"
	"errors"
	"net"
	"sync"

	"github.com/jmalloc/dissolve/src/dissolve/mdns/transport"

	"github.com/jmalloc/twelf/src/twelf"
	"github.com/miekg/dns"
	"golang.org/x/sync/errgroup"
)

// Server is a multicast DNS (mDNS) server.
type Server struct {
	answerer    Answerer
	ifaces      []net.Interface
	disableIPv4 bool
	disableIPv6 bool
	logger      twelf.Logger
}

// ServerOption is a function that applies an option to a server created by
// NewServer().
type ServerOption func(*Server) error

// UseLogger returns a server option that sets the logger used by the server.
func UseLogger(l twelf.Logger) ServerOption {
	return func(s *Server) error {
		s.logger = l
		return nil
	}
}

// UseInterfaces returns a server option that sets the network interfaces on
// which the server listens for mDNS messages.
func UseInterfaces(ifaces []net.Interface) ServerOption {
	return func(s *Server) error {
		s.ifaces = ifaces
		return nil
	}
}

// DisableIPv4 is a server option that prevents the server from listening for
// IPv4 messages.
func DisableIPv4(s *Server) error {
	s.disableIPv4 = true
	return nil
}

// DisableIPv6 is a server option that prevents the server from listening for
// IPv6 messages.
func DisableIPv6(s *Server) error {
	s.disableIPv6 = true
	return nil
}

// NewServer returns a new mDNS server.
func NewServer(
	answerer Answerer,
	options ...ServerOption,
) (*Server, error) {
	s := &Server{
		answerer: answerer,
	}

	for _, opt := range options {
		if err := opt(s); err != nil {
			return nil, err
		}
	}

	if len(s.ifaces) == 0 {
		ifaces, err := multicastInterfaces()
		if err != nil {
			return nil, err
		}

		s.ifaces = ifaces
	}

	if s.logger == nil {
		s.logger = twelf.DefaultLogger
	}

	return s, nil
}

// Acquire attempts to "take control" of the given DNS name by probing the
// network to see if any other mDNS responder is already responding for that
// name.
//
// Once acquired, the name is "defended" against other mDNS responders taking
// control of the name. See https://tools.ietf.org/html/rfc6762#section-8.1 for
// information on mDNS probing.
func (s *Server) Acquire(ctx context.Context, name string) error {
	panic("ni")
}

// Release stops the server from "defending" a name that was previously acquired.
func (s *Server) Release(ctx context.Context, name string) error {
	panic("ni")
}

// Run response to mDNS messages until ctx is canceled or an error occurs.
func (s *Server) Run(ctx context.Context) error {
	if s.disableIPv4 && s.disableIPv6 {
		return errors.New("both IPv4 and IPv6 are disabled")
	}

	g, ctx := errgroup.WithContext(ctx)

	if !s.disableIPv4 {
		g.Go(func() error {
			return s.receive(
				ctx,
				&transport.IPv4Transport{
					Interfaces: s.ifaces,
					Logger:     s.logger,
				},
			)
		})
	}

	if !s.disableIPv6 {
		g.Go(func() error {
			return s.receive(
				ctx,
				&transport.IPv6Transport{
					Interfaces: s.ifaces,
					Logger:     s.logger,
				},
			)
		})
	}

	err := g.Wait()

	if err == context.Canceled {
		return nil
	}

	return err
}

// handle handles a DNS message in a UDP packet.
func (s *Server) handle(ctx context.Context, in *transport.InboundPacket) {
	defer in.Close()

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
		s.logger.DebugString("received mDNS message with non-zero TC flag")
	} else if err != nil {
		s.logger.Log("error parsing mDNS message: %s", err)
		return
	}

	if m.Response {
		err = s.handleResponse(ctx, in, m)
	} else {
		err = s.handleQuery(ctx, in, m)
	}

	if err != nil {
		s.logger.Log("error handling mDNS message: %s", err)
	}
}

func (s *Server) handleQuery(
	ctx context.Context,
	in *transport.InboundPacket,
	query *dns.Msg,
) error {
	err := validateQuery(query)
	if err != nil {
		return err
	}

	var (
		legacy = in.Source.IsLegacy()
		uRes   = newResponse(query, true)
		mRes   = newResponse(query, false)
	)

	for _, rawQ := range query.Question {
		unicast, dnsQ := wantsUnicastResponse(rawQ)

		var (
			q = Question{
				Question: dnsQ,
				Query:    query,
				Source:   in.Source,
			}
			a = Answer{}
		)

		err = s.answerer.Answer(ctx, &q, &a)
		if err != nil {
			return err
		}

		if !a.Unique.IsEmpty() {
			// TODO(jmalloc): probe/announce uniquely-scoped records before
			// providing answers to them.
		}

		if unicast || legacy {
			a.appendToMessage(uRes)
		} else {
			a.appendToMessage(mRes)
		}
	}

	_, err = transport.SendUnicastResponse(in, uRes)
	if err != nil {
		return err
	}

	_, err = transport.SendMulticastResponse(in, mRes)
	if err != nil {
		return err
	}

	// this is a no-op unless compiled with the 'debug' build tag
	dumpRequestResponse(in, query, uRes, mRes)

	return nil
}

func (s *Server) handleResponse(
	ctx context.Context,
	in *transport.InboundPacket,
	res *dns.Msg,
) error {
	// TODO(jmalloc): we need to "defend" our records here
	// https://tools.ietf.org/html/rfc6762#section-9
	return nil
}

// receive starts goroutines to handle each DNS message received via t.
func (s *Server) receive(ctx context.Context, t transport.Transport) error {
	if err := t.Listen(); err != nil {
		return err
	}
	defer t.Close()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		<-ctx.Done()
		cancel()      // cancel calls to s.handle()
		_ = t.Close() // break out of t.Read() when the context is canceled
	}()

	var g sync.WaitGroup
	defer g.Wait()

	for {
		in, err := t.Read()
		if err != nil {
			// TODO(jmalloc): check for "closed" error and return ctx.Err() instead
			return err
		}

		g.Add(1)
		go func() {
			defer g.Done()
			s.handle(ctx, in)
		}()
	}
}
