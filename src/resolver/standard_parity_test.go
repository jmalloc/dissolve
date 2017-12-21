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
		subject, ref Resolver
		ctx          context.Context
		cancel       func()
	)

	BeforeEach(func() {
		subject = &StandardResolver{}
		ref = &net.Resolver{}

		c, f := context.WithTimeout(context.Background(), 3*time.Second)
		ctx, cancel = c, f // assign in separate statement to silence "go vet" error
	})

	AfterEach(func() {
		cancel()
	})

	Describe("LookupAddr", func() {
		It("returns the same results as the reference implementation", func() {
			s, err := subject.LookupAddr(ctx, "8.8.8.8")
			Expect(err).ShouldNot(HaveOccurred())

			r, err := ref.LookupAddr(ctx, "8.8.8.8")
			Expect(err).ShouldNot(HaveOccurred())

			sort.Strings(s)
			sort.Strings(r)

			Expect(s).To(Equal(r))
		})
	})

	// Describe("LookupCNAME", func() {
	//     (ctx context.Context, host string) (cname string, err error)
	// })
	//
	// Describe("LookupHost", func() {
	//     (ctx context.Context, host string) (addrs []string, err error)
	// })
	//
	// Describe("LookupIPAddr", func() {
	//     (ctx context.Context, host string) ([]net.IPAddr, error)
	// })
	//
	// Describe("LookupMX", func() {
	//     (ctx context.Context, name string) ([]*net.MX, error)
	// })
	//
	// Describe("LookupNS", func() {
	//     (ctx context.Context, name string) ([]*net.NS, error)
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
