package debounce_test

import (
	"fmt"
	"time"

	"github.com/floatdrop/debounce"
)

func ExampleNew() {
	// Create a debounced function with 500ms delay
	debounced := debounce.New(500 * time.Millisecond)

	// This will only execute once, after 500ms
	debounced(func() {
		fmt.Println("Hello, World!")
	})

	debounced(func() {
		fmt.Println("This will be executed instead")
	})

	// Wait for execution
	time.Sleep(1 * time.Second)
	// Output: This will be executed instead
}

func ExampleNewFunc() {
	a := 0
	// Create a debounced function with 500ms delay
	debouncedFunc := debounce.NewFunc(func() {
		a += 1
		fmt.Println(a)
	}, 500*time.Millisecond)

	// This will only execute once, after 500ms
	debouncedFunc()

	debouncedFunc()

	// Wait for execution
	time.Sleep(1 * time.Second)
	// Output: 1
}

func ExampleWithMaxCalls() {
	debounced := debounce.New(500*time.Millisecond, debounce.WithMaxCalls(2))

	debounced(func() {
		fmt.Println("One")
	})

	debounced(func() {
		fmt.Println("Two")
	})

	debounced(func() {
		fmt.Println("Three")
	})

	time.Sleep(1 * time.Second)
	// Output: Two
	// Three
}
