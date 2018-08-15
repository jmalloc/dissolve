package transport

import (
	"net"

	"github.com/miekg/dns"
)

// Port is the mDNS port number.
const Port = 5353

// Transport is an interface for communicating via UDP.
type Transport interface {
	// Listen starts listening for UDP packets on the given interface.
	Listen(iface *net.Interface) error

	// Read reads the next packet from the transport.
	Read() (*InboundPacket, error)

	// Write sends a packet via the transport.
	Write(*OutboundPacket) error

	// Group returns the multicast group address for this transport.
	Group() *net.UDPAddr

	// Close closes the transport, preventing further reads and writes.
	Close() error
}

// SendResponse sends a DNS message as a response to an inbound packet.
func SendResponse(in *InboundPacket, to *net.UDPAddr, m *dns.Msg) (bool, error) {
	if len(m.Question) == 0 &&
		len(m.Answer) == 0 &&
		len(m.Ns) == 0 &&
		len(m.Extra) == 0 {
		return false, nil
	}

	out, err := NewOutboundPacket(
		Endpoint{
			InterfaceIndex: in.Source.InterfaceIndex,
			Address:        to,
		},
		m,
	)
	if err != nil {
		return false, err
	}
	defer out.Close()

	return true, in.Transport.Write(out)
}

// SendUnicastResponse sends a DNS message as a unicast response to an inbound
// packet.
func SendUnicastResponse(in *InboundPacket, m *dns.Msg) (bool, error) {
	return SendResponse(in, in.Source.Address, m)
}

// SendMulticastResponse sends a DNS message as a multicast response to an
// inbound packet.
func SendMulticastResponse(in *InboundPacket, m *dns.Msg) (bool, error) {
	return SendResponse(in, in.Transport.Group(), m)

}
