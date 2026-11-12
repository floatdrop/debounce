package debounce_test

import (
	"fmt"
	"runtime"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/floatdrop/debounce/v3"
)

// incr returns a function that increments c.
func incr(c *atomic.Int32) func() {
	return func() { c.Add(1) }
}

// assertCalls waits until every other goroutine in the synctest bubble is
// blocked, so that due timer callbacks have finished, and then checks c.
// synctest.Wait does not advance the fake clock, so timers that are not due
// yet stay pending.
func assertCalls(t *testing.T, c *atomic.Int32, want int32) {
	t.Helper()
	synctest.Wait()
	if got := c.Load(); got != want {
		t.Fatalf("got %d executions, want %d", got, want)
	}
}

// Trailing

func TestTrailing_SingleCall(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		var n atomic.Int32
		d := debounce.New(debounce.WithTrailing(100 * time.Millisecond))

		d.Do(incr(&n))
		assertCalls(t, &n, 0)

		time.Sleep(100 * time.Millisecond)
		assertCalls(t, &n, 1)
	})
}

func TestTrailing_OnlyLastFunctionRuns(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		var n, last atomic.Int32
		d := debounce.New(debounce.WithTrailing(100 * time.Millisecond))

		for i := range int32(5) {
			d.Do(func() {
				n.Add(1)
				last.Store(i)
			})
		}

		time.Sleep(100 * time.Millisecond)
		assertCalls(t, &n, 1)
		if got := last.Load(); got != 4 {
			t.Fatalf("ran function %d, want 4", got)
		}
	})
}

func TestTrailing_EachCallRestartsTheWait(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		var n atomic.Int32
		d := debounce.New(debounce.WithTrailing(100 * time.Millisecond))

		d.Do(incr(&n))
		time.Sleep(50 * time.Millisecond)
		d.Do(incr(&n))

		time.Sleep(99 * time.Millisecond)
		assertCalls(t, &n, 0)

		time.Sleep(time.Millisecond)
		assertCalls(t, &n, 1)
	})
}

func TestTrailing_SteadyCallsPostponeExecution(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		var n atomic.Int32
		d := debounce.New(debounce.WithTrailing(100 * time.Millisecond))

		for range 50 {
			d.Do(incr(&n))
			time.Sleep(20 * time.Millisecond)
		}
		assertCalls(t, &n, 0)

		time.Sleep(80 * time.Millisecond)
		assertCalls(t, &n, 1)
	})
}

func TestTrailing_SeparateBursts(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		var n atomic.Int32
		d := debounce.New(debounce.WithTrailing(50 * time.Millisecond))

		for i := range int32(3) {
			d.Do(incr(&n))
			time.Sleep(50 * time.Millisecond)
			assertCalls(t, &n, i+1)
		}
	})
}

func TestTrailing_NonPositiveDurationRunsImmediately(t *testing.T) {
	t.Parallel()
	for _, after := range []time.Duration{0, -time.Millisecond} {
		t.Run(after.String(), func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				var n atomic.Int32
				d := debounce.New(debounce.WithTrailing(after))

				d.Do(incr(&n))
				assertCalls(t, &n, 1)
			})
		})
	}
}

// Leading

func TestLeading_RunsInCallingGoroutine(t *testing.T) {
	t.Parallel()
	var n atomic.Int32
	d := debounce.New(debounce.WithLeading(true))

	d.Do(incr(&n))

	// No waiting: the leading call must have completed before Do returned.
	if got := n.Load(); got != 1 {
		t.Fatalf("got %d executions right after Do, want 1", got)
	}
}

func TestLeading_OnlyFirstCallRuns(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		var n, first atomic.Int32
		first.Store(-1)
		d := debounce.New(debounce.WithLeading(true))

		for i := range int32(5) {
			d.Do(func() {
				n.Add(1)
				first.Store(i)
			})
		}

		assertCalls(t, &n, 1)
		if got := first.Load(); got != 0 {
			t.Fatalf("ran function %d, want 0", got)
		}
	})
}

