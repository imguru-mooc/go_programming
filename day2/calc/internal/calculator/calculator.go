package calculator

import (
	"errors"
	"fmt"
	"math"
)

// ErrDivByZero는 0으로 나눌 때 반환된다
var ErrDivByZero = errors.New("0으로 나눌 수 없습니다")

// Calculate는 연산자와 두 피연산자를 받아 결과를 반환한다
func Calculate(op string, a, b float64) (float64, error) {
	switch op {
	case "add":
		return a + b, nil
	case "sub":
		return a - b, nil
	case "mul":
		return a * b, nil
	case "div":
		if b == 0 {
			return 0, fmt.Errorf("a=%v, b=%v: %w", a, b, ErrDivByZero)
		}
		return a / b, nil
	case "mod":
		if b == 0 {
			return 0, fmt.Errorf("a=%v, b=%v: %w", a, b, ErrDivByZero)
		}
		return math.Mod(a, b), nil
	default:
		return 0, fmt.Errorf("지원하지 않는 연산: %q (지원: add, sub, mul, div)", op)
	}
}
