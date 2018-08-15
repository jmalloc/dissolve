package database

import "github.com/miekg/dns"

// Database is a record of all hosts known to the server.
type Database struct {
	Hosts map[string]Host
}

// Query calls fn for each RRSet that matches q.
func (db *Database) Query(q dns.Question, fn func(RRSet)) {
	if h, ok := db.Hosts[q.Name]; ok {
		h.Query(q, fn)
	}
}
