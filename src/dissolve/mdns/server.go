package mdns

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/jmalloc/twelf/src/twelf"
	"github.com/miekg/dns"
	"golang.org/x/sync/errgroup"
)

var test_mutex sync.Mutex

// Server is a multicast DNS (mDNS) server.
type Server struct {
	// Handler is used to find answers to DNS questions. It must not be nil.
	Handler Handler

	// Interfaces is the set of network interfaces that the server listens on.
	// If it is empty, the server listens on all multicast capable interfaces.
	Interfaces []net.Interface

	// If true, DisableIPv4 prevents the server from listening for IPv4 requests.
	DisableIPv4 bool

	// If true, DisableIPv6 prevents the server from listening for IPv6 requests.
	DisableIPv6 bool

	// Logger is the target for log messages.
	// twelf.DefaultLogger is used if it is nil.
	Logger twelf.Logger
}

// Run answers mDNS requests until ctx is canceled or an error occurs.
func (s *Server) Run(ctx context.Context) error {
	if s.Handler == nil {
		return errors.New("server must have an answerer")
	}

	g, ctx := errgroup.WithContext(ctx)

	ifaces := s.Interfaces
	if len(ifaces) == 0 {
		var err error
		ifaces, err = multicastInterfaces()
		if err != nil {
			return err
		}
	}

	logger := s.Logger
	if logger == nil {
		logger = twelf.DefaultLogger
	}

	if s.DisableIPv4 && s.DisableIPv6 {
		return errors.New("both IPv4 and IPv6 are disabled")
	}

	if !s.DisableIPv4 {
		g.Go(func() error {
			conn, err := listen4()
			if err != nil {
				logger.Log("unable to listen for IPv4 mDNS requests: %s", err)
				return err
			}
			defer conn.Close()

			ms := &server{
				conn,
				s.Handler,
				ifaces,
				logger,
			}

			return ms.Run(ctx)
		})
	}

	if !s.DisableIPv6 {
		g.Go(func() error {
			conn, err := listen6()
			if err != nil {
				logger.Log("unable to listen for IPv6 mDNS requests: %s", err)
				return err
			}
			defer conn.Close()

			ms := &server{
				conn,
				s.Handler,
				ifaces,
				logger,
			}

			return ms.Run(ctx)
		})
	}

	err := g.Wait()

	if err == context.Canceled {
		return nil
	}

	return err
}

// server is an mDNS server that services for either IPv4 or IPv6.
type server struct {
	conn       udpConn
	handler    Handler
	interfaces []net.Interface
	logger     twelf.Logger
}

func (s *server) Run(ctx context.Context) error {
	if err := s.joinGroup(); err != nil {
		return err
	}

	// close the connection if the context is cancelled
	// this is what allows us to break out of calls to s.con.Recv()
	go func() {
		<-ctx.Done()
		_ = s.conn.Close()
	}()

	buf := make([]byte, 65536)

	for {
		n, src, err := s.conn.Recv(buf)
		if err != nil {
			s.logger.Log("error reading mDNS packet: %s", err)
			// TODO(jmalloc): check for "closed" error and return ctx.Err() instead
			return err
		}

		m := &dns.Msg{}

		if err := m.Unpack(buf[:n]); err != nil {
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
				// request anyway, without many guarantees as to the validitity of the
				// message. We also do not currently support the behavior specified above.
				//
				// Because our DNS responder will not be the only multicast responder on the
				// machine (ie the host OS provides its own) this may not even be possible
				// to implement correctly. See https://tools.ietf.org/html/rfc6762#section-15.2
				// for more information.
				s.logger.DebugString("responding immediately to mDNS query with non-zero TC flag")
			} else {
				s.logger.Log("error parsing mDNS packet: %s", err)
				continue
			}
		}

		if m.Response {
			err = s.handleResponse(ctx, src, m)
		} else {
			err = s.handleRequest(ctx, src, m)
		}

		if err != nil {
			s.logger.Log("error handling mDNS message: %s", err)
		}
	}
}

func (s *server) handleRequest(ctx context.Context, src Source, m *dns.Msg) error {
	req, err := NewRequest(src, m)
	if err != nil {
		return err
	}

	var res, uc, mc *Response

	for _, q := range m.Question {
		unicast, q := WantsUnicastResponse(q)

		// select one of uc or mc to pass to the handler
		if unicast || req.IsLegacy() {
			if uc == nil {
				uc = NewResponse(m, true)
			}
			res = uc
		} else {
			if mc == nil {
				mc = NewResponse(m, false)
			}
			res = mc
		}

		if err := s.handler.HandleQuestion(ctx, req, res, q); err != nil {
			return err
		}
	}

	uOk, err := s.sendResponse(req, uc, src.Address)
	if err != nil {
		return err
	}

	mOk, err := s.sendResponse(req, mc, s.conn.Addr())

	if uOk || mOk {
		test_mutex.Lock()
		fmt.Println("----------------------------------------------------------------------------")
		fmt.Println("")

		if req.IsLegacy() {
			fmt.Printf("--- LEGACY REQUEST <-- %s\n", req.Source.Address)
		} else {
			fmt.Printf("--- REQUEST <-- %s\n", req.Source.Address)
		}
		fmt.Println("")
		fmt.Println(req.Message)

		if uOk {
			fmt.Printf("--- UNICAST RESPONSE --> %s\n", src.Address)
			fmt.Println("")
			fmt.Println(uc.Message)
		}

		if mOk {
			fmt.Printf("--- MULTICAST RESPONSE --> %s\n", s.conn.Addr())
			fmt.Println("")
			fmt.Println(mc.Message)
		}

		test_mutex.Unlock()
	}

	return nil
}

func (s *server) handleResponse(ctx context.Context, src Source, m *dns.Msg) error {
	// TODO(jmalloc): we need to "defend" our records here
	// https://tools.ietf.org/html/rfc6762#section-9
	return nil
}

func (s *server) sendResponse(
	req *Request,
	res *Response,
	addr *net.UDPAddr,
) (bool, error) {
	if res == nil || len(res.Message.Answer) == 0 {
		return false, nil
	}

	buf, err := res.Message.Pack()
	if err != nil {
		return true, err
	}

	return true, s.conn.Send(
		buf,
		req.Source.Interface,
		addr,
	)
}

func (s *server) joinGroup() error {
	ok := false
	pc := s.conn.PacketConn()
	addr := &net.UDPAddr{
		IP: s.conn.Addr().IP,
	}

	for _, i := range s.interfaces {
		if err := pc.JoinGroup(&i, addr); err != nil {
			s.logger.Debug(
				"unable to join the '%s' multicast group on the '%s' interface: %s",
				addr.IP,
				i.Name,
				err,
			)

			continue
		}

		ok = true
	}

	if ok {
		return nil
	}

	return fmt.Errorf(
		"unable to join the '%s' multicast group on any interfaces",
		addr.IP,
	)
}
