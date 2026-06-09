package main

import (
	"fmt"
	"strconv"
)

func fizzbuzz(n int) string {
	s := ""
	if n%3 == 0 {
		s += "Fizz"
	}
	if n%5 == 0 {
		s += "Buzz"
	}
	if s == "" {
		s = strconv.Itoa(n)
	}

	return s
}

func main() {
	for i := 1; i <= 30; i++ {
		fmt.Println(fizzbuzz(i))
	}
}

/*
package main

import (
	"fmt"
	"strconv"
)

func fizzbuzz(n int) string {
	switch {
	case n%15 == 0:
		return "FizzBuzz"
	case n%3 == 0:
		return "Fizz"
	case n%5 == 0:
		return "Buzz"
	default:
		return strconv.Itoa(n)
	}
}

func main() {
	for i := 1; i <= 30; i++ {
		fmt.Println(fizzbuzz(i))
	}
}
*/

/*
package main

import "fmt"

func main() {
	for i := 1; i <= 30; i++ {
		switch {
		case i%15 == 0:
			fmt.Println("FizzBuzz")
		case i%3 == 0:
			fmt.Println("Fizz")
		case i%5 == 0:
			fmt.Println("Buzz")
		default:
			fmt.Println(i)
		}
	}
}
*/
