package resolver

import (
	"bytes"
	"fmt"
	"net"
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
