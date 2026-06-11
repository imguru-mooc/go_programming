package main

import (
    "fmt"
    "math/rand"
    "time"
)

// 느린 작업 시뮬레이션
func slowQuery(id int) string {
    time.Sleep(time.Duration(rand.Intn(3000)) * time.Millisecond)
    return fmt.Sprintf("결과-%d", id)
}

// 타임아웃이 있는 호출
func queryWithTimeout(id int, timeout time.Duration) (string, error) {
    ch := make(chan string, 1)
    go func() {
        ch <- slowQuery(id)
    }()

    select {
    case r := <-ch:
        return r, nil
    case <-time.After(timeout):
        return "", fmt.Errorf("타임아웃 (id=%d, %v)", id, timeout)
    }
}

// 주기적 모니터
func monitor(done <-chan struct{}) {
    ticker := time.NewTicker(500 * time.Millisecond)
    defer ticker.Stop()

    count := 0
    for {
        select {
        case <-ticker.C:
            count++
            fmt.Printf("모니터 틱 #%d\n", count)
        case <-done:
            fmt.Println("모니터 종료")
            return
        }
    }
}

func main() {
    rand.Seed(time.Now().UnixNano())

    // 1. 타임아웃 있는 호출들
    for i := 1; i <= 5; i++ {
        result, err := queryWithTimeout(i, 1500*time.Millisecond)
        if err != nil {
            fmt.Printf("[%d] 실패: %v\n", i, err)
        } else {
            fmt.Printf("[%d] 성공: %s\n", i, result)
        }
    }

    // 2. 주기적 모니터
    fmt.Println("\n--- Ticker 데모 ---")
    done := make(chan struct{})
    go monitor(done)
    time.Sleep(2 * time.Second)
    close(done)
    time.Sleep(100 * time.Millisecond)
}
