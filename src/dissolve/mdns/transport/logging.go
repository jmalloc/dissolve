package transport

import (
	"net"
	"sort"
	"strings"

	"github.com/jmalloc/twelf/src/twelf"
)

func logListening(logger twelf.Logger, addr *net.UDPAddr, ifaces []net.Interface) {
	names := make([]string, len(ifaces))

	for i, iface := range ifaces {
		names[i] = iface.Name
	}

	sort.Strings(names)

	logger.Debug(
		"listening for mDNS requests on %s (%s)",
		addr,
		strings.Join(names, ", "),
	)
}

func logListenError(logger twelf.Logger, addr *net.UDPAddr, err error) {
	logger.Log("unable to listen for mDNS requests on %s: %s", addr, err)
}

func logReadError(logger twelf.Logger, addr *net.UDPAddr, err error) {
	logger.Log("unable to read mDNS packet via %s: %s", addr, err)
}

func logWriteError(logger twelf.Logger, dest, addr *net.UDPAddr, err error) {
	logger.Log("unable to send mDNS packet to %s via %s: %s", dest, addr, err)
}
