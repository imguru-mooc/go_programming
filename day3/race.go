package main

import (
	"fmt"
	"sync"
)

var (
	counter int
	mu      sync.Mutex
)

func main() {
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer mu.Unlock()
			mu.Lock()
			counter++ // 보호 없음 — race!
		}()
	}
	wg.Wait()
	fmt.Println(counter) // 1000이 안 나옴!
}
