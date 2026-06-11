package checker

import (
    "context"
    "fmt"
    "net/http"
    "time"
)

// Target은 검사 대상
type Target struct {
    URL string
}

// Result는 검사 결과
type Result struct {
    URL        string
    StatusCode int
    Latency    time.Duration
    Err        error
    Attempts   int
}

func (r Result) String() string {
    if r.Err != nil {
        return fmt.Sprintf("❌ %s — 에러: %v, (시도 %v회)", r.URL, r.Err, r.Attempts)
    }
    return fmt.Sprintf("✅ %s — %d (%v, 시도 %v회)", r.URL, r.StatusCode, r.Latency, r.Attempts)
}

// Check은 단일 URL을 검사한다
func Check(ctx context.Context, url string) Result {
    start := time.Now()

    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return Result{URL: url, Err: err}
    }

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return Result{URL: url, Err: err, Latency: time.Since(start)}
    }
    defer resp.Body.Close()

    return Result{
        URL:        url,
        StatusCode: resp.StatusCode,
        Latency:    time.Since(start),
    }
}

type RetryConfig struct {
    MaxAttempts int
    InitialDelay time.Duration
}

var DefaultRetry = RetryConfig{
    MaxAttempts:  3,
    InitialDelay: 200 * time.Millisecond,
}

// CheckWithRetry는 실패 시 지수 백오프로 재시도
func CheckWithRetry(ctx context.Context, url string, cfg RetryConfig) Result {
    var result Result
    delay := cfg.InitialDelay

    for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
        result = Check(ctx, url)
        result.Attempts = attempt

        // 성공
        if result.Err == nil {
            return result
        }

        // 4xx는 재시도 의미 없음
        if result.StatusCode >= 400 && result.StatusCode < 500 {
            return result
        }

        // 마지막 시도였으면 종료
        if attempt == cfg.MaxAttempts {
            return result
        }

        // 백오프 대기
        select {
        case <-time.After(delay):
        case <-ctx.Done():
            result.Err = ctx.Err()
            return result
        }
        delay *= 2
    }
    return result
}
