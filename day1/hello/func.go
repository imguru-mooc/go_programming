package main

import (
	"fmt"
	"os"
)

func main() {
	data, err := os.ReadFile("hello.txt")
	if err != nil {
		fmt.Println("파일 읽기 실패:", err)
		return
	}
	fmt.Println(string(data))
}

/*
func divide(a, b int) (quotient int, remainder int) {
	quotient = a / b
	remainder = a % b
	return // naked return — 자동으로 quotient, remainder 반환
}

func main() {
	quot, rem := divide(10, 3)
	fmt.Println(quot, rem) // 3 1
}
*/

/*
func divide(a, b int) (int, int) {
	return a / b, a % b
}

func main() {
	quot, rem := divide(10, 3)
	fmt.Println(quot, rem) // 3 1
}
*/

/*
// go
func add(a, b int) int {
	return a + b
}

func main() {
	fmt.Printf("%d\n", add(1, 2))
}
*/
