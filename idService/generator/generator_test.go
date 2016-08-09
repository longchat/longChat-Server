package generator

import (
	"math/rand"
	"testing"
)

func BenchmarkGenerate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		var a uint64 = uint64(rand.Intn(1000))
		b.StartTimer()
		generate(a)
	}
}
