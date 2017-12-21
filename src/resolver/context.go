package resolver

import (
	"context"
	"time"
)

// WithMulticastWait returns a new context that specifies the minimum time to
// wait when performing a multicast DNS query. This value is used by resolvers
// to decide when to stop waiting for additional responses.
//
// If the parent's wait time is already longer than w, parent is returned.
func WithMulticastWait(parent context.Context, w time.Duration) context.Context {
	if e, ok := parent.Value(multicastWaitKey).(time.Duration); ok && e > w {
		return parent
	}

	return context.WithValue(parent, multicastWaitKey, w)
}

// MulticastWait returns the minimum time to wait for responses when performing
// a multicast query on behalf of ctx. ok is false if no wait duration is
// specified.
func MulticastWait(ctx context.Context) (w time.Duration, ok bool) {
	w, ok = ctx.Value(multicastWaitKey).(time.Duration)
	return
}

// ResolveMulticastWait resolves the minimum wait time specified by ctx to a
// a point in time. If ctx does not specify a wait duration, w is added to the
// current time. If the context deadline occurs sooner than this resolved time,
// it is returned instead.
func ResolveMulticastWait(ctx context.Context, w time.Duration) time.Time {
	if e, ok := ctx.Value(multicastWaitKey).(time.Duration); ok {
		w = e
	}

	t := time.Now().Add(w)

	if d, ok := ctx.Deadline(); ok {
		if d.Before(t) {
			return d
		}
	}

	return t
}

type multicastWaitKeyType struct{}

var multicastWaitKey multicastWaitKeyType
