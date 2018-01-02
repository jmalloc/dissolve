package resolver

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"sort"
	"strconv"
)

// ipToArpa returns the "arpa." domain name used to lookup the given IP in
// a PTR record. It returns (ip, false) if ip is not an IP address.
func ipToArpa(ip string) (string, bool) {
	v6 := net.ParseIP(ip)
	if v6 == nil {
		return ip, false
	}

	if v4 := v6.To4(); v4 != nil {
		return fmt.Sprintf(
			"%d.%d.%d.%d.in-addr.arpa.",
			v4[3],
			v4[2],
			v4[1],
			v4[0],
		), true
	}

	buf := &bytes.Buffer{}
	for idx := 15; idx >= 0; idx-- {
		octet := int64(v6[idx])
		high := octet >> 4
		low := octet & 0xf

		buf.WriteString(strconv.FormatInt(low, 16))
		buf.WriteRune('.')
		buf.WriteString(strconv.FormatInt(high, 16))
		buf.WriteRune('.')
	}

	buf.WriteString("ip6.arpa.")

	return buf.String(), true
}

// sortSRV sorts SRV records by priority, and shuffles records within each
// priority grouping by weight, as per https://tools.ietf.org/html/rfc2782.
func sortSRV(s []*net.SRV) {
	if len(s) <= 1 {
		return
	}

	sort.Slice(s, func(i int, j int) bool {
		a, b := s[i], s[j]

		if a.Priority == b.Priority {
			// RFC: "To select a target to be contacted next, arrange all SRV
			// RRs (that have not been ordered yet) in any order, except that all
			// those with weight 0 are placed at the beginning of the list."
			return a.Weight < b.Weight
		}

		return a.Priority < b.Priority
	})

	i := 0
	p := s[0].Priority
	for j, rec := range s {
		if rec.Priority != p {
			shuffleSRV(s[i:j])
			i = j
			p = rec.Priority
		}
	}

	shuffleSRV(s[i:])
}

// shuffleSRV randomly reorders s according to the weights specified by each SRV
// record as per https://tools.ietf.org/html/rfc2782.
func shuffleSRV(s []*net.SRV) {
	// RFC: "Compute the sum of the weights of those RRs"
	var sum int
	for _, rec := range s {
		sum += int(rec.Weight)
	}

	// if the sum is zero, all weights must be zero
	if sum == 0 {
		return
	}

	for i := range s {
		// RFC: "choose a uniform random number between 0 and the sum computed (inclusive)"
		r := rand.Intn(sum)
		a := 0

		for j, rec := range s[i:] {
			// RFC: "with each RR associate the running sum in the selected order"
			a += int(rec.Weight)

			// RFC: "select the RR whose running sum value is the first in the
			// selected order which is greater than or equal to the random
			// number selected"
			if a >= r {
				// RFC: "Remove this SRV RR from the set of the unordered SRV
				// RRs and apply the described algorithm to the unordered SRV
				// RRs to select the next target host"

				// We do this by moving the selected element to the "next"
				// position the slice (i)
				s[i], s[j] = s[j], s[i]
				sum -= int(rec.Weight)
			}
		}

		// RFC: "Continue the ordering process until there are no unordered SRV RRs"
	}
}
