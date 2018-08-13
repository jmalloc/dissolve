//+build debug

package mdns

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/jmalloc/dissolve/src/dissolve/mdns/transport"
	"github.com/miekg/dns"
)

var logMutex sync.Mutex

func indent(s string) string {
	return "\t" + strings.Replace(s, "\n", "\n\t", -1)
}

func dumpRequestResponse(
	in *transport.InboundPacket,
	query *dns.Msg,
	unicast *dns.Msg,
	multicast *dns.Msg,
) {
	logMutex.Lock()
	defer logMutex.Unlock()

	fmt.Fprintln(os.Stderr, strings.Repeat("-", 80))
	fmt.Fprintln(os.Stderr, "")

	fmt.Fprintf(os.Stderr, "QUERY FROM %s", in.Source.Address)
	if in.Source.IsLegacy() {
		fmt.Fprintf(os.Stderr, " (legacy)")
	}
	fmt.Fprintln(os.Stderr, "\n")
	fmt.Fprintln(os.Stderr, indent(query.String()))

	if len(unicast.Answer) > 0 {
		fmt.Fprintln(os.Stderr, "UNICAST RESPONSE")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, indent(unicast.String()))
	}

	if len(multicast.Answer) > 0 {
		fmt.Fprintln(os.Stderr, "MULTICAST RESPONSE")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, indent(multicast.String()))
	}
}
