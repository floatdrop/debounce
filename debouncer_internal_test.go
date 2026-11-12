package debounce

import (
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

// These tests call onTimer directly to simulate a timer callback that is
// delivered late: the timer fired, but the callback only got the lock after
// Do or Stop had changed the state. That interleaving is too narrow to hit
// reliably from the outside.

// Regression: in the original implementation a late callback ran the
// function passed to the latest Do immediately, skipping its quiet period.
func TestOnTimer_LateCallbackAfterDo(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var n atomic.Int32
		d := New(WithTrailing(100 * time.Millisecond))

		d.Do(func() { n.Add(1) })
		d.onTimer()
		synctest.Wait()
		if got := n.Load(); got != 0 {
			t.Fatalf("late callback ran the function early (%d executions)", got)
		}

		time.Sleep(100 * time.Millisecond)
		synctest.Wait()
		if got := n.Load(); got != 1 {
			t.Fatalf("got %d executions, want 1", got)
		}
	})
}

func TestOnTimer_LateCallbackAfterStop(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var n atomic.Int32
		d := New(WithTrailing(100 * time.Millisecond))

		d.Do(func() { n.Add(1) })
		d.Stop()
		d.onTimer() // the timer fired just before Stop, but its callback runs after

		time.Sleep(time.Second)
		synctest.Wait()
		if got := n.Load(); got != 0 {
			t.Fatalf("got %d executions, want 0", got)
		}
	})
}

func TestOnTimer_LateCallbackAfterStopAndDo(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var first, second atomic.Int32
		d := New(WithTrailing(100 * time.Millisecond))

		d.Do(func() { first.Add(1) })
		d.Stop()
		d.Do(func() { second.Add(1) })
		d.onTimer() // late callback from the first burst

		time.Sleep(99 * time.Millisecond)
		synctest.Wait()
		if first.Load() != 0 || second.Load() != 0 {
			t.Fatalf("ran too early: first=%d second=%d", first.Load(), second.Load())
		}

		time.Sleep(time.Second)
		synctest.Wait()
		if first.Load() != 0 || second.Load() != 1 {
			t.Fatalf("got first=%d second=%d, want 0 and 1", first.Load(), second.Load())
		}
	})
}
