// 5.4 — CPU 프로파일 분석 대상 (전체 코드)
package sortdemo

import "math/rand"

// 의도적으로 느린 버블 정렬 — O(n²)
func slowSort(data []int) []int {
	n := len(data)
	for i := 0; i < n; i++ {
		for j := 0; j < n-i-1; j++ {
			if data[j] > data[j+1] {
				data[j], data[j+1] = data[j+1], data[j]
			}
		}
	}
	return data
}

// 비교용 — 표준 라이브러리 정렬을 쓰는 빠른 버전은 벤치마크에서
func randomData(n int) []int {
	data := make([]int, n)
	for i := range data {
		data[i] = rand.Intn(1_000_000)
	}
	return data
}
