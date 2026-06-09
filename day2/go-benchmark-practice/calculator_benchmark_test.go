package calculator

import "testing"

func BenchmarkAdd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Add(10, 20)
	}
}

func BenchmarkMul(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Mul(10, 20)
	}
}

func BenchmarkCalculateAdd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = Calculate("add", 10, 20)
	}
}

func BenchmarkCalculate(b *testing.B) {
	benchmarks := []struct {
		name string
		op   string
		a    int
		b    int
	}{
		{"add", "add", 10, 20},
		{"mul", "mul", 10, 20},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = Calculate(bm.op, bm.a, bm.b)
			}
		})
	}
}

func BenchmarkCalculateWithResetTimer(b *testing.B) {
	op := "add"
	a := 10
	c := 20

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = Calculate(op, a, c)
	}
}
