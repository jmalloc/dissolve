package resolver_test

import (
	"context"
	"net"
	"time"

	. "github.com/jmalloc/dissolve/src/resolver"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StandardResolver (net.Resolver parity)", func() {
	var (
		subject, builtin Resolver
		ctx              context.Context
		cancel           func()
	)

	BeforeEach(func() {
		subject = &StandardResolver{}
		builtin = &net.Resolver{}

		c, f := context.WithTimeout(context.Background(), 3*time.Second)
		ctx, cancel = c, f // assign in separate statement to silence "go vet" error
	})

	AfterEach(func() {
		cancel()
	})

	Describe("LookupAddr", func() {
		It("returns the same results as the built-in implementation", func() {
			s, err := subject.LookupAddr(ctx, "8.8.8.8")
			Expect(err).ShouldNot(HaveOccurred())

			r, err := builtin.LookupAddr(ctx, "8.8.8.8")
			Expect(err).ShouldNot(HaveOccurred())

			Expect(s).To(ConsistOf(r))
		})
	})

	Describe("LookupCNAME", func() {
		It("returns the same results as the built-in implementation", func() {
			s, err := subject.LookupCNAME(ctx, "mail.icecave.com.au")
			Expect(err).ShouldNot(HaveOccurred())

			r, err := builtin.LookupCNAME(ctx, "mail.icecave.com.au")
			Expect(err).ShouldNot(HaveOccurred())

			Expect(s).To(Equal(r))
		})
	})

	Describe("LookupHost", func() {
		It("returns the same results as the built-in implementation", func() {
			s, err := subject.LookupHost(ctx, "www.icecave.com.au")
			Expect(err).ShouldNot(HaveOccurred())

			r, err := builtin.LookupHost(ctx, "www.icecave.com.au")
			Expect(err).ShouldNot(HaveOccurred())

			Expect(s).To(ConsistOf(r))
		})
	})

	// Describe("LookupIPAddr", func() {
	//     (ctx context.Context, host string) ([]net.IPAddr, error)
	// })

	Describe("LookupMX", func() {
		It("returns the same results as the built-in implementation", func() {
			s, err := subject.LookupMX(ctx, "icecave.com.au")
			Expect(err).ShouldNot(HaveOccurred())

			r, err := builtin.LookupMX(ctx, "icecave.com.au")
			Expect(err).ShouldNot(HaveOccurred())

			Expect(s).To(HaveLen(len(r)))

			// expect preferences to be the same at each entry
			for idx := 0; idx < len(r); idx++ {
				a, b := s[idx], r[idx]
				Expect(a.Pref).To(Equal(b.Pref))
			}
		})
	})

	Describe("LookupNS", func() {
		It("returns the same results as the built-in implementation", func() {
			s, err := subject.LookupNS(ctx, "icecave.com.au")
			Expect(err).ShouldNot(HaveOccurred())

			r, err := builtin.LookupNS(ctx, "icecave.com.au")
			Expect(err).ShouldNot(HaveOccurred())

			Expect(s).To(ConsistOf(r))
		})
	})
	//
	// Describe("LookupPort", func() {
	//     (ctx context.Context, network, service string) (port int, err error)
	// })
	//
	// Describe("LookupSRV", func() {
	//     (ctx context.Context, service, proto, name string) (cname string, addrs []*net.SRV, err error)
	// })
	//
	Describe("LookupTXT", func() {
		It("returns the same results as the built-in implementation", func() {
			s, err := subject.LookupTXT(ctx, "icecave.com.au")
			Expect(err).ShouldNot(HaveOccurred())

			r, err := builtin.LookupTXT(ctx, "icecave.com.au")
			Expect(err).ShouldNot(HaveOccurred())

			Expect(s).To(ConsistOf(r))
		})
	})
})
