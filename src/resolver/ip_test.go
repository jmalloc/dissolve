package resolver

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ipToArpa", func() {
	It("returns an in-addr.arpa address for IPv4 addresses", func() {
		addr, ok := ipToArpa("192.168.60.30")

		Expect(addr).To(Equal("30.60.168.192.in-addr.arpa."))
		Expect(ok).To(BeTrue())
	})

	It("returns an ip6.arpa address for IPv6 addresses", func() {
		addr, ok := ipToArpa("2001:db8::567:89ab")

		Expect(addr).To(Equal("b.a.9.8.7.6.5.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.8.b.d.0.1.0.0.2.ip6.arpa."))
		Expect(ok).To(BeTrue())
	})
})
