package debounce

import "time"

// config holds the settings applied by Options.
type config struct {
	leading       bool
	trailing      bool
	after         time.Duration
	debounceLimit int
	timeLimit     time.Duration
}

// Option configures a [Debouncer]. Pass options to [New] or [NewFunc].
type Option func(*config)

// WithTrailing runs the most recent function once after has passed without
// calls to Do. Every call restarts the wait.
//
// The function runs in its own goroutine. A zero or negative duration runs it
// as soon as possible after the call.
//
// WithTrailing also defines when a burst ends, so that the next call can be
// a leading one (see [WithLeading]).
func WithTrailing(after time.Duration) Option {
	return func(c *config) {
		c.trailing = true
		c.after = after
	}
}

// WithLeading runs the first call of each burst immediately, in the
// goroutine that called Do. Later calls in the same burst are debounced as
// usual.
//
// A burst ends after the [WithTrailing] delay passes without calls. Without
// WithTrailing it never ends, so only the very first call (and the first
// call after each [Debouncer.Stop]) runs on the leading edge.
func WithLeading(enabled bool) Option {
	return func(c *config) {
		c.leading = enabled
	}
}

// WithDebounceLimit runs the n-th call accumulated since the last execution
// immediately, in the goroutine that called Do. This bounds how many calls
// can be dropped in a row.
//
// A value of n <= 0 disables the limit, which is the default.
func WithDebounceLimit(n int) Option {
	return func(c *config) {
		c.debounceLimit = n
	}
}

// WithTimeLimit runs a call immediately, in the goroutine that called Do, if
// at least limit has passed since the first call accumulated after the last
// execution. This bounds how long a steady stream of calls can postpone an
// execution.
//
// The limit is only checked when Do is called. Combined with
// WithTrailing(after), a call waits at most about limit + after.
//
// A value of limit <= 0 disables the limit, which is the default.
func WithTimeLimit(limit time.Duration) Option {
	return func(c *config) {
		c.timeLimit = limit
	}
}
