package main

import "fmt"

type Student struct {
	Name   string
	Scores []int
}

func average(scores []int) float64 {
	if len(scores) == 0 {
		return 0
	}
	sum := 0
	for _, s := range scores {
		sum += s
	}
	return float64(sum) / float64(len(scores))
}

func main() {
	students := []Student{
		{"Alice", []int{90, 85, 92}},
		{"Bob", []int{70, 65, 80}},
		{"Carol", []int{95, 100, 88}},
	}

	for _, s := range students {
		avg := average(s.Scores)
		fmt.Printf("%s: %.2f\n", s.Name, avg)
	}
}
