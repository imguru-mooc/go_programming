// 4.6 — runtime/debug.Stack으로 패닉 스택 로깅 (전체 코드)
package main

import (
	"log/slog"
	"os"
	"runtime/debug"
)

func mayPanic(data []int, idx int) (result int) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("패닉 복구",
				"panic", r,
				"stack", string(debug.Stack()))
			result = -1 // named return으로 복구값 지정
		}
	}()
	return data[idx] // idx가 범위 밖이면 panic
}

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))

	data := []int{10, 20, 30}
	slog.Info("정상 접근", "value", mayPanic(data, 1))
	slog.Info("범위 밖 접근", "value", mayPanic(data, 99)) // panic → 복구 → -1
	slog.Info("프로그램은 죽지 않고 계속 실행됨")
}
