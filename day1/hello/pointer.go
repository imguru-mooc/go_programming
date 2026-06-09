package main

import "fmt"

func main() {
	p := new(int)
	fmt.Println(*p)
	*p = 42
	fmt.Println(*p)
}

/*
package main

import "fmt"

func zero(p *int) {
	p[0] = 0 // *(p+0) = 0
}

func main() {
	x := [4]int{1, 2, 3, 4}
	//x[0] = 0
	zero(&x[0])
	fmt.Println(x[0])
}
*/

/*
package main

import "fmt"

func zero(p *int) {
	*p = 0
}

func main() {
	x := 100
	zero(&x)
	fmt.Println(x) // 0
}
*/
