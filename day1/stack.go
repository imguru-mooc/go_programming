package main

import (
	"errors"
	"fmt"
)

// =======================
// int 전용 Stack
// =======================

type IntStack struct {
	items []int
}

func NewIntStack() *IntStack {
	return &IntStack{
		items: make([]int, 0, 8),
	}
}

func (s *IntStack) Push(v int) {
	s.items = append(s.items, v)
}

func (s *IntStack) Pop() (int, error) {
	if len(s.items) == 0 {
		return 0, errors.New("스택이 비어있음")
	}

	top := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]

	return top, nil
}

func (s *IntStack) Peek() (int, bool) {
	if len(s.items) == 0 {
		return 0, false
	}

	return s.items[len(s.items)-1], true
}

func (s *IntStack) Len() int {
	return len(s.items)
}

func (s *IntStack) IsEmpty() bool {
	return len(s.items) == 0
}

func main() {
	// int 스택
	s := NewIntStack()

	s.Push(10)
	s.Push(20)
	s.Push(30)

	if top, ok := s.Peek(); ok {
		fmt.Println("Peek:", top)
	}

	for !s.IsEmpty() {
		v, _ := s.Pop()
		fmt.Println("Pop:", v)
	}
}
