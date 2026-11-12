package debounce_test

import (
	"fmt"
	"time"

	"github.com/floatdrop/debounce/v3"
)

func Example() {
	d := debounce.New(
		debounce.WithTrailing(100*time.Millisecond),
		debounce.WithTimeLimit(time.Second),
	)
	defer d.Stop()

	for _, query := range []string{"g", "go", "gol", "gola", "golan", "golang"} {
		d.Do(func() { fmt.Println("search:", query) })
	}

	time.Sleep(300 * time.Millisecond)
	// Output: search: golang
}

func ExampleNewFunc() {
	saves := 0
	save, stop := debounce.NewFunc(func() {
		saves++
		fmt.Println("saved", saves)
	}, debounce.WithTrailing(100*time.Millisecond))
	defer stop()

	save()
	save()
	save()

	time.Sleep(300 * time.Millisecond)
	// Output: saved 1
}

func ExampleWithLeading() {
	d := debounce.New(
		debounce.WithLeading(true),
		debounce.WithTrailing(100*time.Millisecond),
	)

	d.Do(func() { fmt.Println("first") }) // runs immediately
	d.Do(func() { fmt.Println("second") })
	d.Do(func() { fmt.Println("third") }) // runs 100ms later

	time.Sleep(300 * time.Millisecond)
	// Output:
	// first
	// third
}

func ExampleWithDebounceLimit() {
	d := debounce.New(
		debounce.WithTrailing(100*time.Millisecond),
		debounce.WithDebounceLimit(2),
	)

	d.Do(func() { fmt.Println("one") })
	d.Do(func() { fmt.Println("two") })   // the 2nd call runs immediately
	d.Do(func() { fmt.Println("three") }) // runs 100ms later

	time.Sleep(300 * time.Millisecond)
	// Output:
	// two
	// three
}

func ExampleDebouncer_DoAsync() {
	done := make(chan struct{})
	d := debounce.New(debounce.WithLeading(true))

	// With WithLeading, Do would run this in the calling goroutine.
	d.DoAsync(func() {
		fmt.Println("hello from another goroutine")
		close(done)
	})

	<-done
	// Output: hello from another goroutine
}
