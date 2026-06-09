package main

import "fmt"

func main() {
	s := []int{1, 2, 3}    // 배열 선언과 비슷하지만 [] 안에 숫자 없음
	s = append(s, 4)       // [1 2 3 4]
	s = append(s, 5, 6, 7) // [1 2 3 4 5 6 7]

	fmt.Println(len(s)) // 7 — 길이
	fmt.Println(cap(s)) // 8 (구현에 따라 다름) — 용량
}

/*
func modify(a [3]int) {
	a[0] = 999 // 복사본을 수정 — 원본은 그대로!
}
func main() {
	arr := [3]int{1, 2, 3}
	modify(arr)
	fmt.Println(arr) // [1 2 3] — 안 변함!
}
*/
