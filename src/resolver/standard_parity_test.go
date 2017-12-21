package resolver_test

import (
	"context"
	"net"
	"sort"
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

			sort.Strings(s)
			sort.Strings(r)

			Expect(s).To(Equal(r))
		})
	})

	// Describe("LookupCNAME", func() {
	//     (ctx context.Context, host string) (cname string, err error)
	// })

	Describe("LookupHost", func() {
		It("returns the same results as the built-in implementation", func() {
			s, err := subject.LookupHost(ctx, "github.com")
			Expect(err).ShouldNot(HaveOccurred())

			r, err := builtin.LookupHost(ctx, "github.com")
			Expect(err).ShouldNot(HaveOccurred())

			sort.Strings(s)
			sort.Strings(r)

			Expect(s).To(Equal(r))
		})
	})

	// Describe("LookupIPAddr", func() {
	//     (ctx context.Context, host string) ([]net.IPAddr, error)
	// })
	//
	// Describe("LookupMX", func() {
	//     (ctx context.Context, name string) ([]*net.MX, error)
	// })
	//
	// Describe("LookupNS", func() {
	// 	It("returns the same results as the built-in implementation", func() {
	// 		s, err := subject.LookupNS(ctx, "github.com")
	// 		Expect(err).ShouldNot(HaveOccurred())
	//
	// 		r, err := builtin.LookupNS(ctx, "github.com")
	// 		Expect(err).ShouldNot(HaveOccurred())
	//
	// 		Expect(s).To(Equal(r))
	// 	})
	// })
	//
	// Describe("LookupPort", func() {
	//     (ctx context.Context, network, service string) (port int, err error)
	// })
	//
	// Describe("LookupSRV", func() {
	//     (ctx context.Context, service, proto, name string) (cname string, addrs []*net.SRV, err error)
	// })
	//
	// Describe("LookupTXT", func() {
	//     (ctx context.Context, name string) ([]string, error)
	// })
})
