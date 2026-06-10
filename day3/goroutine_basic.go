package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

func main() {
	fmt.Println("CPU 수:", runtime.NumCPU())
	fmt.Println("시작 시 고루틴 수:", runtime.NumGoroutine())

	var wg sync.WaitGroup

	// 100만 개 고루틴 생성 — 진짜 가능한지 확인!
	start := time.Now()
	const N = 1_000_000

	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_ = id * 2 // 간단한 작업
		}(i)
	}

	fmt.Println("최대 고루틴 수:", runtime.NumGoroutine())
	wg.Wait()
	fmt.Printf("%d개 고루틴 완료: %v\n", N, time.Since(start))
	fmt.Println("종료 시 고루틴 수:", runtime.NumGoroutine())
}
