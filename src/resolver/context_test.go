package resolver_test

import (
	"context"
	"time"

	. "github.com/jmalloc/dissolve/src/resolver"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("WithMulticastWait", func() {
	It("sets the duration if the parent does not already specify one", func() {
		ctx := WithMulticastWait(context.Background(), 123)

		w, _ := MulticastWait(ctx)
		Expect(w).To(Equal(time.Duration(123)))
	})

	It("sets the duration if the parent's duration is shorter", func() {
		ctx := WithMulticastWait(context.Background(), 5)
		ctx = WithMulticastWait(ctx, 123)

		w, _ := MulticastWait(ctx)
		Expect(w).To(Equal(time.Duration(123)))
	})

	It("returns the parents duration if it's longer", func() {
		ctx := WithMulticastWait(context.Background(), 123)
		ctx = WithMulticastWait(ctx, 5)

		w, _ := MulticastWait(ctx)
		Expect(w).To(Equal(time.Duration(123)))
	})
})

var _ = Describe("MulticastWait", func() {
	It("returns the duration", func() {
		ctx := WithMulticastWait(context.Background(), 123)

		w, ok := MulticastWait(ctx)
		Expect(ok).To(BeTrue())
		Expect(w).To(Equal(time.Duration(123)))
	})

	It("returns false if no duration is set", func() {
		_, ok := MulticastWait(context.Background())
		Expect(ok).To(BeFalse())
	})
})

var _ = Describe("ResolveMulticastWait", func() {
	Context("when the context has a wait duration", func() {
		var ctx context.Context

		BeforeEach(func() {
			ctx = WithMulticastWait(context.Background(), 30*time.Second)
		})

		It("builds the threshold from the wait duration", func() {
			t := ResolveMulticastWait(ctx, time.Minute)

			expected := time.Now().Add(30 * time.Second)
			Expect(t).To(BeTemporally("~", expected))
		})

		It("returns the context deadline if it's before the wait duration", func() {
			c, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			t := ResolveMulticastWait(c, time.Minute)

			expected, _ := c.Deadline()
			Expect(t).To(BeTemporally("~", expected))
		})

		It("ignores the context deadline if it's after the wait duration", func() {
			c, cancel := context.WithTimeout(ctx, 45*time.Second)
			defer cancel()

			t := ResolveMulticastWait(c, time.Minute)

			expected := time.Now().Add(30 * time.Second)
			Expect(t).To(BeTemporally("~", expected))
		})
	})

	Context("when the context does not have a wait duration", func() {
		It("builds the threshold from the w argument", func() {
			t := ResolveMulticastWait(context.Background(), time.Minute)

			expected := time.Now().Add(time.Minute)
			Expect(t).To(BeTemporally("~", expected))
		})

		It("returns the context deadline if it's before the w argument", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			t := ResolveMulticastWait(ctx, time.Minute)

			expected, _ := ctx.Deadline()
			Expect(t).To(BeTemporally("~", expected))
		})

		It("ignores the context deadline if it's after the w argument", func() {
			c, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()

			t := ResolveMulticastWait(c, time.Minute)

			expected := time.Now().Add(time.Minute)
			Expect(t).To(BeTemporally("~", expected))
		})
	})
})
