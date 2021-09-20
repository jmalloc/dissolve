package transport

import (
	"net"

	"github.com/dogmatiq/dodeca/logging"
)

func logListening(logger logging.Logger, addr *net.UDPAddr, iface *net.Interface) {
	logging.Debug(
		logger,
		"listening for mDNS requests on %s (%s)",
		addr,
		iface.Name,
	)
}

func logListenError(logger logging.Logger, addr *net.UDPAddr, err error) {
	logging.Log(
		logger,
		"unable to listen for mDNS requests on %s: %s",
		addr,
		err,
	)
}

func logReadError(logger logging.Logger, addr *net.UDPAddr, err error) {
	logging.Log(
		logger,
		"unable to read mDNS packet via %s: %s",
		addr,
		err,
	)
}

func logWriteError(logger logging.Logger, dest, addr *net.UDPAddr, err error) {
	logging.Log(
		logger,
		"unable to send mDNS packet to %s via %s: %s",
		dest,
		addr,
		err,
	)
}
