package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// Middleware는 http.Handler를 받아 다른 http.Handler를 반환하는 함수 타입입니다.
type Middleware func(http.Handler) http.Handler

// 1. 로깅 미들웨어: 요청의 메서드, URL, 처리 시간을 기록합니다.
func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// 다음 핸들러(또는 미들웨어) 호출
		next.ServeHTTP(w, r)
		
		// 처리가 끝난 후 로그 출력
		log.Printf("[%s] %s - %v", r.Method, r.URL.Path, time.Since(start))
	})
}

// 2. 인증 미들웨어: 헤더에 Authorization 값이 있는지 검증합니다.
func auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			// 토큰이 없으면 401 에러를 반환하고 체인을 중단합니다.
			http.Error(w, "인증되지 않은 요청입니다. (Unauthorized)", http.StatusUnauthorized)
			return
		}
		
		// 토큰이 있으면 다음 핸들러로 진행
		next.ServeHTTP(w, r)
	})
}

// 3. 실제 비즈니스 로직을 수행할 메인 핸들러
func secretHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "옜다 비밀 정보! 🤫 (인증 성공)")
}

func main() {
	mux := http.NewServeMux()
	
	// 원래 핸들러 등록
	mux.HandleFunc("/api/secret", secretHandler)

	// 미들웨어 체이닝 (양파 껍질처럼 감싸집니다)
	// 1. 요청이 들어오면 logging이 먼저 받음
	// 2. logging이 auth에게 넘김
	// 3. auth가 통과되면 최종적으로 mux(/api/secret)가 실행됨
	handler := logging(auth(mux))

	fmt.Println("서버가 :8080 포트에서 시작되었습니다...")
	err := http.ListenAndServe(":8080", handler)
	if err != nil {
		log.Fatal("서버 시작 실패: ", err)
	}
}
