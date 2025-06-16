package debounce_test

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/floatdrop/debounce/v2"
)

func ExampleNew() {
	debouncer := debounce.New(debounce.WithDelay(200 * time.Millisecond))
	debouncer.Do(func() { fmt.Println("Hello") })
	debouncer.Do(func() { fmt.Println("World") })
	time.Sleep(time.Second)
	debouncer.Close()
	// Output: World
}

func ExampleDebouncer_Func() {
	var counter int32
	debouncer := debounce.New(debounce.WithDelay(200 * time.Millisecond)).Func(func() {
		atomic.AddInt32(&counter, 1)
	})
	debouncer()
	debouncer()
	time.Sleep(time.Second)
	fmt.Println(atomic.LoadInt32(&counter))
	// Output: 1
}

func ExampleChan() {
	in := make(chan int)
	out := debounce.Chan(in, debounce.WithDelay(200*time.Millisecond))

	go func() {
		for value := range out {
			fmt.Println(value)
		}
	}()

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

	time.Sleep(time.Second)
	// Output: 4
}
