package main

import "fmt"

func main() {
	// 1. 슬라이스 기본
	nums := []int{10, 20, 30}
	nums = append(nums, 40, 50)
	fmt.Println("nums:", nums)
	fmt.Println("len:", len(nums), "cap:", cap(nums))

	// 2. 슬라이싱
	sub := nums[1:4]
	fmt.Println("sub:", sub)

	// 3. 합계 함수
	fmt.Println("sum:", sum(nums))

	// 4. 맵으로 단어 빈도수
	text := []string{"go", "is", "fun", "go", "is", "fast", "go"}
	count := wordCount(text)
	for k, v := range count {
		fmt.Printf("%s: %d\n", k, v)
	}
}

func sum(s []int) int {
	total := 0
	for _, v := range s {
		total += v
	}
	return total
}

func wordCount(words []string) map[string]int {
	m := make(map[string]int)
	for _, w := range words {
		m[w]++
	}
	return m
}
