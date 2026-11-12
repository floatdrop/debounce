# Debounce

[![CI](https://github.com/floatdrop/debounce/actions/workflows/ci.yaml/badge.svg)](https://github.com/floatdrop/debounce/actions/workflows/ci.yaml)
![Coverage](https://img.shields.io/badge/Coverage-100.0-brightgreen)
[![Go Reference](https://pkg.go.dev/badge/github.com/floatdrop/debounce/v3.svg)](https://pkg.go.dev/github.com/floatdrop/debounce/v3)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A small, fast Go library that turns a burst of calls into fewer executions. It can run on the trailing edge (after calls stop), on the leading edge (on the first call), or both. Optional limits make sure that a steady stream of calls cannot postpone execution forever.

- **Composable options:** leading and trailing edges, a call-count limit and a time limit, in any combination
- **Safe for concurrent use** from any number of goroutines
- **No allocations** on the hot path, about 5–20 ns per call
- **No dependencies** outside the standard library

> [!TIP]
> Debouncing drops calls on purpose. If every call must eventually run, or you need a steady rate, use a rate limiter such as [`golang.org/x/time/rate`](https://pkg.go.dev/golang.org/x/time/rate) instead.

## Installation

```bash
go get github.com/floatdrop/debounce/v3
```

Requires Go 1.25 or newer.

## Usage

```go
package main

import (
	"fmt"
	"time"

	"github.com/floatdrop/debounce/v3"
)

func main() {
	d := debounce.New(
		debounce.WithTrailing(200*time.Millisecond), // run 200ms after the last call...
		debounce.WithTimeLimit(2*time.Second),       // ...but don't let calls postpone it for much longer than 2s
	)
	defer d.Stop()

	d.Do(func() { fmt.Println("Hello") })
	d.Do(func() { fmt.Println("World") })

	time.Sleep(time.Second)
	// Output: World
}
```

Only the most recent function runs: each call to `Do` replaces the pending one.

If you always debounce the same function, `NewFunc` wraps it for you:

```go
save, stop := debounce.NewFunc(saveDocument, debounce.WithTrailing(time.Second))
defer stop()

for range keystrokes {
	save() // saveDocument runs once typing pauses for a second
}
```

## Options

| Option                 | Effect                                                                          |
| ---------------------- | ------------------------------------------------------------------------------- |
| `WithTrailing(after)`  | Run the most recent function once `after` has passed without calls.            |
| `WithLeading(true)`    | Run the first call of a burst immediately.                                      |
| `WithDebounceLimit(n)` | Run the `n`-th call since the last execution immediately.                       |
| `WithTimeLimit(d)`     | Run a call immediately if the oldest pending call has waited `d` or longer.     |

Most of the time you want `WithTrailing`, plus `WithTimeLimit` if calls can arrive in a steady stream. A limit of zero (or less) disables it.

### How options combine

Here `Do` is called with functions `a` to `e`, 10 ms apart, and the trailing delay is 25 ms. Each row shows which functions run, and when:

```
time                         0     10    20    30    40             65 ms
Do called with               a     b     c     d     e
─────────────────────────────┼─────┼─────┼─────┼─────┼──────────────┼──▶
Trailing                                                            e
Leading + Trailing           a                                      e
Trailing + DebounceLimit(3)              c                          e
Trailing + TimeLimit(25ms)                     d                    e
Leading                      a
Leading + DebounceLimit(3)   a                 d
```

The rules behind it:

- **Calls form bursts.** A burst starts with the first call and ends once the trailing delay passes with no further calls. `WithLeading` runs the first call of each burst. Without `WithTrailing` a burst never ends (until `Stop`), so the leading call runs only once.
- **Limits run calls during a burst.** They count calls and time since the last execution of any kind, and they are checked when `Do` is called. So a pending call waits at most about time limit + trailing delay.
- **Pending calls need a trailing edge to run.** Without `WithTrailing`, calls made after the last execution stay pending. That is why `e` never runs in the last two rows.

## Behavior notes

- **Where functions run.** Leading and limit-triggered executions run synchronously in the goroutine that called `Do`. Trailing executions run in a new goroutine. So a function can start while a previous one is still running. If your function blocks, use `DoAsync`, which always runs it in a new goroutine.
- **Stopping.** `Stop` cancels the pending execution and resets the debouncer, which can be used again afterwards. It does not wait for a function that is already running.
- **Reentrancy.** It is safe to call `Do` from inside a debounced function.
- **No options, no execution.** A debouncer without `WithTrailing`, `WithLeading` or a limit never runs anything.

## Migrating from v1 and v2

| v1                       | v2                       | v3                                    |
| ------------------------ | ------------------------ | ------------------------------------- |
| `New(after)`             | `New(WithDelay(after))`  | `New(WithTrailing(after))`            |
| `WithMaxCalls(n)`        | `WithLimit(n)`           | `WithDebounceLimit(n)`                |
| `WithMaxWait(d)`         |                          | `WithTimeLimit(d)`                    |
| `debounced(fn)`          | `d.Do(fn)`               | `d.Do(fn)` or `d.DoAsync(fn)`         |
| `NewFunc(fn, after)`     | `d.Func(fn)`             | `NewFunc(fn, WithTrailing(after))`    |
|                          | `d.Close()`              | `d.Stop()`                            |
|                          | `Chan(ch, opts...)`      | removed                               |

v2 ran every function in a new goroutine. Use `DoAsync` to keep that behavior.

## Benchmarks

```bash
go test -bench=. -benchmem
```

```
goos: darwin
goarch: arm64
pkg: github.com/floatdrop/debounce/v3
cpu: Apple M3 Max
BenchmarkDo_NoOptions-14             	258259876	         4.531 ns/op	       0 B/op	       0 allocs/op
BenchmarkDo_Leading-14               	268857356	         4.504 ns/op	       0 B/op	       0 allocs/op
BenchmarkDo_Trailing-14              	68640889	        18.08 ns/op	       0 B/op	       0 allocs/op
BenchmarkDo_LeadingAndTrailing-14    	66356386	        18.05 ns/op	       0 B/op	       0 allocs/op
BenchmarkDo_DebounceLimit-14         	268075526	         4.486 ns/op	       0 B/op	       0 allocs/op
BenchmarkDo_TimeLimit-14             	67949530	        17.49 ns/op	       0 B/op	       0 allocs/op
BenchmarkDo_AllOptions-14            	66625795	        18.26 ns/op	       0 B/op	       0 allocs/op
BenchmarkDo_Parallel-14              	 8033890	       127.0 ns/op	       0 B/op	       0 allocs/op
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
