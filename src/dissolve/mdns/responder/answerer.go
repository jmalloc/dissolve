package responder

import (
	"context"
	"net"

	"github.com/jmalloc/dissolve/src/dissolve/mdns"

	"github.com/miekg/dns"
)

// Answerer is an interface that provides answers to DNS questions.
type Answerer interface {
	// Answer populates an answer to a single DNS question.
	// The implementation must allow concurrent calls.
	Answer(context.Context, *Question, *Answer) error
}

// Question encapsulates a DNS question.
type Question struct {
	dns.Question

	Query     *dns.Msg
	Interface net.Interface
}

// Answer is an answer to a DNS question.
type Answer struct {
	// Unique contains the records that belong to "unique" record sets.
	//
	// A "unique" resource record set is one where all the records with
	// that name, rrtype, and rrclass are conceptually under the control
	// or ownership of a single responder, and it is expected that at
	// most one responder should respond to a query for that name,
	// rrtype, and rrclass.
	//
	// See // https://tools.ietf.org/html/rfc6762#section-2.
	Unique ResponseSections

	// SharedScope contains the records that belong to "shared" record sets.
	//
	// A "shared" resource record set is is one where several Multicast DNS
	// responders may have records with the same name, rrtype, and
	// rrclass, and several responders may respond to a particular query.
	//
	// See // https://tools.ietf.org/html/rfc6762#section-2.
	Shared ResponseSections
}

// appendToMessage appends the answer's records to m.
func (a *Answer) appendToMessage(m *dns.Msg, legacy bool) {
	if legacy {
		m.Answer = append(m.Answer, a.Unique.AnswerSection...)
		m.Ns = append(m.Ns, a.Unique.AuthoritySection...)
		m.Extra = append(m.Extra, a.Unique.AdditionalSection...)
	} else {
		m.Answer = appendUnique(m.Answer, a.Unique.AnswerSection)
		m.Ns = appendUnique(m.Ns, a.Unique.AuthoritySection)
		m.Extra = appendUnique(m.Extra, a.Unique.AdditionalSection)
	}

	m.Answer = append(m.Answer, a.Shared.AnswerSection...)
	m.Ns = append(m.Ns, a.Shared.AuthoritySection...)
	m.Extra = append(m.Extra, a.Shared.AdditionalSection...)

}

// appendUnique appends copies of the records in source to target, with the
// "unique record" bit set.
func appendUnique(target, source []dns.RR) []dns.RR {
	for _, r := range source {
		target = append(
			target,
			mdns.SetUniqueRecord(r),
		)
	}

	return target
}

// ResponseSections contains the various response sections of a response to a
// DNS query.
type ResponseSections struct {
	AnswerSection     []dns.RR
	AuthoritySection  []dns.RR
	AdditionalSection []dns.RR
}

// IsEmpty returns true if the response does not contain any records.
func (rs *ResponseSections) IsEmpty() bool {
	return len(rs.AnswerSection) == 0 &&
		len(rs.AuthoritySection) == 0 &&
		len(rs.AdditionalSection) == 0
}

// Answer appends records to the "answer" section of the answer.
func (rs *ResponseSections) Answer(records ...dns.RR) {
	rs.AnswerSection = append(rs.AnswerSection, records...)
}

// Authority appends records to the "authority" section of the answer.
func (rs *ResponseSections) Authority(records ...dns.RR) {
	rs.AuthoritySection = append(rs.AuthoritySection, records...)
}

// Additional appends records to the "additional" section of the answer.
func (rs *ResponseSections) Additional(records ...dns.RR) {
	rs.AdditionalSection = append(rs.AdditionalSection, records...)
}

// UnionAnswerer is an answerer that combines answers from multiple answerers.
type UnionAnswerer []Answerer

// Answer populates an answer to a single DNS question.
// The implementation must allow concurrent calls.
func (an UnionAnswerer) Answer(ctx context.Context, q *Question, a *Answer) error {
	for _, x := range an {
		if err := x.Answer(ctx, q, a); err != nil {
			return err
		}
	}

	return nil
}
