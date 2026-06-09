package main

import "fmt"

// Shape 인터페이스
type Shape interface {
	Area() float64
	Perimeter() float64
}

type Rectangle struct {
	Width, Height float64
}

func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

func (r Rectangle) Perimeter() float64 {
	return 2 * (r.Width + r.Height)
}

type Circle struct {
	Radius float64
}

func (c Circle) Area() float64 {
	return 3.14159 * c.Radius * c.Radius
}

func (c Circle) Perimeter() float64 {
	return 2 * 3.14159 * c.Radius
}

// 인터페이스만 보고 동작 — 다형성
func describe(s Shape) {
	fmt.Printf("Area=%.2f  Perimeter=%.2f\n", s.Area(), s.Perimeter())
}

func main() {
	shapes := []Shape{
		Rectangle{Width: 3, Height: 4},
		Circle{Radius: 5},
	}

	for _, s := range shapes {
		describe(s)
	}
}
