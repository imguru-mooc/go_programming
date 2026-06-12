// 3.6 — 인터페이스 기반 mock (전체 코드)
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// 표준 *http.Client도 이 인터페이스를 만족 (Do 메서드 보유)
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type APIService struct {
	baseURL string
	client  HTTPClient
}

func NewAPIService(baseURL string, client HTTPClient) *APIService {
	return &APIService{baseURL: baseURL, client: client}
}

func (s *APIService) GetUser(id int) (*User, error) {
	url := fmt.Sprintf("%s/users/%d", s.baseURL, id)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("요청 실패: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	var u User
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return nil, fmt.Errorf("디코딩 실패: %w", err)
	}
	return &u, nil
}

func main() {
	// 실제 사용 시: 진짜 http.Client 주입
	svc := NewAPIService("https://api.example.com", &http.Client{})
	_ = svc
	fmt.Println("이 프로그램의 핵심은 service_test.go — go test -v 로 확인하세요")
}
