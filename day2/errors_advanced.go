package main

import (
	"errors"
	"fmt"
)

var ErrInsufficientFunds = errors.New("잔액 부족")

type WithdrawError struct {
	AccountID string
	Requested float64
	Available float64
}

func (e *WithdrawError) Error() string {
	return fmt.Sprintf("계좌 %s에서 출금 실패: 요청=%.2f, 잔액=%.2f",
		e.AccountID, e.Requested, e.Available)
}

// Unwrap을 구현하면 errors.Is가 sentinel 비교 가능
func (e *WithdrawError) Unwrap() error {
	return ErrInsufficientFunds
}

func withdraw(accountID string, amount float64, balance float64) error {
	if amount > balance {
		return &WithdrawError{
			AccountID: accountID,
			Requested: amount,
			Available: balance,
		}
	}
	return nil
}

func safeDivide(a, b int) (result int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("패닉 복구: %v", r)
		}
	}()
	return a / b, nil // b=0이면 panic 발생
}

func main() {
	// 1. 사용자 정의 에러 + 래핑
	err := withdraw("ACC-001", 5000, 1000)
	if err != nil {
		fmt.Println(err)

		// sentinel 검사
		if errors.Is(err, ErrInsufficientFunds) {
			fmt.Println("→ 잔액 부족으로 분류")
		}

		// 타입 추출
		var we *WithdrawError
		if errors.As(err, &we) {
			fmt.Printf("→ 부족액: %.2f\n", we.Requested-we.Available)
		}
	}

	// 2. panic / recover
	fmt.Println()
	if _, err := safeDivide(10, 0); err != nil {
		fmt.Println("safeDivide:", err)
	}
	if result, err := safeDivide(10, 2); err == nil {
		fmt.Println("safeDivide:", result)
	}
}
