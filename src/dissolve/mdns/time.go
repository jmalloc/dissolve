package mdns

import (
	"context"
	"math/rand"
	"time"
)

// randT returns a random duraction between 0 and d, inclusive.
func randT(d time.Duration) time.Duration {
	return randTBetween(0, d)
}

// randTBetween returns a random duraction between min and max, inclusive.
func randTBetween(min, max time.Duration) time.Duration {
	return time.Duration(
		rand.Int63n(
			int64(max-min),
		) + int64(max),
	)
}

// sleep sleeps for a duration of d, or until ctx is canceled.
// It runs nil if the sleep duration passes before ctx is canceled.
func sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