func TestLeading_WithoutTrailingTheBurstNeverEnds(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		var n atomic.Int32
		d := debounce.New(debounce.WithLeading(true))

		for range 5 {
			d.Do(incr(&n))
			time.Sleep(time.Hour)
		}

		assertCalls(t, &n, 1)
	})
}

func TestLeading_StopStartsANewBurst(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		var n atomic.Int32
		d := debounce.New(debounce.WithLeading(true))

		d.Do(incr(&n))
		d.Do(incr(&n))
		d.Stop()
		d.Do(incr(&n))

		assertCalls(t, &n, 2)
	})
}

// Leading + trailing

func TestLeadingTrailing_BothEdgesRun(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		var n atomic.Int32
		d := debounce.New(
			debounce.WithLeading(true),
			debounce.WithTrailing(100*time.Millisecond),
		)

		d.Do(incr(&n))
		assertCalls(t, &n, 1)

		d.Do(incr(&n))
		d.Do(incr(&n))
		assertCalls(t, &n, 1)

		time.Sleep(100 * time.Millisecond)
		assertCalls(t, &n, 2)
	})
}

func TestLeadingTrailing_SingleCallRunsOnce(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		var n atomic.Int32
		d := debounce.New(
			debounce.WithLeading(true),
			debounce.WithTrailing(100*time.Millisecond),
		)

		d.Do(incr(&n))
		time.Sleep(time.Second)

		// The leading call consumed the burst's only call: nothing is left
		// for the trailing edge.
		assertCalls(t, &n, 1)
	})
}

// Regression: in the original implementation a lone leading call never ended
// its burst, so the next call, however much later, waited for the trailing
// delay instead of running immediately.
func TestLeadingTrailing_LeadsAgainAfterQuietPeriod(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		var n atomic.Int32
		d := debounce.New(
			debounce.WithLeading(true),
			debounce.WithTrailing(100*time.Millisecond),
		)

		d.Do(incr(&n)) // leading
		time.Sleep(100 * time.Millisecond)
		assertCalls(t, &n, 1) // also lets the timer end the burst

		d.Do(incr(&n)) // new burst: leading again
		assertCalls(t, &n, 2)

		d.Do(incr(&n)) // same burst: trailing
		time.Sleep(99 * time.Millisecond)
		assertCalls(t, &n, 2)
		time.Sleep(time.Millisecond)
		assertCalls(t, &n, 3)

		d.Do(incr(&n)) // the trailing execution ended the burst: leading again
		assertCalls(t, &n, 4)
	})
}

func TestLeadingTrailing_StopCancelsTrailing(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		var n atomic.Int32
		d := debounce.New(
			debounce.WithLeading(true),
			debounce.WithTrailing(100*time.Millisecond),
		)

		d.Do(incr(&n))
		d.Do(incr(&n))
		d.Stop()

		time.Sleep(time.Second)
		assertCalls(t, &n, 1)
	})
}

// Debounce limit

func TestDebounceLimit_RunsEveryNthCall(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		var n atomic.Int32
		d := debounce.New(debounce.WithDebounceLimit(3))

		d.Do(incr(&n))
		d.Do(incr(&n))
		assertCalls(t, &n, 0)

		d.Do(incr(&n))
		assertCalls(t, &n, 1)

		d.Do(incr(&n))
		d.Do(incr(&n))
		assertCalls(t, &n, 1)

		d.Do(incr(&n))
		assertCalls(t, &n, 2)
	})
}

func TestDebounceLimit_RunsInCallingGoroutine(t *testing.T) {
	t.Parallel()
	var n atomic.Int32
	d := debounce.New(debounce.WithDebounceLimit(1))

	d.Do(incr(&n))

	if got := n.Load(); got != 1 {
		t.Fatalf("got %d executions right after Do, want 1", got)
	}
}

