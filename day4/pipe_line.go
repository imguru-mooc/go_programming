package main

import (
	"fmt"
	"sync"
)

// Stage 1: 정수 생성
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

// Stage 2: 제곱
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

// Stage 3: 합계 누적 출력
func runningSum(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        sum := 0
        for n := range in {
            sum += n
            out <- sum
        }
    }()
    return out
}

// Fan-out: 같은 입력 채널을 N개 워커가 소비
// Fan-in: N개 출력을 하나로 합침

func merge(channels ...<-chan int) <-chan int {
    out := make(chan int)
    var wg sync.WaitGroup

    output := func(c <-chan int) {
        defer wg.Done()
        for v := range c {
            out <- v
        }
    }

    wg.Add(len(channels))
    for _, c := range channels {
        go output(c)
    }

    go func() {
        wg.Wait()
        close(out)
    }()

    return out
}

func main() {
    in := generate(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)

    // Fan-out: 3개 워커가 동시에 square
    c1 := square(in)
    c2 := square(in)  // 같은 in을 공유 — 자동 부하 분산
    c3 := square(in)

    // Fan-in
    for v := range merge(c1, c2, c3) {
        fmt.Println(v)
    }
}
