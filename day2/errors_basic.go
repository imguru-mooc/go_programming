package main

import (
	"errors"
	"fmt"
)

type DivisionError struct {
	Dividend float64
	Divisor  float64
}

func (e *DivisionError) Error() string {
	return fmt.Sprintf("나눗셈 오류: %v / %v", e.Dividend, e.Divisor)
}

func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, &DivisionError{Dividend: a, Divisor: b}
	}
	return a / b, nil
}

func main() {
	// 케이스 1: errors.New
	err1 := errors.New("단순 에러")
	fmt.Println(err1)

	// 케이스 2: fmt.Errorf
	err2 := fmt.Errorf("값 %d는 허용 범위(%d~%d)를 벗어남", 200, 0, 100)
	fmt.Println(err2)

	// 케이스 3: 사용자 정의 에러
	_, err := divide(10, 0)
	if err != nil {
		fmt.Println(err)
	}

	// 케이스 4: 정상 동작
	result, err := divide(10, 2)
	if err != nil {
		fmt.Println("에러:", err)
		return
	}
	fmt.Println("결과:", result)
}
