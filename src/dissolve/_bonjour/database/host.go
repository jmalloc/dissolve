package database

import "github.com/miekg/dns"

// Scope is an enumeration of the "scopes" of resource records, as described in
// https://tools.ietf.org/html/rfc6762#section-2.
type Scope int

const (
	// SharedScope denotes a resource record set as "shared".
	//
	// A "shared" resource record set is is one where several Multicast DNS
	// responders may have records with the same name, rrtype, and
	// rrclass, and several responders may respond to a particular query.
	SharedScope Scope = 0

	// UniqueScope denotes a resource record set as "unique".
	//
	// A "unique" resource record set is one where all the records with
	// that name, rrtype, and rrclass are conceptually under the control
	// or ownership of a single responder, and it is expected that at
	// most one responder should respond to a query for that name,
	// rrtype, and rrclass.
	UniqueScope Scope = 1
)

// Host is a container for RRSets for the same host name.
type Host struct {
	Name   string
	Scope  Scope
	RRSets map[uint32]RRSet
}

// Query calls fn for each RRSet that matches q.
func (h *Host) Query(q dns.Question, fn func(RRSet)) {
	if q.Qclass == dns.ClassANY && q.Qtype == dns.TypeANY {
		for _, rrs := range h.RRSets {
			fn(rrs)
		}
	} else if q.Qclass == dns.ClassANY {
		for k, rrs := range h.RRSets {
			if hasType(k, q.Qtype) {
				fn(rrs)
			}
		}
	} else if q.Qtype == dns.TypeANY {
		for k, rrs := range h.RRSets {
			if hasClass(k, q.Qclass) {
				fn(rrs)
			}
		}
	} else {
		k := makeKey(q.Qclass, q.Qtype)
		if rrs, ok := h.RRSets[k]; ok {
			fn(rrs)
		}
	}
}

func hasClass(k uint32, c uint16) bool {
	return uint16(k>>16) == c
}

func hasType(k uint32, t uint16) bool {
	return uint16(k&0xff) == t
}

func makeKey(c, t uint16) uint32 {
	return uint32(c)<<16 | uint32(t)
}