// Regression: in the original implementation a limit of 0, as from an unset
// config field, ran every call immediately.
func TestDebounceLimit_NonPositiveDisablesLimit(t *testing.T) {
	t.Parallel()
	for _, limit := range []int{0, -1} {
		t.Run(fmt.Sprint(limit), func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				var n atomic.Int32
				d := debounce.New(
					debounce.WithTrailing(100*time.Millisecond),
					debounce.WithDebounceLimit(limit),
				)

				for range 10 {
					d.Do(incr(&n))
				}
				assertCalls(t, &n, 0)

				time.Sleep(100 * time.Millisecond)
				assertCalls(t, &n, 1)
			})
		})
	}
}

func TestDebounceLimit_WithTrailing_NoExtraTrailingExecution(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		var n atomic.Int32
		d := debounce.New(
			debounce.WithTrailing(100*time.Millisecond),
			debounce.WithDebounceLimit(3),
		)

		for range 3 {
			d.Do(incr(&n))
		}
		assertCalls(t, &n, 1)

		time.Sleep(time.Second)
		assertCalls(t, &n, 1)
	})
}

func TestDebounceLimit_TrailingExecutionResetsCount(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		var n atomic.Int32
		d := debounce.New(
			debounce.WithTrailing(100*time.Millisecond),
			debounce.WithDebounceLimit(3),
		)

		d.Do(incr(&n))
		d.Do(incr(&n))
		time.Sleep(100 * time.Millisecond)
		assertCalls(t, &n, 1) // trailing

		d.Do(incr(&n))
		d.Do(incr(&n))
		assertCalls(t, &n, 1) // only 2 calls since the last execution

		d.Do(incr(&n))
		assertCalls(t, &n, 2) // limit
	})
}

// Regression: in the original implementation a limit-triggered execution
// ended the burst, so with WithLeading the very next call ran as well.
func TestDebounceLimit_WithLeading_NoBackToBackExecutions(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		var ran []int
		d := debounce.New(
			debounce.WithLeading(true),
			debounce.WithDebounceLimit(3),
		)

		for i := range 10 {
			d.Do(func() { ran = append(ran, i) })
		}

		// Leading call, then every 3rd call after it.
		if want := []int{0, 3, 6, 9}; !slices.Equal(ran, want) {
			t.Fatalf("ran calls %v, want %v", ran, want)
		}
	})
}

// Time limit

func TestTimeLimit_RunsCallAfterLimit(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		var n atomic.Int32
		d := debounce.New(debounce.WithTimeLimit(100 * time.Millisecond))

		d.Do(incr(&n))
		time.Sleep(50 * time.Millisecond)
		d.Do(incr(&n))
		time.Sleep(49 * time.Millisecond)
		d.Do(incr(&n))
		assertCalls(t, &n, 0)

		time.Sleep(time.Millisecond)
		d.Do(incr(&n))
		assertCalls(t, &n, 1)
	})
}

func TestTimeLimit_IsCheckedOnlyOnDo(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		var n atomic.Int32
		d := debounce.New(debounce.WithTimeLimit(100 * time.Millisecond))

		d.Do(incr(&n))
		time.Sleep(time.Second)
		assertCalls(t, &n, 0)

		d.Do(incr(&n))
		assertCalls(t, &n, 1)
	})
}

func TestTimeLimit_RunsLatestFunction(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		var last atomic.Int32
		d := debounce.New(debounce.WithTimeLimit(100 * time.Millisecond))

		for i := range int32(11) {
			d.Do(func() { last.Store(i) })
			time.Sleep(10 * time.Millisecond)
		}

		synctest.Wait()
		if got := last.Load(); got != 10 {
			t.Fatalf("ran function %d, want 10", got)
		}
	})
}

