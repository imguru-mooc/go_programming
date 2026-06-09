package main

import "fmt"

type Point struct {
	X, Y int
}

func (p Point) String() string {
	return fmt.Sprintf("(%d, %d)", p.X, p.Y)
}

func main() {
	p := Point{3, 4}
	fmt.Println(p) // (3, 4) — String() 자동 호출
}

/*
package main

import "fmt"

// 인터페이스 정의 — "Speak 메서드를 가진 모든 것"
type Speaker interface {
	Speak() string
}

// Dog 구조체
type Dog struct{}

func (d Dog) Speak() string {
	return "멍멍!"
}

// Cat 구조체
type Cat struct{}

func (c Cat) Speak() string {
	return "야옹!"
}

func main() {
	// 어디에도 "Dog는 Speaker이다"라고 명시하지 않았지만,
	// Speak() 메서드를 가지므로 자동으로 Speaker 인터페이스 만족
	var s Speaker

	s = Dog{}
	fmt.Println(s.Speak()) // 멍멍!

	s = Cat{}
	fmt.Println(s.Speak()) // 야옹!
}
*/
