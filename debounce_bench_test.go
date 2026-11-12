package debounce_test

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/floatdrop/debounce/v3"
)

func benchmarkDo(b *testing.B, opts ...debounce.Option) {
	var n atomic.Int64
	d := debounce.New(opts...)
	defer d.Stop()
	fn := func() { n.Add(1) }

	b.ReportAllocs()
	for b.Loop() {
		d.Do(fn)
	}
}

func BenchmarkDo_NoOptions(b *testing.B) {
	benchmarkDo(b)
}

func BenchmarkDo_Leading(b *testing.B) {
	benchmarkDo(b, debounce.WithLeading(true))
}

func BenchmarkDo_Trailing(b *testing.B) {
	benchmarkDo(b, debounce.WithTrailing(time.Millisecond))
}

func BenchmarkDo_LeadingAndTrailing(b *testing.B) {
	benchmarkDo(b, debounce.WithLeading(true), debounce.WithTrailing(time.Millisecond))
}

func BenchmarkDo_DebounceLimit(b *testing.B) {
	benchmarkDo(b, debounce.WithDebounceLimit(10))
}

func BenchmarkDo_TimeLimit(b *testing.B) {
	benchmarkDo(b, debounce.WithTimeLimit(time.Millisecond))
}

func BenchmarkDo_AllOptions(b *testing.B) {
	benchmarkDo(b,
		debounce.WithLeading(true),
		debounce.WithTrailing(time.Millisecond),
		debounce.WithDebounceLimit(10),
		debounce.WithTimeLimit(time.Millisecond),
	)
}

func BenchmarkDo_Parallel(b *testing.B) {
	var n atomic.Int64
	d := debounce.New(debounce.WithTrailing(time.Millisecond))
	defer d.Stop()
	fn := func() { n.Add(1) }

	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			d.Do(fn)
		}
	})
}
