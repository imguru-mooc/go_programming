package main

import "fmt"

// 두 값을 교환
func swap(a, b *int) {
	*a, *b = *b, *a
}

// 안전한 나눗셈 — 다중 반환값
func safeDiv(a, b int) (int, error) {
	if b == 0 {
		return 0, fmt.Errorf("0으로 나눌 수 없습니다")
	}
	return a / b, nil
}

func main() {
	// 1. 포인터로 교환
	x, y := 10, 20
	swap(&x, &y)
	fmt.Println("after swap:", x, y) // 20 10

	// 2. 다중 반환값 + 에러 처리
	result, err := safeDiv(10, 0)
	if err != nil {
		fmt.Println("에러:", err)
	} else {
		fmt.Println("결과:", result)
	}

	result, err = safeDiv(10, 3)
	if err != nil {
		fmt.Println("에러:", err)
	} else {
		fmt.Println("결과:", result)
	}
}
