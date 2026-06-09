package calculator

import "fmt"

func Add(a int, b int) int {
	return a + b
}

func Mul(a int, b int) int {
	return a * b
}

func Calculate(op string, a int, b int) (int, error) {
	switch op {
	case "add":
		return a + b, nil
	case "mul":
		return a * b, nil
	default:
		return 0, fmt.Errorf("지원하지 않는 연산입니다: %s", op)
	}
}
