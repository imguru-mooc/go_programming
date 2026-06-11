package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"time"
)

func fetch(url string, timeout time.Duration) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err // 타임아웃이면 ctx.Err() = DeadlineExceeded
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func main() {
	// 테스트용 HTTP 서버 생성
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		switch path {
		case "/fast":
			fmt.Fprintln(w, "빠른 응답입니다")

		case "/slow":
			time.Sleep(3 * time.Second)
			fmt.Fprintln(w, "느린 응답입니다")

		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	fmt.Println("테스트 서버 주소:", server.URL)

	// 1. 정상 요청 테스트
	fmt.Println("\n=== 빠른 요청 테스트 ===")

	body, err := fetch(server.URL+"/fast", 2*time.Second)
	if err != nil {
		fmt.Println("에러:", err)
	} else {
		fmt.Println("응답:", string(body))
	}

	// 2. 타임아웃 요청 테스트
	fmt.Println("\n=== 느린 요청 타임아웃 테스트 ===")

	body, err = fetch(server.URL+"/slow", 1*time.Second)
	if err != nil {
		fmt.Println("에러:", err)

		if context.DeadlineExceeded == context.Cause(context.Background()) {
			fmt.Println("이 코드는 의미 없음")
		}

		if err == context.DeadlineExceeded {
			fmt.Println("직접 비교: context deadline exceeded")
		} else {
			fmt.Println("http 요청 에러:", err)
		}
	} else {
		fmt.Println("응답:", string(body))
	}
}
