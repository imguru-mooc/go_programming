// 1.2 보충 — Context와 결합한 HTTP 요청 (전체 코드)
// 3초 타임아웃 자동 취소를 실제로 관찰할 수 있도록
// 지연 시간을 인자로 받는 로컬 서버 핸들러를 함께 띄웁니다.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

func fetchWithTimeout(url string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println("성공! 상태:", resp.StatusCode)
	return nil
}

func main() {
	// 느린 서버를 내장해서 시연 (실무에선 외부 URL)
	mux := http.NewServeMux()
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(5 * time.Second): // 5초 걸리는 응답
			w.Write([]byte("늦은 응답"))
		case <-r.Context().Done(): // 클라이언트가 끊으면 즉시 중단
			return
		}
	})
	mux.HandleFunc("/fast", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("빠른 응답"))
	})
	srv := &http.Server{Addr: ":8081", Handler: mux}
	go srv.ListenAndServe()
	time.Sleep(100 * time.Millisecond) // 서버 기동 대기

	fmt.Println("--- 빠른 엔드포인트 (3초 제한) ---")
	if err := fetchWithTimeout("http://localhost:8081/fast", 3*time.Second); err != nil {
		fmt.Println("에러:", err)
	}

	fmt.Println("--- 느린 엔드포인트 (3초 제한, 5초 걸림) ---")
	start := time.Now()
	err := fetchWithTimeout("http://localhost:8081/slow", 3*time.Second)
	fmt.Printf("%.1f초 후 결과: %v\n", time.Since(start).Seconds(), err)
	if errors.Is(err, context.DeadlineExceeded) {
		fmt.Println("→ context.DeadlineExceeded로 자동 취소 확인!")
	}
	srv.Close()
}
