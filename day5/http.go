package main

import (
    "fmt"
    "io"
    "net/http"
)

func main() {
    resp, err := http.Get("https://github.com")
    if err != nil {
        fmt.Println("에러:", err)
        return
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        fmt.Println("읽기 실패:", err)
        return
    }

    fmt.Println("상태:", resp.StatusCode)
    fmt.Println("길이:", len(body), "바이트")
    fmt.Println(string(body))
}
