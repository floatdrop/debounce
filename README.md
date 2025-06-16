# Debounce

[![CI](https://github.com/floatdrop/debounce/actions/workflows/ci.yaml/badge.svg)](https://github.com/floatdrop/debounce/actions/workflows/ci.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/floatdrop/debounce)](https://goreportcard.com/report/github.com/floatdrop/debounce)
[![Go Coverage](https://github.com/floatdrop/debounce/wiki/coverage.svg)](https://raw.githack.com/wiki/floatdrop/debounce/coverage.html)
[![Go Reference](https://pkg.go.dev/badge/github.com/floatdrop/debounce/v2.svg)](https://pkg.go.dev/github.com/floatdrop/debounce/v2)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A simple, thread-safe debounce library for Go that delays function execution until after a specified duration has elapsed since the last invocation. Perfect for rate limiting, reducing redundant operations, and optimizing performance in high-frequency scenarios.

## Features

- **Zero allocations**: No allocations on sunbsequent debounce calls
- **Thread-safe**: Safe for concurrent use across multiple goroutines
- **Channel support**: Can be used on top of `chan` with [Chan](https://pkg.go.dev/github.com/floatdrop/debounce/v2#Chan) function.
- **Configurable delays and limits**: Set custom behaviour with [WithDelay](https://pkg.go.dev/github.com/floatdrop/debounce/v2#WithDelay) and [WithLimit](https://pkg.go.dev/github.com/floatdrop/debounce/v2#WithLimit) options
- **Zero dependencies**: Built using only Go standard library

## Installation

```bash
go get github.com/floatdrop/debounce/v2
```

## Usage

```golang
import (
	"fmt"
	"time"

	"github.com/floatdrop/debounce/v2"
)

func main() {
	debouncer := debounce.New(debounce.WithDelay(200 * time.Millisecond))
	debouncer.Do(func() { fmt.Println("Hello") })
	debouncer.Do(func() { fmt.Println("World") })
	time.Sleep(time.Second)
	// Output: World
}
```

## Benchmarks

```bash
go test -bench=. -benchmem
```

```
goos: darwin
goarch: arm64
pkg: github.com/floatdrop/debounce/v2
cpu: Apple M3 Max
BenchmarkDebounce_Insert-14    	 3318151	       341.9 ns/op	       0 B/op	       0 allocs/op
BenchmarkDebounce_Do-14        	 4025568	       393.9 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	github.com/floatdrop/debounce/v2	8.574s
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
