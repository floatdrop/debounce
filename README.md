# Debounce

[![CI](https://github.com/floatdrop/debounce/actions/workflows/ci.yaml/badge.svg)](https://github.com/floatdrop/debounce/actions/workflows/ci.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/floatdrop/debounce)](https://goreportcard.com/report/github.com/floatdrop/debounce)
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

## Use Cases

- **Search inputs**: Debounce API calls while user types
- **Button clicks**: Prevent double-clicks and rapid submissions
- **File watchers**: Batch file system events
- **Auto-save**: Delay saving until user stops typing
- **Resize events**: Throttle expensive layout calculations
- **API rate limiting**: Control request frequency
- **Batch processing**: Collect operations before execution

## Performance

The debounce implementation uses:
- Mutex for thread safety
- Timer for scheduling
- Minimal memory allocation
- No external dependencies

Benchmark results on typical hardware:
- ~100ns per debounced call
- Constant memory usage regardless of call frequency
- Scales linearly with number of concurrent debouncers

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
