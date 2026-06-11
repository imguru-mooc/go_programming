package main

import (
    "fmt"
    "time"
)

// 패턴 1: 이벤트 결합 (Coalescing)
func coalesceDemo() {
    fmt.Println("\n=== Pattern 1: Event Coalescing ===")
    notify := make(chan struct{}, 1)

    // 빠르게 10번 트리거 — 하나만 큐에 남음
    for i := 0; i < 10; i++ {
        select {
        case notify <- struct{}{}:
            fmt.Println("알림 큐잉됨")
        default:
            fmt.Println("이미 대기 중 — 드롭")
        }
    }

    // 소비
    <-notify
    fmt.Println("알림 소비")
}

// 패턴 2: nil 채널로 case 토글
func toggleDemo() {
    fmt.Println("\n=== Pattern 2: nil Channel Toggle ===")
    in := make(chan int)
    var out chan int  // nil

    go func() {
        for i := 1; i <= 5; i++ {
            in <- i
        }
        close(in)
    }()

    var pending int
    hasPending := false

    for {
        var send chan<- int
        if hasPending {
            send = out
            // 실제 소비자가 없으니 시뮬레이션
            // 여기서는 그냥 출력
            fmt.Printf("송신 준비: %d\n", pending)
            hasPending = false
            continue
        }

        select {
        case v, ok := <-in:
            if !ok {
                return
            }
            pending = v
            hasPending = true
        case send <- pending:
            // 실제론 여기서 소비자에게 전달됨
        }
    }
}

// 패턴 3: for-select with shutdown
func workerDemo() {
    fmt.Println("\n=== Pattern 3: for-select Worker ===")
    jobs := make(chan int)
    done := make(chan struct{})

    go func() {
        for {
            select {
            case job, ok := <-jobs:
                if !ok {
                    fmt.Println("워커: 채널 닫힘, 종료")
                    return
                }
                fmt.Printf("워커: 작업 %d 처리\n", job)
            case <-done:
                fmt.Println("워커: 강제 종료 신호")
                return
            }
        }
    }()

    for i := 1; i <= 3; i++ {
        jobs <- i
    }
    time.Sleep(100 * time.Millisecond)
    close(done)
    time.Sleep(100 * time.Millisecond)
}

func main() {
    coalesceDemo()
    workerDemo()
}
