package main

import (
    "context"
    "errors"
    "fmt"
    "time"
)

type ctxKey string

const (
    requestIDKey ctxKey = "request-id"
    userIDKey    ctxKey = "user-id"
)

// 헬퍼 함수들
func WithRequestID(ctx context.Context, id string) context.Context {
    return context.WithValue(ctx, requestIDKey, id)
}

func RequestIDFrom(ctx context.Context) string {
    if v, ok := ctx.Value(requestIDKey).(string); ok {
        return v
    }
    return "unknown"
}

// 깊은 단계의 함수 - ctx-aware
func dbQuery(ctx context.Context) error {
    reqID := RequestIDFrom(ctx)
    fmt.Printf("[%s] DB 쿼리 시작\n", reqID)

    select {
    case <-time.After(2 * time.Second):  // 느린 쿼리 시뮬레이션
        fmt.Printf("[%s] DB 쿼리 완료\n", reqID)
        return nil
    case <-ctx.Done():
        fmt.Printf("[%s] DB 쿼리 중단: %v\n", reqID, ctx.Err())
        return ctx.Err()
    }
}

// 중간 계층
func service(ctx context.Context) error {
    reqID := RequestIDFrom(ctx)
    fmt.Printf("[%s] 서비스 호출\n", reqID)
    return dbQuery(ctx)
}

// 핸들러 - 1초 데드라인 설정
func handler(ctx context.Context) error {
    ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
    defer cancel()

    return service(ctx)
}

func main() {
    // 요청 1: 정상 (느린 작업이지만 데드라인 짧음)
    ctx1 := WithRequestID(context.Background(), "req-001")
    if err := handler(ctx1); err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            fmt.Println("→ 응답: 504 Gateway Timeout")
        }
    }

    fmt.Println()

    // 요청 2: 수동 취소
    ctx2, cancel := context.WithCancel(WithRequestID(context.Background(), "req-002"))
    go func() {
        time.Sleep(300 * time.Millisecond)
        cancel()
    }()

    if err := handler(ctx2); err != nil {
        fmt.Printf("종료: %v\n", err)
    }
}
