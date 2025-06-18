package debounce

// Debouncer wraps a debounced channel of functions,
// allowing callers to submit or wrap functions that will only be executed
// according to the debounce configuration (e.g., delay, limit).
type Debouncer struct {
	inputCh     chan func()   // Channel to receive submitted functions
	debouncedCh <-chan func() // Debounced output channel from Chan
}

// New creates a new Debouncer instance.
// Submitted functions will be debounced according to the provided options,
// such as WithDelay or WithLimit.
//
// Each debounced function is executed in its own goroutine to avoid blocking the Debouncer.
func New(opts ...Option) *Debouncer {
	inputCh := make(chan func())
	debouncedCh := Chan(inputCh, opts...)

	go func() {
		for f := range debouncedCh {
			// Execute function without blocking the debounce processing
			go f()
		}
	}()

	return &Debouncer{
		inputCh:     inputCh,
		debouncedCh: debouncedCh,
	}
}

// Do submits a function f to be executed according to the debounce rules.
// Only the most recent function may be executed, depending on delay and limit configuration.
func (d *Debouncer) Do(f func()) {
	d.inputCh <- f
}

// Func returns a debounced wrapper of the given function f.
// Each call to the returned function submits f to the debouncer.
// Depending on debounce configuration, f may not be executed immediately—or at all—
// if subsequent calls override it before the debounce conditions are met.
func (d *Debouncer) Func(f func()) func() {
	return func() {
		d.inputCh <- f
	}
}

// Closes underlying channel in Debouncer instance.
func (d *Debouncer) Close() {
	close(d.inputCh)
}
