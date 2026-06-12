// 1.2 보충 — http.Client로 타임아웃·헤더 제어 (전체 코드)
// 실행 전 다른 터미널에서 1.3의 서버(server.go)를 띄워두세요.
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func fetchWithClient(url, token string) error {
	client := &http.Client{
		Timeout: 5 * time.Second, // 전체 요청 타임아웃
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("요청 생성 실패: %w", err)
	}
	req.Header.Set("User-Agent", "MyApp/1.0")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("요청 실패: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("본문 읽기 실패: %w", err)
	}

	fmt.Println("상태:", resp.StatusCode)
	fmt.Println("본문:", string(body))
	return nil
}

func main() {
	url := "http://localhost:8080/hello?name=Client"
	if len(os.Args) > 1 {
		url = os.Args[1]
	}
	if err := fetchWithClient(url, "demo-token"); err != nil {
		fmt.Println("에러:", err)
		os.Exit(1)
	}
}