func TestTimeLimit_WindowRestartsAfterExecution(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		var n atomic.Int32
		d := debounce.New(debounce.WithTimeLimit(50 * time.Millisecond))

		d.Do(incr(&n))
		time.Sleep(50 * time.Millisecond)
		d.Do(incr(&n))
		assertCalls(t, &n, 1)

		time.Sleep(time.Second)
		d.Do(incr(&n)) // first call of a new window, not over the limit
		assertCalls(t, &n, 1)

		time.Sleep(50 * time.Millisecond)
		d.Do(incr(&n))
		assertCalls(t, &n, 2)
	})
}

// Regression: in the original implementation a time limit of 0 ran every
// call immediately.
func TestTimeLimit_NonPositiveDisablesLimit(t *testing.T) {
	t.Parallel()
	for _, limit := range []time.Duration{0, -time.Millisecond} {
		t.Run(limit.String(), func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				var n atomic.Int32
				d := debounce.New(
					debounce.WithTrailing(100*time.Millisecond),
					debounce.WithTimeLimit(limit),
				)

				for range 10 {
					d.Do(incr(&n))
					time.Sleep(10 * time.Millisecond)
				}
				assertCalls(t, &n, 0)

				time.Sleep(90 * time.Millisecond)
				assertCalls(t, &n, 1)
			})
		})
	}
}

func TestTimeLimit_WithTrailing_BoundsSteadyStream(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		var n atomic.Int32
		d := debounce.New(
			debounce.WithTrailing(50*time.Millisecond),
			debounce.WithTimeLimit(100*time.Millisecond),
		)

		// A call every 10ms for 1s: trailing alone would never run.
		for range 100 {
			d.Do(incr(&n))
			time.Sleep(10 * time.Millisecond)
		}
		assertCalls(t, &n, 9) // at 100ms, 210ms, ..., 980ms

		time.Sleep(40 * time.Millisecond)
		assertCalls(t, &n, 10) // trailing
	})
}

// Regression: in the original implementation a limit-triggered execution
// ended the burst, so with WithLeading the very next call ran as well.
func TestTimeLimit_WithLeading_NoBackToBackExecutions(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		start := time.Now()
		var ran []time.Duration
		d := debounce.New(
			debounce.WithLeading(true),
			debounce.WithTimeLimit(100*time.Millisecond),
		)

		for range 300 {
			d.Do(func() { ran = append(ran, time.Since(start)) })
			time.Sleep(time.Millisecond)
		}

		// Leading call at 0; the call at 1ms starts the first window.
		want := []time.Duration{0, 101 * time.Millisecond, 202 * time.Millisecond}
		if !slices.Equal(ran, want) {
			t.Fatalf("ran at %v, want %v", ran, want)
		}
	})
}

// Stop

func TestStop_CancelsPendingExecution(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		var n atomic.Int32
		d := debounce.New(debounce.WithTrailing(100 * time.Millisecond))

		d.Do(incr(&n))
		d.Stop()

		time.Sleep(time.Second)
		assertCalls(t, &n, 0)
	})
}

func TestStop_DebouncerIsReusable(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		var n atomic.Int32
		d := debounce.New(debounce.WithTrailing(100 * time.Millisecond))

		d.Do(incr(&n))
		d.Stop()
		d.Do(incr(&n))

		time.Sleep(100 * time.Millisecond)
		assertCalls(t, &n, 1)
	})
}

func TestStop_ResetsLimits(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		var n atomic.Int32
		d := debounce.New(
			debounce.WithDebounceLimit(3),
			debounce.WithTimeLimit(100*time.Millisecond),
		)

		d.Do(incr(&n))
		d.Do(incr(&n))
		time.Sleep(time.Second)
		d.Stop()

		d.Do(incr(&n))
		d.Do(incr(&n))
		assertCalls(t, &n, 0)

		d.Do(incr(&n))
		assertCalls(t, &n, 1)
	})
}

