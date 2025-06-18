// Package debounce provides a generic debounce utility for channels.
// It allows throttling the rate of emitted values using optional delay and/or limit mechanisms.
package debounce

import (
	"time"
)

// options encapsulates the debounce configuration: delay and limit.
type options struct {
	limit int
	delay time.Duration
}

// Option is a functional option for configuring the debouncer.
type Option func(*options)

// WithLimit sets a maximum number of times the debounce delay can be reset
// before a value is forcibly emitted. This acts as a safeguard against constant bouncing.
func WithLimit(limit int) Option {
	return func(options *options) {
		options.limit = limit
	}
}

// WithDelay sets the debounce delay — the amount of quiet time (no new inputs)
// required before emitting the most recent value.
func WithDelay(d time.Duration) Option {
	return func(options *options) {
		options.delay = d
	}
}

// Chan wraps an input channel and returns a debounced output channel.
// Debouncing behavior is defined by the combination of WithDelay and WithLimit:
//   - WithDelay delays value emission until no new values are received for `delay`.
//   - WithLimit limits the number of delay resets (i.e., bouncing) before emission is forced.
//
// If both are set, a value will be emitted after either the `delay` passes without new input,
// or after the delay has been reset `limit` times.
//
// If delay is 0, the function returns the input channel unmodified.
func Chan[T any](in <-chan T, opts ...Option) <-chan T {
	var options options
	for _, opt := range opts {
		opt(&options)
	}

	// Optimization: no debouncing if delay is zero
	if options.delay == 0 {
		return in
	}

	out := make(chan T, 1)
	go func() {
		defer close(out)

		var (
			timer     *time.Timer // Timer to manage delay
			lastValue T           // Last received value
			hasValue  bool        // Whether a value is currently pending emission
			count     int         // Number of delay resets since last emission
		)

		// Function to return the timer channel or nil if timer is not set
		// This avoids blocking on the timer channel if no timer is set
		timerOrNil := func() <-chan time.Time {
			if timer != nil {
				return timer.C
			}
			return nil
		}

		for {
			select {
			case v, ok := <-in:
				if !ok {
					// Input channel closed — emit any pending value.
					if hasValue {
						out <- lastValue
					}
					return
				}

				lastValue = v
				hasValue = true

				// On every new input, increment the reset count.
				count++

				// Force emit if limit reached
				if options.limit != 0 && count >= options.limit {
					out <- v
					hasValue = false
					count = 0
					if timer != nil {
						timer.Stop()
					}
					continue
				}

				// Start or reset the delay timer
				if timer == nil {
					timer = time.NewTimer(options.delay)
				} else {
					timer.Reset(options.delay)
				}
			case <-timerOrNil():
				// Delay has passed — emit the last value
				if hasValue {
					out <- lastValue
					hasValue = false
					count = 0
				}
			}
		}
	}()
	return out
}
