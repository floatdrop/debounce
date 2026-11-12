package debounce

import (
	"sync"
	"time"
)

// Debouncer coalesces calls to [Debouncer.Do] according to its options.
// Create one with [New].
//
// A Debouncer is safe for concurrent use by multiple goroutines.
type Debouncer struct {
	cfg   config    // set by New, read-only afterwards
	epoch time.Time // reference point for now

	mu    sync.Mutex
	timer *time.Timer // trailing timer, created on first use

	active      bool          // a burst is in progress
	fn          func()        // most recent pending function, nil if none
	calls       int           // calls accumulated since the last execution
	firstCallAt time.Duration // when the first accumulated call was made
	quietUntil  time.Duration // when the burst ends unless Do is called again
}

// Do submits fn for debounced execution.
//
// Depending on the options, fn runs immediately in the calling goroutine
// (leading edge or a limit reached), later in a separate goroutine (trailing
// edge), or not at all if another call replaces it first. See the package
// documentation for the exact rules.
//
// It is safe to call Do from within fn.
func (d *Debouncer) Do(fn func()) {
	d.mu.Lock()

	var now time.Duration
	if d.cfg.trailing || d.cfg.timeLimit > 0 {
		now = d.now()
	}

	starting := !d.active
	d.active = true

	if d.cfg.trailing {
		d.quietUntil = now + d.cfg.after
		if starting {
			d.startTimer()
		}
	}

	if starting && d.cfg.leading {
		d.mu.Unlock()
		fn()
		return
	}

	if d.calls == 0 {
		d.firstCallAt = now
	}
	d.calls++

	if d.limitReached(now) {
		d.calls = 0
		d.fn = nil
		d.mu.Unlock()
		fn()
		return
	}

	d.fn = fn
	d.mu.Unlock()
}

// DoAsync is like [Debouncer.Do], but fn always runs in a new goroutine, even
// when Do would run it in the calling goroutine.
//
// DoAsync(fn) is equivalent to Do(func() { go fn() }).
func (d *Debouncer) DoAsync(fn func()) {
	d.Do(func() { go fn() })
}

// Stop cancels the pending execution, if any, and resets the Debouncer to its
// initial state. The Debouncer can be used again afterwards.
//
// Stop does not wait for an execution that has already started.
func (d *Debouncer) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil {
		d.timer.Stop()
	}
	d.active = false
	d.fn = nil
	d.calls = 0
}

// now returns the time elapsed since d was created. Unlike time.Now, it only
// reads the monotonic clock, which makes it noticeably cheaper on the Do path.
func (d *Debouncer) now() time.Duration {
	return time.Since(d.epoch)
}

// limitReached reports whether the accumulated calls hit the debounce limit
// or the time limit. It must be called with d.mu held.
func (d *Debouncer) limitReached(now time.Duration) bool {
	return (d.cfg.debounceLimit > 0 && d.calls >= d.cfg.debounceLimit) ||
		(d.cfg.timeLimit > 0 && now-d.firstCallAt >= d.cfg.timeLimit)
}

// startTimer arms the trailing timer for a new burst. It must be called with
// d.mu held.
//
// The timer is not reset on every call. Do only moves quietUntil forward, and
// onTimer goes back to sleep if it fires before quietUntil.
func (d *Debouncer) startTimer() {
	if d.timer == nil {
		d.timer = time.AfterFunc(d.cfg.after, d.onTimer)
	} else {
		d.timer.Reset(d.cfg.after)
	}
}

// onTimer runs in the timer's goroutine. It ends the burst if the quiet period
// has passed, running the pending function if there is one.
//
// The timer can fire while Do or Stop holds the lock, in which case this
// callback runs after the state has changed under it. So it checks the current
// state rather than assuming that its firing is still relevant.
func (d *Debouncer) onTimer() {
	d.mu.Lock()

	if !d.active {
		// Stop was called after the timer fired.
		d.mu.Unlock()
		return
	}

	if wait := d.quietUntil - d.now(); wait > 0 {
		// Do was called since the timer was armed.
		d.timer.Reset(wait)
		d.mu.Unlock()
		return
	}

	d.active = false
	fn := d.fn
	d.fn = nil
	d.calls = 0
	d.mu.Unlock()

	if fn != nil {
		fn()
	}
}
