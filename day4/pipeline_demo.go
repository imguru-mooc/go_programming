package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Stage 1: 정수 생성
func gen(ctx context.Context, n int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for i := 1; i <= n; i++ {
			select {
			case out <- i:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// Stage 2: 제곱 (느린 작업)
func square(ctx context.Context, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			time.Sleep(50 * time.Millisecond)
			select {
			case out <- n * n:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// Fan-in
func merge(ctx context.Context, chs ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup
	wg.Add(len(chs))

	for _, c := range chs {
		go func(c <-chan int) {
			defer wg.Done()
			for v := range c {
				select {
				case out <- v:
				case <-ctx.Done():
					return
				}
			}
		}(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 입력
	input := gen(ctx, 20)

	// Fan-out: 4개 워커로 square 처리
	workers := make([]<-chan int, 4)
	for i := range workers {
		workers[i] = square(ctx, input)
	}

	// Fan-in
	out := merge(ctx, workers...)

	// 결과
	total := 0
	count := 0
	for v := range out {
		total += v
		count++
	}
	fmt.Printf("처리한 개수: %d, 합계: %d\n", count, total)

	if ctx.Err() != nil {
		fmt.Println("종료 사유:", ctx.Err())
	}
}
