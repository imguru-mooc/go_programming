package main

import (
	"context"
	"fmt"
	"time"
)

func worker(ctx context.Context, id int) {
	fmt.Printf("[worker %d] 시작\n", id)
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("[worker %d] 중단: %v\n", id, ctx.Err())
			return
		case <-time.After(300 * time.Millisecond):
			fmt.Printf("[worker %d] 진행 중...\n", id)
		}
	}
}

// 자식 작업이 또 자식을 만드는 경우 — 전파 확인
func parentJob(ctx context.Context) {
	for i := 1; i <= 3; i++ {
		go worker(ctx, i)
	}

	<-ctx.Done()
	fmt.Println("[부모] 중단:", ctx.Err())
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	go parentJob(ctx)

	// 1초 후 모든 작업 취소
	time.Sleep(1 * time.Second)
	fmt.Println("--- cancel() 호출 ---")
	cancel()

	// 정리 시간
	time.Sleep(500 * time.Millisecond)
	fmt.Println("--- 메인 종료 ---")
}
