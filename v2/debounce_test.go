package debounce_test

import (
	"slices"
	"testing"
	"time"

	"github.com/floatdrop/debounce/v2"
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

	expected := []int{4}
	result := collect(out, 1*time.Second)
	if !slices.Equal(expected, result) {
		t.Errorf("expected result = %v, got %v", expected, result)
	}
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

	expected := []int{1, 2, 3}
	result := collect(out, 1*time.Second)
	if !slices.Equal(expected, result) {
		t.Errorf("expected result = %v, got %v", expected, result)
	}
}

func TestDebounce_WithLimit(t *testing.T) {
	in := make(chan int)
	out := debounce.Chan(in, debounce.WithDelay(200*time.Millisecond), debounce.WithLimit(3))

	go func() {
		in <- 1
		time.Sleep(50 * time.Millisecond)
		in <- 2
		time.Sleep(50 * time.Millisecond)
		in <- 3
		time.Sleep(50 * time.Millisecond)
		in <- 4
		time.Sleep(50 * time.Millisecond)
		in <- 5
		time.Sleep(50 * time.Millisecond)
		in <- 6
		time.Sleep(50 * time.Millisecond)
		in <- 7
		time.Sleep(300 * time.Millisecond) // wait longer than debounce delay
		close(in)
	}()

	expected := []int{3, 6, 7}
	result := collect(out, 1*time.Second)
	if !slices.Equal(expected, result) {
		t.Errorf("expected result = %v, got %v", expected, result)
	}
}

func TestDebounce_ChannelCloses(t *testing.T) {
	in := make(chan int)
	out := debounce.Chan(in, debounce.WithDelay(100*time.Millisecond))

	go func() {
		in <- 42
		close(in)
	}()

	expected := []int{42}
	result := collect(out, 1*time.Second)
	if !slices.Equal(expected, result) {
		t.Errorf("expected result = %v, got %v", expected, result)
	}
}

func TestDebounce_EmptyChannelCloses(t *testing.T) {
	in := make(chan int)
	out := debounce.Chan(in, debounce.WithDelay(100*time.Millisecond))

	go func() {
		close(in)
	}()

	expected := []int(nil)
	result := collect(out, 1*time.Second)
	if !slices.Equal(expected, result) {
		t.Errorf("expected result = %v, got %v", expected, result)
	}
}

func TestDebounce_ZeroDelay(t *testing.T) {
	in := make(chan int)
	out := debounce.Chan(in)
	if (<-chan int)(in) != out {
		t.Errorf("expected result = %v, got %v", (<-chan int)(in), out)
	}
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
