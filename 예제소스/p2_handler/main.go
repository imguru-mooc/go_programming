// 2.6 — HTTP 핸들러의 표준 JSON 패턴 (전체 코드)
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type CreateUserResponse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "잘못된 요청", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.Name == "" {
		http.Error(w, "name 필수", http.StatusBadRequest)
		return
	}

	resp := CreateUserResponse{
		ID:        42,
		Name:      req.Name,
		Email:     req.Email,
		CreatedAt: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /users", createUser)
	fmt.Println("서버 시작: http://localhost:8080")
	http.ListenAndServe(":8080", mux)
}
