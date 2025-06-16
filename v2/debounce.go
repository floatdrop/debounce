package debounce

import (
	"time"
)

// Option is a functional option for configuring the debouncer.
type Option func(*debounceOptions)

type debounceOptions struct {
	limit int
	delay time.Duration
}

// WithLimit sets the maximum number of incoming inputs before
// passing most recent value downstream.
func WithLimit(limit int) Option {
	return func(options *debounceOptions) {
		options.limit = limit
	}
}

// WithDelay sets time.Duration specifying how long to wait after the last input
// before sending the most recent value downstream.
func WithDelay(d time.Duration) Option {
	return func(options *debounceOptions) {
		options.delay = d
	}
}

// Chan wraps incoming channel and returns channel that emits the last value only
// after no new values are received for the given delay or limit.
//
// If no delay provided - zero delay assumed, so function returns in chan as result.
func Chan[T any](in <-chan T, opts ...Option) <-chan T {
	var options debounceOptions
	for _, opt := range opts {
		opt(&options)
	}

	// If there is no duration - every incoming element must be passed downstream.
	if options.delay == 0 {
		return in
	}

	out := make(chan T, 1)
	go func() {
		defer close(out)

		var (
			timer  *time.Timer = time.NewTimer(options.delay)
			value  T
			hasVal bool
			count  int
		)

		// Function to return the timer channel or nil if timer is not set
		// This avoids blocking on the timer channel if no timer is set
		timerOrNil := func() <-chan time.Time {
			if timer != nil && hasVal {
				return timer.C
			}
			return nil
		}

		for {
			select {
			case v, ok := <-in:
				if !ok { // Input channel is closed, wrapping up
					if hasVal {
						out <- value
					}
					break
				}

				if options.limit != 0 { // If WithLimit specified as non-zero value start counting and emitting
					count++
					if count >= options.limit {
						out <- v
						hasVal = false
						timer.Stop()
						continue
					}
				}

				value = v
				hasVal = true
				timer.Reset(options.delay)
			case <-timerOrNil():
				out <- value
				hasVal = false
			}
		}
	}()
	return out
}
