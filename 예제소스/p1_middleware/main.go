// 1.5 — 미들웨어 패턴 (전체 코드)
package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

type Middleware func(http.Handler) http.Handler

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}

func auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// 체인 헬퍼: chain(h, logging, auth) → logging(auth(h))
func chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

func secretHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "🔐 비밀 데이터")
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/secret", secretHandler)

	// 미들웨어 체이닝: 요청 → logging → auth → mux
	handler := chain(mux, logging, auth)

	fmt.Println("서버 시작: http://localhost:8080")
	http.ListenAndServe(":8080", handler)
}
