package main

import (
    "fmt"
    "net/http"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, %s!", r.URL.Query().Get("name"))
}

func main() {
    http.HandleFunc("/hello", helloHandler)
    fmt.Println("서버 시작: http://localhost:8080")
    http.ListenAndServe(":8080", nil)
}
