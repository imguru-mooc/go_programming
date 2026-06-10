package main

import (
	"fmt"
	"time"
)

// 작업자 — 수신 전용 채널
func worker(id int, jobs <-chan int, results chan<- int) {
	for j := range jobs {
		fmt.Printf("worker %d: %d 처리 중...\n", id, j)
		time.Sleep(100 * time.Millisecond)
		results <- j * 2
	}
	fmt.Printf("worker %d: 종료\n", id)
}

func main() {
	jobs := make(chan int)
	results := make(chan int)

	// 워커 3개 시작
	for w := 1; w <= 3; w++ {
		go worker(w, jobs, results)
	}

	// 작업 5개 송신 후 채널 닫기
	go func() {
		for j := 1; j <= 5; j++ {
			jobs <- j
		}
		close(jobs) // 더 이상 작업 없음 알림
	}()

	// 결과 5개 수신
	for r := 1; r <= 5; r++ {
		v := <-results
		fmt.Println("결과:", v)
	}
}
