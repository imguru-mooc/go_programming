package main

import (
	"fmt"
	"sync"
	"time"
)

// 생성자
func generate(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			out <- n
		}
	}()
	return out
}

// 변환 단계
func square(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			out <- n * n
		}
	}()
	return out
}

// Fan-in
func merge(chs ...<-chan int) <-chan int {
	var wg sync.WaitGroup
	out := make(chan int)

	output := func(c <-chan int) {
		defer wg.Done()
		for v := range c {
			out <- v
		}
	}

	wg.Add(len(chs))
	for _, c := range chs {
		go output(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func main() {
	// 파이프라인: 생성 → 제곱 → 수집
	in := generate(1, 2, 3, 4, 5)

	// Fan-out: 두 워커가 제곱 처리
	c1 := square(in)
	// 주의: 같은 in을 두 번 못 씀 - 별도 생성
	c2 := square(generate(6, 7, 8, 9, 10))

	// Fan-in: 결과 합치기
	for v := range merge(c1, c2) {
		fmt.Println(v)
	}

	// select + timeout 데모
	timeoutDemo()
}

func timeoutDemo() {
	ch := make(chan string)
	go func() {
		time.Sleep(1 * time.Second)
		ch <- "느린 응답"
	}()

	select {
	case v := <-ch:
		fmt.Println("받음:", v)
	case <-time.After(2 * time.Second):
		fmt.Println("타임아웃!")
	}
}
