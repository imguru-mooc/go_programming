package sortdemo

import (
	"slices"
	"sort"
	"testing"
)

func TestSlowSort(t *testing.T) {
	data := []int{5, 2, 9, 1, 7}
	got := slowSort(data)
	if !slices.IsSorted(got) {
		t.Errorf("정렬 안 됨: %v", got)
	}
}

func BenchmarkSlowSort(b *testing.B) {
	src := randomData(2000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		data := slices.Clone(src) // 매번 동일한 무작위 입력
		b.StartTimer()
		slowSort(data)
	}
}

func BenchmarkStdSort(b *testing.B) {
	src := randomData(2000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		data := slices.Clone(src)
		b.StartTimer()
		sort.Ints(data)
	}
}
