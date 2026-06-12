// 5.9 — runtime/trace (전체 코드)
package main

import (
	"fmt"
	"log"
	"os"
	"runtime/trace"
	"sync"
	"time"
)

func worker(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	sum := 0
	for i := 0; i < 50_000_000; i++ { // CPU 작업
		sum += i
	}
	time.Sleep(10 * time.Millisecond) // I/O 흉내
	_ = sum
}

func main() {
	f, err := os.Create("trace.out")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if err := trace.Start(f); err != nil {
		log.Fatal(err)
	}
	defer trace.Stop()

	// 측정 대상: 고루틴 4개의 동시 작업
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go worker(i, &wg)
	}
	wg.Wait()
	fmt.Println("완료 — trace.out 생성됨")
}
