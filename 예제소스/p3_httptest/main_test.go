// 3.7 — httptest.NewServer로 외부 API 모킹 (전체 코드)
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// 테스트 대상: 외부 API의 상태를 조회하는 함수
func fetchStatus(baseURL string) (string, error) {
	resp, err := http.Get(baseURL + "/status")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", err
	}
	return body.Status, nil
}

func TestFetchStatus(t *testing.T) {
	// 가짜 외부 API 서버 — 임의 포트에서 실제로 기동
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 요청 검증도 가능
		if r.URL.Path != "/status" {
			t.Errorf("예상치 못한 경로: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":"ok"}`)
	}))
	defer server.Close()

	got, err := fetchStatus(server.URL) // server.URL = http://127.0.0.1:임의포트
	if err != nil {
		t.Fatal(err)
	}
	if got != "ok" {
		t.Errorf("status = %q, want ok", got)
	}
}
