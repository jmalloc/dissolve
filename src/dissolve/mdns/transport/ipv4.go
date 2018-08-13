package transport

import (
	"net"

	"github.com/jmalloc/twelf/src/twelf"

	ipvx "golang.org/x/net/ipv4"
)

var (
	// IPv4Group is the multicast group used for mDNS over IPv4.
	//
	// See https://tools.ietf.org/html/rfc6762#section-3.
	IPv4Group = net.ParseIP("224.0.0.251")

	// IPv4GroupAddress is the address to which mDNS queries are sent when using IPv4.
	//
	// See https://tools.ietf.org/html/rfc6762#section-3.
	IPv4GroupAddress = &net.UDPAddr{IP: IPv4Group, Port: Port}

	// IPv4ListenAddress is the address to which the mDNS server binds when using
	// IPv4. Note that the multicast group address is NOT used in order to control
	// more precisely which network interfaces join the multicast group.
	IPv4ListenAddress = &net.UDPAddr{IP: net.ParseIP("224.0.0.0"), Port: Port}
)

// IPv4Transport is an IPv4-based UDP transport.
type IPv4Transport struct {
	Interfaces []net.Interface
	Logger     twelf.Logger
	pc         *ipvx.PacketConn
}

// Listen starts listening for UDP packets over this interface.
func (t *IPv4Transport) Listen() error {
	addr := IPv4ListenAddress
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		logListenError(t.Logger, addr, err)
		return err
	}

	logListening(t.Logger, addr)

	t.pc = ipvx.NewPacketConn(conn)
	t.pc.SetControlMessage(ipvx.FlagInterface, true)

	if err := joinGroup(
		t.pc,
		IPv4Group,
		t.Interfaces,
		t.Logger,
	); err != nil {
		t.pc.Close()
		return err
	}

	return nil
}

// Read reads the next packet from the transport.
func (t *IPv4Transport) Read() (*InboundPacket, error) {
	buf := getBuffer()

	n, cm, src, err := t.pc.ReadFrom(buf)
	if err != nil {
		putBuffer(buf)
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
func (t *IPv4Transport) Write(p *OutboundPacket) error {
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
func (t *IPv4Transport) Group() *net.UDPAddr {
	return IPv4GroupAddress
}

// Close closes the transport, preventing further reads and writes.
func (t *IPv4Transport) Close() error {
	return t.pc.Close()
}
