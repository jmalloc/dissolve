//+build !debug

package responder

import (
	"github.com/jmalloc/dissolve/src/dissolve/mdns/transport"
	"github.com/miekg/dns"
)

func dumpRequestResponse(
	in *transport.InboundPacket,
	query *dns.Msg,
	unicast *dns.Msg,
	multicast *dns.Msg,
) {
}
