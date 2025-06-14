// Copyright © 2019 Bjørn Erik Pedersen <bjorn.erik.pedersen@gmail.com>.
// Copyright © 2025 Vsevolod Strukchinsky <floatdrop@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package debounce provides a debouncer func.
package debounce

import (
	"math"
	"sync"
	"time"
)

// Option is a functional option for configuring the debouncer.
type Option func(*debouncer)

const NoLimit = -1

// WithMaxCalls sets the maximum number of calls before the debounced function is executed.
// By default, there is no limit.
func WithMaxCalls(count int) Option {
	return func(d *debouncer) {
		d.maxCalls = count
	}
}

// WithMaxWait sets the maximum wait time before the debounced function is executed.
func WithMaxWait(limit time.Duration) Option {
	return func(d *debouncer) {
		d.maxWait = limit
	}
}

// New returns a debounced function that takes another functions as its argument.
// This function will be called when the debounced function stops being called
// for the given duration.
// The debounced function can be invoked with different functions, if needed,
// the last one will win.
func New(after time.Duration, options ...Option) func(f func()) {
	d := &debouncer{
		after:     after,
		startWait: time.Now(),
		maxWait:   time.Duration(math.MaxInt64), // effectively no limit
		maxCalls:  NoLimit,
	}

	// Creating timer and immediately stop it, so there will be always allocated Timer
	d.timer = time.AfterFunc(time.Duration(math.MaxInt64), func() {
		// d.f should always be set at this time, but just to be safe we
		// check for nil (in case timer somehow fires before any debounce calls)
		if d.f != nil {
			d.f()
		}
	})
	d.timer.Stop()

	for _, opt := range options {
		opt(d)
	}

	return func(f func()) {
		d.add(f)
	}
}

// NewFunc return debounce function, that calls f every time.
func NewFunc(f func(), after time.Duration, options ...Option) func() {
	debounce := New(after, options...)

	return func() {
		debounce(f)
	}
}

type debouncer struct {
	mu    sync.Mutex
	after time.Duration
	timer *time.Timer

	calls    int
	maxCalls int

	startWait time.Time
	maxWait   time.Duration

	// Stores last function to debounce. Will be called after specified duration.
	f func()
}

func (d *debouncer) callLimitReached() bool {
	return d.maxCalls != -1 && d.calls >= d.maxCalls
}

func (d *debouncer) timeLimitReached() bool {
	return time.Since(d.startWait) >= d.maxWait
}

func (d *debouncer) add(f func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Refreshing function reference, so d.timer will call right function
	d.f = f

	// If this is a first call, store startWait time
	if d.calls == 0 {
		d.startWait = time.Now()
	}

	// Counting calls
	d.calls += 1

	// If the function has been called more than the limit, or if the wait time
	// has exceeded the limit, execute the function immediately.
	if d.callLimitReached() || d.timeLimitReached() {
		d.timer.Stop()
		d.calls = 0
		fn := d.f // Capture function
		go fn()   // Execute outside mutex to avoid blocking
	} else {
		// Restarting timer, if limits were ok
		d.timer.Reset(d.after)
	}
}
