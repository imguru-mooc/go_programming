package calculator

import (
	"errors"
	"testing"
)

func TestCalculate_BasicOps(t *testing.T) {
	tests := []struct {
		name     string
		op       string
		a, b     float64
		expected float64
	}{
		{"add", "add", 3, 5, 8},
		{"sub", "sub", 10, 4, 6},
		{"mul", "mul", 6, 7, 42},
		{"div", "div", 10, 2, 5},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Calculate(tc.op, tc.a, tc.b)
			if err != nil {
				t.Fatalf("예상치 못한 에러: %v", err)
			}
			if got != tc.expected {
				t.Errorf("기대값 %v, 받은 값 %v", tc.expected, got)
			}
		})
	}
}

func TestCalculate_DivByZero(t *testing.T) {
	_, err := Calculate("div", 10, 0)
	if !errors.Is(err, ErrDivByZero) {
		t.Errorf("ErrDivByZero를 기대했으나 %v를 받음", err)
	}
}

func TestCalculate_UnknownOp(t *testing.T) {
	_, err := Calculate("pow", 2, 3)
	if err == nil {
		t.Error("에러가 발생해야 함")
	}
}

func BenchmarkCalculate(b *testing.B) {
	benchmarks := []struct {
		name string
		op   string
		a, b float64
	}{
		{"add", "add", 1.5, 2.5},
		{"mul", "mul", 12.34, 56.78},
		{"div", "div", 100.0, 3.0},
		{"mod", "mod", 100.0, 7.0},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = Calculate(bm.op, bm.a, bm.b)
			}
		})
	}
}

// 단순 함수 직접 호출과 비교
func BenchmarkDirectAdd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = 1.5 + 2.5
	}
}
