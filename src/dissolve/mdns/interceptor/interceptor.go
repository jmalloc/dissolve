package interceptor

import (
	"context"
	"encoding/binary"
	"io"
	"net"
	"time"

	"github.com/jmalloc/dissolve/src/dissolve/mdns/transport"
	"github.com/jmalloc/dissolve/src/dissolve/names"
	"github.com/miekg/dns"
)

// interceptor accepts DNS queries over a net.Conn (a pipe made by Dialer), and
// forwards them via unicast or multicast, as appropriate.
type interceptor struct {
	ctx  context.Context
	net  string
	addr string
	dial func(context.Context, string, string) (net.Conn, error)

	conn    net.Conn
	domains []names.FQDN
}

// Run reads DNS queries from i.conn and forwards them via unicast or multicast.
func (i *interceptor) Run() error {
	defer i.conn.Close()

	for {
		query, err := readMessageTCP(i.conn)
		if err != nil {
			return err
		}

		reply, err := i.forward(query)
		if err != nil {
			return err
		}

		err = writeMessageTCP(i.conn, reply)
		if err != nil {
			return err
		}
	}
}

// isMulticast returns true if query contains questions exclusively for
// multicast domains.
func (i *interceptor) isMulticast(query []byte) bool {
	var m dns.Msg

	if err := m.Unpack(query); err != nil {
		return false
	}

	if len(m.Question) == 0 {
		return false
	}

	for _, q := range m.Question {
		n, err := names.Parse(q.Name)
		if err != nil {
			return false
		}

		if !i.isMulticastName(n) {
			return false
		}
	}

	return true
}

// isMulticastName returns true if n is within one of the multicast domain
// names.
func (i *interceptor) isMulticastName(n names.Name) bool {
	if !n.IsQualified() {
		return false
	}

	f := n.(names.FQDN)

	for _, d := range i.domains {
		if f.IsWithin(d) {
			return true
		}
	}

	return false
}

// forward sends a query and awaits the response.
func (i *interceptor) forward(query []byte) ([]byte, error) {
	if i.isMulticast(query) {
		return i.multicast(query)
	}

	return i.unicast(query)
}

// forward sends a query via unicast and awaits the response.
func (i *interceptor) unicast(query []byte) ([]byte, error) {
	conn, err := i.dial(i.ctx, i.net, i.addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// TODO: use the read deadline set on the other end of the pipe
	conn.SetReadDeadline(
		time.Now().Add(5 * time.Second),
	)

	if _, ok := conn.(net.PacketConn); ok {
		return i.unicastUDP(conn, query)
	}

	return i.unicastTCP(conn, query)
}

func (i *interceptor) unicastTCP(conn net.Conn, query []byte) ([]byte, error) {
	if err := writeMessageTCP(conn, query); err != nil {
		return nil, err
	}

	if c, ok := conn.(closeWrite); ok {
		if err := c.CloseWrite(); err != nil {
			return nil, err
		}
	}

	return readMessageTCP(conn)
}

func (i *interceptor) unicastUDP(conn net.Conn, query []byte) ([]byte, error) {
	if _, err := conn.Write(query); err != nil {
		return nil, err
	}

	buf := make([]byte, 512) // as per RFC-1035

	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}

	return buf[:n], nil
}

// forward sends a query via multicast and awaits the response.
func (i *interceptor) multicast(query []byte) ([]byte, error) {
	conn, err := net.ListenUDP("udp", nil)
	if err != nil {
		return nil, err
	}

	// TODO: use the read deadline set on the other end of the pipe
	conn.SetReadDeadline(
		time.Now().Add(5 * time.Second),
	)

	_, err = conn.WriteTo(query, transport.IPv4GroupAddress)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 512) // as per RFC-1035

	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}

	return buf[:n], nil
}

// closeWrite is an interface for connections that allow closing of the write
// side of the pipe.
type closeWrite interface {
	CloseWrite() error
}

// readMessageTCP reads a DNS message from conn.
func readMessageTCP(conn net.Conn) ([]byte, error) {
	var length uint16
	if err := binary.Read(
		conn,
		binary.BigEndian,
		&length,
	); err != nil {
		return nil, err
	}

	buf := make([]byte, length)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return nil, err
	}

	return buf, nil
}

// writeMessageTCP writes a DNS message to conn.
func writeMessageTCP(conn net.Conn, buf []byte) error {
	if err := binary.Write(
		conn,
		binary.BigEndian,
		uint16(len(buf)),
	); err != nil {
		return err
	}

	_, err := conn.Write(buf)
	return err
}
