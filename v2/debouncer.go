package debounce

type Debouncer struct {
	inputCh     chan func()
	debouncedCh <-chan func()
}

// Creates new Debouncer instance that will call provided functions with debounce.
func New(opts ...Option) *Debouncer {
	inputCh := make(chan func())
	debouncedCh := Chan(inputCh, opts...)

	go func() {
		for f := range debouncedCh {
			go f() // Do not block reading channel for f execution
		}
	}()

	return &Debouncer{
		inputCh:     inputCh,
		debouncedCh: debouncedCh,
	}
}

// Do adds function f to be executed with debounce.
func (d *Debouncer) Do(f func()) {
	d.inputCh <- f
}

// Func returns func wrapper of function f, that will execute function f with debounce on call.
func (d *Debouncer) Func(f func()) func() {
	return func() {
		d.inputCh <- f
	}
}

// Closes underlying channel in Debouncer instance.
func (d *Debouncer) Close() {
	close(d.inputCh)
}
