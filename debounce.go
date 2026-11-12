// Package debounce coalesces bursts of function calls into fewer executions.
//
// A [Debouncer] receives functions through [Debouncer.Do] and decides, based on
// its options, which of them actually run:
//
//   - [WithTrailing] runs the most recent function once calls have stopped for
//     a given duration. This is the classic debounce.
//   - [WithLeading] runs the first call of a burst immediately.
//   - [WithDebounceLimit] runs every n-th call, so steady traffic still makes
//     progress.
//   - [WithTimeLimit] runs a call immediately once the oldest pending call has
//     waited for a given duration.
//
// Options can be combined. A typical setup debounces with a short trailing
// delay and caps the total delay with a time limit:
//
//	d := debounce.New(
//		debounce.WithTrailing(100*time.Millisecond),
//		debounce.WithTimeLimit(time.Second),
//	)
//	d.Do(save) // save runs after 100ms without calls, or after about a second of continuous calls
//
// # Bursts
//
// Calls are grouped into bursts. A burst starts with the first call to Do and,
// if WithTrailing(after) is set, ends once after has passed without further
// calls. Without WithTrailing a burst only ends when [Debouncer.Stop] is called.
//
// Between executions Do only records the function it was given. Only the most
// recent one runs; earlier ones are dropped. Every execution, whether leading,
// limit-triggered or trailing, clears the pending function and resets the
// limit counters.
//
// # Where functions run
//
// Leading and limit-triggered executions run synchronously in the goroutine
// that called Do. Trailing executions run in their own goroutine. A function
// may therefore start while a previous one is still running. Use
// [Debouncer.DoAsync] to run every execution in its own goroutine.
package debounce

import "time"

// New returns a Debouncer configured with the given options.
//
// Configure at least [WithTrailing] or [WithLeading]: without either, calls
// only run when a limit is reached, and with no options at all they never run.
func New(options ...Option) *Debouncer {
	d := &Debouncer{epoch: time.Now()}
	for _, opt := range options {
		opt(&d.cfg)
	}
	return d
}

// NewFunc returns a debounced version of fn and a function that cancels any
// pending execution of it.
//
// It is a shorthand for creating a Debouncer with [New] and always passing fn
// to [Debouncer.Do]:
//
//	save, cancel := debounce.NewFunc(saveDocument, debounce.WithTrailing(time.Second))
//	defer cancel()
//
//	save() // queued
//	save() // queued again; saveDocument runs once, a second after this call
func NewFunc(fn func(), options ...Option) (debounced func(), stop func()) {
	d := New(options...)
	return func() { d.Do(fn) }, d.Stop
}
