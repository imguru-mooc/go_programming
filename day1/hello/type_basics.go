package main

import "fmt"

func main() {
	var a int = 10
	var b = 20
	c := 30

	var d int
	var e string

	fmt.Printf("a=%d, b=%d, c=%d\n", a, b, c)
	fmt.Printf("d=%d, e=%q\n", d, e)

	var x int = 100
	var y float64 = float64(x) * 1.5
	fmt.Printf("y=%f\n", y)

	const (
		Sunday = (iota + 1) * 10
		Monday
		Tuesday
	)
	fmt.Println(Sunday, Monday, Tuesday)

	s := "Hello, 한글!"
	fmt.Println("바이트 길이:", len(s))
	fmt.Println("문자   길이:", len([]rune(s)))
}
