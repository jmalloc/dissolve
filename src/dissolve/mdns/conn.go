package mdns

import "net"

// udpConn is an interface for performing UDP communication.
// It abstracts the differences between dealing with UDP over IPv4 and IPv6.
type udpConn interface {
	PacketConn() packetConn
	Addr() *net.UDPAddr
	Recv(buf []byte) (int, Source, error)
	Send(buf []byte, iface int, addr *net.UDPAddr) error
	Close() error
}

// packetConn contains the methods common to *ipv4.PacketConn and *ipv6.PacketConn.
type packetConn interface {
	JoinGroup(*net.Interface, net.Addr) error
}
