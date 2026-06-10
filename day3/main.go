package main

import (
	"fmt"
	"runtime"
	"time"
)

func main() {
	fmt.Println("goroutines:", runtime.NumGoroutine())
	time.Sleep(100 * time.Millisecond)
}
