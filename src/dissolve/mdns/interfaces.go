package mdns

import (
	"errors"
	"net"
)

// multicastInterfaces returns the list of network interfaces that are enabled
// and support
func multicastInterfaces() ([]net.Interface, error) {
	candidates, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var matches []net.Interface
	const flags = net.FlagUp | net.FlagMulticast

	for _, i := range candidates {
		if (i.Flags & flags) != 0 {
			matches = append(matches, i)
		}
	}

	if len(matches) == 0 {
		return nil, errors.New("no multicast interfaces available")
	}

	return matches, nil
}
