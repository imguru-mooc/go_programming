package main

import (
	"errors"
	"fmt"
)

func sqrt(x float64) (float64, error) {
	if x < 0 {
		return 0, errors.New("음수의 제곱근을 계산할 수 없습니다")
	}
	// 실제 계산...
	return 0, nil
}

func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("0으로 나눌 수 없습니다 (a=%v)", a)
	}
	return a / b, nil
}

func main() {
	if _, err := sqrt(-4); err != nil {
		fmt.Println("에러:", err)
	}

	if _, err := divide(10, 0); err != nil {
		fmt.Println("에러:", err)
	}
}
