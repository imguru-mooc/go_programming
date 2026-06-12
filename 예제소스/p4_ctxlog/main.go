// 4.4 — Context로 로거 전달 (전체 코드)
package main

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
)

type ctxKey string

const loggerKey ctxKey = "logger"

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func LoggerFrom(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return l
	}
	return slog.Default() // 없으면 기본 로거 — nil panic 방지
}

func handler(w http.ResponseWriter, r *http.Request) {
	reqID := r.Header.Get("X-Request-ID")
	logger := slog.With("request_id", reqID)
	ctx := WithLogger(r.Context(), logger)

	doWork(ctx)
	w.Write([]byte("OK"))
}

func doWork(ctx context.Context) {
	log := LoggerFrom(ctx)
	log.Info("작업 처리") // 자동으로 request_id 포함
	deeperWork(ctx)
}

func deeperWork(ctx context.Context) {
	// 몇 단계를 내려가도 ctx만 전달하면 같은 request_id가 따라옴
	LoggerFrom(ctx).Info("깊은 곳의 작업")
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	// 서버 없이 핸들러만 직접 호출해 시연 (httptest 활용)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Request-ID", "req-123")
	w := httptest.NewRecorder()
	handler(w, req)
}
