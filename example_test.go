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
