package main

import "fmt"

type Rectangle struct {
	Width, Height float64
}

// 메서드 — Rectangle 타입에 Area라는 함수를 붙임
func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

func main() {
	rect := Rectangle{Width: 3, Height: 4}
	fmt.Println(rect.Width)  //3
	fmt.Println(rect.Height) //4
	fmt.Println(rect.Area()) //4
}
