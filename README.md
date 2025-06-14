# Debounce

[![CI](https://github.com/floatdrop/debounce/actions/workflows/ci.yaml/badge.svg)](https://github.com/floatdrop/debounce/actions/workflows/ci.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/floatdrop/debounce)](https://goreportcard.com/report/github.com/floatdrop/debounce)
[![Go Coverage](https://github.com/floatdrop/debounce/wiki/coverage.svg)](https://raw.githack.com/wiki/floatdrop/debounce/coverage.html)
[![Go Reference](https://pkg.go.dev/badge/github.com/floatdrop/debounce.svg)](https://pkg.go.dev/github.com/floatdrop/debounce)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A simple, thread-safe debounce library for Go that delays function execution until after a specified duration has elapsed since the last invocation. Perfect for rate limiting, reducing redundant operations, and optimizing performance in high-frequency scenarios.

## Features

- **Zero allocations**: No allocations on sunbsequent debounce calls
- **Thread-safe**: Safe for concurrent use across multiple goroutines
- **Configurable delays**: Set custom debounce durations
- **Call limits**: Execute immediately after a maximum number of calls
- **Time limits**: Execute immediately after a maximum wait time
- **Function flexibility**: Each call can provide a different function to execute
- **Zero dependencies**: Built using only Go standard library

## Installation

```bash
go get github.com/floatdrop/debounce
```

## Usage

https://github.com/floatdrop/debounce/blob/770f96180424dabfea45ca421cce5aa8e57a46f5/example_test.go#L29-L43

## Benchmarks

```bash
go test -bench=BenchmarkSingleCall -benchmem
```

| Benchmark                        | Iterations | Time per Op  | Bytes per Op | Allocs per Op |
|----------------------------------|------------|--------------|--------------|---------------|
| BenchmarkSingleCall-14           | 47227514   | 25.24 ns/op  | 0 B/op       |  0 allocs/op  |

- ~25ns per debounced call
- Constant memory usage regardless of call frequency

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
