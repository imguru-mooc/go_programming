package main

import (
	"fmt"
	"github.com/myname/multipkg/mathx"
	"github.com/myname/multipkg/stringx"
)

func main() {
	fmt.Println("GCD(24, 36) =", mathx.GCD(24, 36))
	fmt.Println("Reverse('안녕 Go') =", stringx.Reverse("안녕 Go"))
}
