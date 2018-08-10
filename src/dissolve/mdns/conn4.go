package mdns

import (
	"net"

	ipvx "golang.org/x/net/ipv4"
)

type udpConn4 struct {
	pc *ipvx.PacketConn
}

func listen4() (*udpConn4, error) {
	conn, err := net.ListenUDP("udp4", IPv4Address)
	if err != nil {
		return nil, err
	}

	pc := ipvx.NewPacketConn(conn)
	pc.SetControlMessage(ipvx.FlagInterface, true)

	return &udpConn4{pc}, nil
}

func (c *udpConn4) PacketConn() packetConn {
	return c.pc
}

func (c *udpConn4) Addr() *net.UDPAddr {
	return IPv4Address
}

func (c *udpConn4) Recv(buf []byte) (int, Source, error) {
	n, cm, src, err := c.pc.ReadFrom(buf)
	return n, Source{cm.IfIndex, src.(*net.UDPAddr)}, err
}

func (c *udpConn4) Send(buf []byte, iface int, addr *net.UDPAddr) error {
	_, err := c.pc.WriteTo(
		buf,
		&ipvx.ControlMessage{IfIndex: iface},
		addr,
	)

	return err
}

func (c *udpConn4) Close() error {
	return c.pc.Close()
}