func TestStop_WithoutCallsIsSafe(t *testing.T) {
	t.Parallel()
	d := debounce.New(debounce.WithTrailing(time.Millisecond))
	d.Stop()
	d.Stop()
}

// DoAsync

// blockUntil returns a function that blocks until release is closed and then
// increments n. If DoAsync ran it inline, the test would deadlock, which
// synctest reports as a failure.
func blockUntil(release <-chan struct{}, n *atomic.Int32) func() {
	return func() {
		<-release
		n.Add(1)
	}
}

func TestDoAsync_LeadingRunsInNewGoroutine(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		var n atomic.Int32
		release := make(chan struct{})
		d := debounce.New(debounce.WithLeading(true))

		d.DoAsync(blockUntil(release, &n))
		assertCalls(t, &n, 0)

		close(release)
		assertCalls(t, &n, 1)
	})
}

func TestDoAsync_LimitRunsInNewGoroutine(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		var n atomic.Int32
		release := make(chan struct{})
		d := debounce.New(debounce.WithDebounceLimit(1))

		d.DoAsync(blockUntil(release, &n))
		assertCalls(t, &n, 0)

		close(release)
		assertCalls(t, &n, 1)
	})
}

func TestDoAsync_Trailing(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		var n atomic.Int32
		d := debounce.New(debounce.WithTrailing(100 * time.Millisecond))

		d.DoAsync(incr(&n))
		d.DoAsync(incr(&n))

		time.Sleep(100 * time.Millisecond)
		assertCalls(t, &n, 1)
	})
}

// Reentrancy

func TestReentrant_DoFromLeadingFunction(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		var n atomic.Int32
		d := debounce.New(
			debounce.WithLeading(true),
			debounce.WithTrailing(100*time.Millisecond),
		)

		d.Do(func() {
			n.Add(1)
			d.Do(incr(&n)) // same burst: becomes the trailing call
		})
		assertCalls(t, &n, 1)

		time.Sleep(100 * time.Millisecond)
		assertCalls(t, &n, 2)
	})
}

func TestReentrant_DoFromTrailingFunction(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		var n atomic.Int32
		d := debounce.New(debounce.WithTrailing(100 * time.Millisecond))

		d.Do(func() {
			n.Add(1)
			d.Do(incr(&n)) // starts a new burst
		})

		time.Sleep(100 * time.Millisecond)
		assertCalls(t, &n, 1)
		time.Sleep(100 * time.Millisecond)
		assertCalls(t, &n, 2)
	})
}

// Misc

func TestNoOptions_NeverRuns(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		var n atomic.Int32
		d := debounce.New()

		for range 10 {
			d.Do(incr(&n))
			time.Sleep(time.Hour)
		}
		assertCalls(t, &n, 0)
	})
}

func TestNewFunc(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		var n atomic.Int32
		fn, stop := debounce.NewFunc(incr(&n),
			debounce.WithLeading(true),
			debounce.WithTrailing(100*time.Millisecond),
		)

		fn()
		fn()
		fn()
		assertCalls(t, &n, 1)

		time.Sleep(100 * time.Millisecond)
		assertCalls(t, &n, 2)

		fn()
		fn()
		stop()
		assertCalls(t, &n, 3) // leading ran before stop
		time.Sleep(time.Second)
		assertCalls(t, &n, 3) // trailing was cancelled
	})
}

