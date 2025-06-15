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

const (
	NoLimitCalls = math.MaxInt
	NoLimitWait  = time.Duration(math.MaxInt64)
)

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

// Returns a debounced function. The provided function will be executed
// after a period of inactivity, or when a maximum number of calls or
// time threshold is reached, if configured.
// The debounced function can be invoked with different functions, if needed,
// the last one will win.
func New(after time.Duration, options ...Option) func(fn func()) {
	d := &debouncer{
		after:     after,
		startWait: time.Now(),
		maxWait:   NoLimitWait,
		maxCalls:  NoLimitCalls,
	}

	// Creating timer and immediately stop it, so there will be always allocated Timer
	d.timer = time.AfterFunc(NoLimitWait, func() {
		d.mu.Lock()
		if d.calls == 0 {
			return // MaxCalls or MaxWait reached, call can be dropped
		}
		d.calls = 0
		d.mu.Unlock()
		d.fn()
	})
	d.timer.Stop()

	for _, opt := range options {
		opt(d)
	}

	return func(fn func()) {
		d.debouncedCall(fn)
	}
}

// NewFunc returns a debounced function that always debounces the provided function.
func NewFunc(fn func(), after time.Duration, options ...Option) func() {
	debounce := New(after, options...)

	return func() {
		debounce(fn)
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
	fn func()
}

func (d *debouncer) callLimitReached() bool {
	return d.maxCalls != NoLimitCalls && d.calls >= d.maxCalls
}

func (d *debouncer) timeLimitReached() bool {
	return d.maxWait != NoLimitWait && time.Since(d.startWait) >= d.maxWait
}

func (d *debouncer) debouncedCall(fn func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Refreshing function reference, so d.timer will call right function
	d.fn = fn

	// If this is a first call, store startWait time
	if d.calls == 0 {
		d.startWait = time.Now()
	}

	// Counting calls
	d.calls++

	// If the function has been called more than the limit, or if the wait time
	// has exceeded the limit, execute the function immediately.
	if d.callLimitReached() || d.timeLimitReached() {
		d.timer.Stop() // Stop the timer to prevent it from firing later
		d.calls = 0
		fn := d.fn
		go fn() // Execute outside mutex to avoid blocking
	} else {
		// Restarting timer, if limits were ok
		d.timer.Reset(d.after)
	}
}
