package transport

import (
	"fmt"
	"net"

	"github.com/jmalloc/twelf/src/twelf"

	ipvx "golang.org/x/net/ipv6"
)

var (
	// IPv6Group is the multicast group used for mDNS over IPv6.
	//
	// See https://tools.ietf.org/html/rfc6762#section-3.
	IPv6Group = net.ParseIP("ff02::fb")

	// IPv6GroupAddress is the address to which mDNS queries are sent when using IPv6.
	//
	// See https://tools.ietf.org/html/rfc6762#section-3.
	IPv6GroupAddress = &net.UDPAddr{IP: IPv6Group, Port: Port}

	// IPv6ListenAddress is the address to which the mDNS server binds when using
	// IPv6. Note that the multicast group address is NOT used in order to control
	// more precisely which network interfaces join the multicast group.
	IPv6ListenAddress = &net.UDPAddr{IP: net.ParseIP("ff02::"), Port: Port}
)

// IPv6Transport is an IPv6-based UDP transport.
type IPv6Transport struct {
	Logger twelf.Logger

	pc *ipvx.PacketConn
}

// Listen starts listening for UDP packets on the given interfaces.
func (t *IPv6Transport) Listen(iface *net.Interface) error {
	addr := IPv6ListenAddress
	conn, err := net.ListenUDP("udp6", addr)
	if err != nil {
		logListenError(t.Logger, addr, err)
		return err
	}

	t.pc = ipvx.NewPacketConn(conn)

	err = t.pc.SetControlMessage(ipvx.FlagInterface, true)
	if err != nil {
		t.pc.Close()
		logListenError(t.Logger, addr, err)
		return err
	}

	err = t.pc.JoinGroup(iface, &net.UDPAddr{
		IP: IPv6Group,
	})
	if err != nil {
		t.pc.Close()
		logListenError(t.Logger, addr, err)
		return err
	}

	logListening(t.Logger, addr, iface)

	return nil
}

// Read reads the next packet from the transport.
func (t *IPv6Transport) Read() (*InboundPacket, error) {
	buf := getBuffer()

	n, cm, src, err := t.pc.ReadFrom(buf)
	if err != nil {
		putBuffer(buf)
		logReadError(t.Logger, t.Group(), err)
		return nil, err
	}

	if cm == nil {
		putBuffer(buf)
		err := fmt.Errorf("empty control message from %s", src)
		logReadError(t.Logger, t.Group(), err)
		return nil, err
	}

	buf = buf[:n]

	return &InboundPacket{
		t,
		Endpoint{
			cm.IfIndex,
			src.(*net.UDPAddr),
		},
		buf,
	}, nil
}

// Write sends a packet via the transport.
func (t *IPv6Transport) Write(p *OutboundPacket) error {
	if _, err := t.pc.WriteTo(
		p.Data,
		&ipvx.ControlMessage{
			IfIndex: p.Destination.InterfaceIndex,
		},
		p.Destination.Address,
	); err != nil {
		logWriteError(t.Logger, p.Destination.Address, t.Group(), err)
		return err
	}

	return nil
}

// Group returns the multicast group address for this transport.
func (t *IPv6Transport) Group() *net.UDPAddr {
	return IPv6GroupAddress
}

// Close closes the transport, preventing further reads and writes.
func (t *IPv6Transport) Close() error {
	return t.pc.Close()
}
