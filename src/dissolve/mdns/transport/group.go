package transport

import (
	"fmt"
	"net"

	"github.com/jmalloc/twelf/src/twelf"
)

// packetConn contains the methods common to *ipv4.PacketConn and *ipv6.PacketConn.
type packetConn interface {
	JoinGroup(*net.Interface, net.Addr) error
}

// joinGroup joins the mDNS multicast group on each of the given interfaces.
func joinGroup(
	pc packetConn,
	group net.IP,
	ifaces []net.Interface,
	logger twelf.Logger,
) ([]net.Interface, error) {
	addr := &net.UDPAddr{
		IP: group,
	}

	joined := make([]net.Interface, 0, len(ifaces))

	for _, i := range ifaces {
		if err := pc.JoinGroup(&i, addr); err != nil {
			logger.Debug(
				"unable to join the '%s' multicast group on the '%s' interface: %s",
				addr.IP,
				i.Name,
				err,
			)
		} else {
			joined = append(joined, i)
		}
	}

	if len(joined) > 0 {
		return joined, nil
	}

	return nil, fmt.Errorf(
		"unable to join the '%s' multicast group on any interfaces",
		addr.IP,
	)
}
