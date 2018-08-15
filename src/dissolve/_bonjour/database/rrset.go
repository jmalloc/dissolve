package database

import "github.com/miekg/dns"

// RRSet is a set of DNS records that share the same name, class and type.
type RRSet struct {
	Class   uint16
	Type    uint16
	Records []dns.RR
}
