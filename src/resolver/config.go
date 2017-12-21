package resolver

import "github.com/miekg/dns"

// DefaultConfig is the default DNS to use.
var DefaultConfig *dns.ClientConfig

func init() {
	if conf, err := dns.ClientConfigFromFile("/etc/resolv.conf"); err == nil {
		DefaultConfig = conf
	} else {
		DefaultConfig = &dns.ClientConfig{
			Servers:  []string{"8.8.8.8", "8.8.4.4"},
			Search:   nil,
			Port:     "53",
			Ndots:    1,
			Timeout:  5,
			Attempts: 2,
		}
	}
}
