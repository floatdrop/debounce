package debounce_test

import (
	"testing"
	"time"

	"github.com/floatdrop/debounce/v2"
)

func BenchmarkDebounce_Do(b *testing.B) {
	debouncer := debounce.New(debounce.WithDelay(100 * time.Millisecond))
	f := func() {}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		debouncer.Do(f)
	}
	b.StopTimer()
	debouncer.Close()
}
