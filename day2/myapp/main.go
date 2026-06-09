package main

import (
	"fmt"
	"github.com/myname/myapp/greeter"
)

func main() {
	fmt.Println(greeter.Hello()) // ✅ OK
	//fmt.Println(greeter.greeting()) // ❌ 컴파일 에러 (비공개)

	cfg := greeter.Config{
		Name: "Go", // ✅ OK
		// secret: "abc",  // ❌ 컴파일 에러
	}
	_ = cfg
}
