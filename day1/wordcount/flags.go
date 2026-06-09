package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var (
		topCount = flag.Int("n", 10, "출력할 상위 단어 개수")
		minLen   = flag.Int("min", 1, "단어의 최소 길이")
	)
	flag.Parse()

	fmt.Println(flag.Args())
	fmt.Println(*topCount)
	fmt.Println(*minLen)
}
