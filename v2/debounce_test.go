package debounce_test

import (
	"testing"
	"time"

	"github.com/floatdrop/debounce/v2"

	"github.com/stretchr/testify/assert"
)

// helper to collect output with a timeout
func collect[T any](ch <-chan T, timeout time.Duration) []T {
	var results []T
	timer := time.NewTimer(timeout)
	for {
		select {
		case v, ok := <-ch:
			if !ok {
				return results
			}
			results = append(results, v)
		case <-timer.C:
			return results
		}
	}
}

func TestDebounce_LastValueOnly(t *testing.T) {
	in := make(chan int)
	out := debounce.Chan(in, debounce.WithDelay(200*time.Millisecond))

	go func() {
		in <- 1
		time.Sleep(50 * time.Millisecond)
		in <- 2
		time.Sleep(50 * time.Millisecond)
		in <- 3
		time.Sleep(50 * time.Millisecond)
		in <- 4
		time.Sleep(300 * time.Millisecond) // wait longer than debounce delay
		close(in)
	}()

	result := collect(out, 1*time.Second)
	assert.Equal(t, []int{4}, result)
}

func TestDebounce_MultipleValuesSpacedOut(t *testing.T) {
	in := make(chan int)
	out := debounce.Chan(in, debounce.WithDelay(100*time.Millisecond))

	go func() {
		in <- 1
		time.Sleep(150 * time.Millisecond)
		in <- 2
		time.Sleep(150 * time.Millisecond)
		in <- 3
		time.Sleep(150 * time.Millisecond)
		close(in)
	}()

	result := collect(out, 1*time.Second)
	assert.Equal(t, []int{1, 2, 3}, result)
}

func TestDebounce_ChannelCloses(t *testing.T) {
	in := make(chan int)
	out := debounce.Chan(in, debounce.WithDelay(100*time.Millisecond))

	go func() {
		in <- 42
		time.Sleep(200 * time.Millisecond)
		close(in)
	}()

	result := collect(out, 1*time.Second)
	assert.Equal(t, []int{42}, result)
}

func BenchmarkDebounce_Insert(b *testing.B) {
	in := make(chan int)
	_ = debounce.Chan(in, debounce.WithDelay(100*time.Millisecond))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		in <- i
	}
	b.StopTimer()
	close(in)
}
