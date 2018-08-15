package responder

import (
	"errors"
	"net"
)

// internetInterface returns the network interface that is used to connect to
// the internet. This is a fairly naive solution that assumes whatever network
// is used to connect to Google's DNS server is the appropriate interface.
func internetInterface() (net.Interface, error) {
	candidates, err := net.Interfaces()
	if err != nil {
		return net.Interface{}, err
	}

	con, err := net.Dial("udp4", "8.8.8.8:53")
	if err != nil {
		return net.Interface{}, err
	}

	ip := con.LocalAddr().(*net.UDPAddr).IP
	con.Close()

	for _, i := range candidates {
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}

		for _, a := range addrs {
			if ipn, ok := a.(*net.IPNet); ok {
				if ipn.IP.Equal(ip) {
					return i, nil
				}
			}
		}
	}

	return net.Interface{}, errors.New("could not find internet network interface")
}
