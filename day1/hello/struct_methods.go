package main

import "fmt"

type Account struct {
	Owner   string
	Balance float64
}

// 잔액 조회 — 값 리시버
func (a Account) GetBalance() float64 {
	return a.Balance
}

// 입금 — 포인터 리시버 (수정 필요)
func (a *Account) Deposit(amount float64) {
	a.Balance += amount
}

// 출금 — 포인터 리시버 + 에러 반환
func (a *Account) Withdraw(amount float64) error {
	if amount > a.Balance {
		return fmt.Errorf("잔액 부족: 요청 %.2f, 잔액 %.2f", amount, a.Balance)
	}
	a.Balance -= amount
	return nil
}

func main() {
	acc := Account{Owner: "Alice", Balance: 1000}

	acc.Deposit(500)
	fmt.Printf("입금 후: %.2f\n", acc.GetBalance())

	if err := acc.Withdraw(2000); err != nil {
		fmt.Println("출금 실패:", err)
	}

	if err := acc.Withdraw(300); err != nil {
		fmt.Println("출금 실패:", err)
	} else {
		fmt.Printf("출금 후: %.2f\n", acc.GetBalance())
	}
}
