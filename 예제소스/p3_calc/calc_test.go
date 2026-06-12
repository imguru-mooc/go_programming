package calc

import (
	"fmt"
	"testing"
)

func TestBasicOps(t *testing.T) {
	tests := []struct {
		name string
		op   func(int, int) int
		a, b int
		want int
	}{
		{"Add 양수", Add, 2, 3, 5},
		{"Add 음수", Add, -1, -1, -2},
		{"Sub", Sub, 10, 4, 6},
		{"Mul", Mul, 6, 7, 42},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.op(tc.a, tc.b); got != tc.want {
				t.Errorf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestDivByZero(t *testing.T) {
	_, err := Div(10, 0)
	if err == nil {
		t.Error("0으로 나눴는데 에러가 없음")
	}
}

func BenchmarkAdd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Add(2, 3)
	}
}

func ExampleAdd() {
	fmt.Println(Add(2, 3))
	// Output: 5
}
