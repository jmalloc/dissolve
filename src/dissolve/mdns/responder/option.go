package responder

import (
	"net"

	"github.com/dogmatiq/dodeca/logging"
)

// Option is a function that applies an option to a server created by
// NewServer().
type Option func(*Responder) error

// UseLogger returns a server option that sets the logger used by the server.
func UseLogger(l logging.Logger) Option {
	return func(r *Responder) error {
		r.logger = l
		return nil
	}
}

// UseInterface sets the network interface that is used by the server.
//
// If this option is not provided, the server will choose the interface used to
// access the internet.
func UseInterface(iface net.Interface) Option {
	return func(r *Responder) error {
		r.iface = &iface
		return nil
	}
}

// DisableIPv4 is a server option that prevents the server from listening for
// IPv4 messages.
func DisableIPv4(r *Responder) error {
	r.disableIPv4 = true
	return nil
}

// DisableIPv6 is a server option that prevents the server from listening for
// IPv6 messages.
func DisableIPv6(r *Responder) error {
	r.disableIPv6 = true
	return nil
}
