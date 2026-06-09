package calculator

import "testing"

func TestAdd(t *testing.T) {
	result := Add(10, 20)

	if result != 30 {
		t.Errorf("원하는 값: %d, 실제 값: %d", 30, result)
	}
}

func TestMul(t *testing.T) {
	result := Mul(10, 20)

	if result != 200 {
		t.Errorf("원하는 값: %d, 실제 값: %d", 200, result)
	}
}

func TestCalculate(t *testing.T) {
	result, err := Calculate("add", 10, 20)

	if err != nil {
		t.Errorf("에러가 발생하면 안 됩니다: %v", err)
	}

	if result != 30 {
		t.Errorf("원하는 값: %d, 실제 값: %d", 30, result)
	}
}
