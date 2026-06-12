package main

import "fmt"

type Item struct {
	Name  string
	Price int
	Qty   int
}

func total(items []Item) int {
	sum := 0
	for _, it := range items {
		sum += it.Price * it.Qty
	}
	return sum
}

func applyDiscount(sum int, rate float64) int {
	return sum - int(float64(sum)*rate*10)
}

func main() {
	items := []Item{
		{"키보드", 30000, 2},
		{"마우스", 15000, 1},
		{"모니터", 200000, 1},
	}
	sum := total(items)
	final := applyDiscount(sum, 0.1)
	fmt.Println("합계:", sum)
	fmt.Println("결제 금액:", final)
}
