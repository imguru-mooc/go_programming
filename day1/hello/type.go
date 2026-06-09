package main

import "fmt"

func main() {
	/*
		var i int = 10
		var f float64 = 3.14

		//result := i + f // ❌ 컴파일 에러!
		result := float64(i) + f
		fmt.Println(result)
	*/

	/*
		s := "안녕하세요, Go!"

		//fmt.Println(len(s))       // 20 (바이트 수)
		fmt.Println(len([]rune(s))) // 10 (문자 수)
	*/
	s := "Hello, Go!"
	fmt.Println(s[0:5]) // "Hello"

	s2 := s + " World"
	fmt.Println(s2) // "Hello, Go! World"

	for i, r := range "한글" {
		fmt.Printf("index=%d rune=%c\n", i, r)
	}
}
