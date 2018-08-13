package transport

import (
	"github.com/miekg/dns"
)

// InboundPacket is a UDP packet received from a transport.
type InboundPacket struct {
	Transport Transport
	Source    Endpoint
	Data      []byte
}

// Message returns the DNS message contained in a packet.
func (p *InboundPacket) Message() (*dns.Msg, error) {
	m := &dns.Msg{}
	return m, m.Unpack(p.Data)
}

// Close returns the packet's data buffer to the pool.
func (p *InboundPacket) Close() {
	putBuffer(p.Data)
	p.Data = nil
}

// OutboundPacket is a UDP packet to be sent by a transport.
type OutboundPacket struct {
	Destination Endpoint
	Data        []byte
}

// Close returns the packet's data buffer to the pool.
func (p *OutboundPacket) Close() {
	putBuffer(p.Data)
	p.Data = nil
}

// NewOutboundPacket marshals the message m into p.Data.
func NewOutboundPacket(dest Endpoint, m *dns.Msg) (*OutboundPacket, error) {
	buf := getBuffer()

	d, err := m.PackBuffer(buf)
	if err != nil {
		putBuffer(buf)
		return nil, err
	}

	return &OutboundPacket{dest, d}, nil
}