// TestTimelines checks the timelines shown in README.md. Keep them in sync.
func TestTimelines(t *testing.T) {
	t.Parallel()

	const trailing = 25 * time.Millisecond
	tests := []struct {
		name string
		opts []debounce.Option
		want []string
	}{
		{
			name: "Trailing",
			opts: []debounce.Option{debounce.WithTrailing(trailing)},
			want: []string{"e@65"},
		},
		{
			name: "Leading+Trailing",
			opts: []debounce.Option{debounce.WithLeading(true), debounce.WithTrailing(trailing)},
			want: []string{"a@0", "e@65"},
		},
		{
			name: "Trailing+DebounceLimit",
			opts: []debounce.Option{debounce.WithTrailing(trailing), debounce.WithDebounceLimit(3)},
			want: []string{"c@20", "e@65"},
		},
		{
			name: "Trailing+TimeLimit",
			opts: []debounce.Option{debounce.WithTrailing(trailing), debounce.WithTimeLimit(25 * time.Millisecond)},
			want: []string{"d@30", "e@65"},
		},
		{
			name: "Leading",
			opts: []debounce.Option{debounce.WithLeading(true)},
			want: []string{"a@0"},
		},
		{
			name: "Leading+DebounceLimit",
			opts: []debounce.Option{debounce.WithLeading(true), debounce.WithDebounceLimit(3)},
			want: []string{"a@0", "d@30"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			synctest.Test(t, func(t *testing.T) {
				start := time.Now()
				var mu sync.Mutex
				var got []string
				d := debounce.New(tt.opts...)

				for _, name := range []string{"a", "b", "c", "d", "e"} {
					d.Do(func() {
						mu.Lock()
						defer mu.Unlock()
						got = append(got, fmt.Sprintf("%s@%d", name, time.Since(start).Milliseconds()))
					})
					time.Sleep(10 * time.Millisecond)
				}
				time.Sleep(time.Second)

				mu.Lock()
				defer mu.Unlock()
				if !slices.Equal(got, tt.want) {
					t.Fatalf("got %v, want %v", got, tt.want)
				}
			})
		})
	}
}

// Concurrency
//
// These use real time: testing/synctest serializes exactly the races they
// are looking for.

// TestConcurrent_NoDoubleExecution checks that a function passed to Do never
// runs twice, for example once on a limit and again on a late timer.
func TestConcurrent_NoDoubleExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("stress test")
	}
	t.Parallel()

	var doubles atomic.Int64
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		d := debounce.New(
			debounce.WithLeading(true),
			debounce.WithTrailing(5*time.Microsecond),
			debounce.WithDebounceLimit(2),
		)

		var wg sync.WaitGroup
		for range 8 {
			wg.Go(func() {
				for i := range 1000 {
					var runs atomic.Int32
					d.Do(func() {
						if runs.Add(1) > 1 {
							doubles.Add(1)
						}
					})
					if i%64 == 0 {
						runtime.Gosched()
					}
				}
			})
		}
		wg.Wait()
		time.Sleep(50 * time.Microsecond)
		d.Stop()
	}

	if n := doubles.Load(); n != 0 {
		t.Fatalf("%d functions ran more than once", n)
	}
}

// TestConcurrent_LastCallAlwaysRuns checks that the last function passed to Do
// runs once the debouncer goes quiet, whatever the interleaving of calls and
// timer callbacks.
func TestConcurrent_LastCallAlwaysRuns(t *testing.T) {
	if testing.Short() {
		t.Skip("stress test")
	}
	t.Parallel()

	const goroutines, calls = 4, 200
	for range 200 {
		d := debounce.New(debounce.WithTrailing(20 * time.Microsecond))

		var lastRuns atomic.Int32
		var seq int
		var mu sync.Mutex // serializes Do so that "last" is well defined
		var wg sync.WaitGroup
		for range goroutines {
			wg.Go(func() {
				for range calls {
					mu.Lock()
					seq++
					v := seq
					d.Do(func() {
						if v == goroutines*calls {
							lastRuns.Add(1)
						}
					})
					mu.Unlock()
					if v%16 == 0 {
						time.Sleep(20 * time.Microsecond)
					}
				}
			})
		}
		wg.Wait()

		waitUntil := time.Now().Add(time.Second)
		for lastRuns.Load() == 0 && time.Now().Before(waitUntil) {
			time.Sleep(100 * time.Microsecond)
		}
		if got := lastRuns.Load(); got != 1 {
			t.Fatalf("last call ran %d times, want 1", got)
		}
	}
}
