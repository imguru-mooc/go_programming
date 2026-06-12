// 1.4 — Go 1.22+ 라우팅 문법 (전체 코드)
package main

import (
	"fmt"
	"net/http"
)

func listUsers(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "사용자 목록")
}

func createUser(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "사용자 생성됨")
}

func getUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id") // path parameter 추출
	fmt.Fprintf(w, "User ID: %s\n", id)
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "User %s 수정됨\n", r.PathValue("id"))
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /users", listUsers)
	mux.HandleFunc("POST /users", createUser)
	mux.HandleFunc("GET /users/{id}", getUser)
	mux.HandleFunc("PUT /users/{id}", updateUser)
	mux.HandleFunc("DELETE /users/{id}", deleteUser)

	fmt.Println("서버 시작: http://localhost:8080")
	http.ListenAndServe(":8080", mux)
}
