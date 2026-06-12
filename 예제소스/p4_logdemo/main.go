package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func main() {
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, opts)))

	mux := http.NewServeMux()
	mux.HandleFunc("/", logMiddleware(handler))

	slog.Info("서버 시작", "addr", ":8080")
	http.ListenAndServe(":8080", mux)
}

func logMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = "auto-" + time.Now().Format("150405")
		}

		logger := slog.With(
			"request_id", reqID,
			"method", r.Method,
			"path", r.URL.Path,
		)
		ctx := context.WithValue(r.Context(), ctxLoggerKey, logger)

		logger.Info("요청 수신")
		next(w, r.WithContext(ctx))
		logger.Info("요청 완료", "duration_ms", time.Since(start).Milliseconds())
	}
}

type ctxKey string

const ctxLoggerKey ctxKey = "logger"

func handler(w http.ResponseWriter, r *http.Request) {
	logger := r.Context().Value(ctxLoggerKey).(*slog.Logger)
	logger.Info("처리 중", "step", "1")
	time.Sleep(100 * time.Millisecond)
	logger.Info("처리 중", "step", "2")
	w.Write([]byte("OK"))
}
