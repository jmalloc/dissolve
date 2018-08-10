package mdns

import (
	"net"

	ipvx "golang.org/x/net/ipv6"
)

// udpConn6 is a UDP connection that communicates using IPv6.
type udpConn6 struct {
	pc *ipvx.PacketConn
}

func listen6() (*udpConn6, error) {
	conn, err := net.ListenUDP("udp6", IPv6Address)
	if err != nil {
		return nil, err
	}

	pc := ipvx.NewPacketConn(conn)
	pc.SetControlMessage(ipvx.FlagInterface, true)

	return &udpConn6{pc}, nil
}

func (c *udpConn6) PacketConn() packetConn {
	return c.pc
}

func (c *udpConn6) Addr() *net.UDPAddr {
	return IPv6Address
}

func (c *udpConn6) Recv(buf []byte) (int, Source, error) {
	n, cm, src, err := c.pc.ReadFrom(buf)
	return n, Source{cm.IfIndex, src.(*net.UDPAddr)}, err
}

func (c *udpConn6) Send(buf []byte, iface int, addr *net.UDPAddr) error {
	_, err := c.pc.WriteTo(
		buf,
		&ipvx.ControlMessage{IfIndex: iface},
		addr,
	)

	return err
}

func (c *udpConn6) Close() error {
	return c.pc.Close()
}
