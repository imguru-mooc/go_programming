package main

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

// mock 구현 — 인터페이스만 만족하면 됨
type mockClient struct {
	response *http.Response
	err      error
}

func (m *mockClient) Do(req *http.Request) (*http.Response, error) {
	return m.response, m.err
}

func TestGetUser_OK(t *testing.T) {
	body := io.NopCloser(strings.NewReader(`{"id":1,"name":"Alice"}`))
	mock := &mockClient{
		response: &http.Response{StatusCode: 200, Body: body},
	}

	svc := NewAPIService("http://fake", mock)
	user, err := svc.GetUser(1)

	if err != nil {
		t.Fatal(err)
	}
	if user.Name != "Alice" {
		t.Errorf("got %s, want Alice", user.Name)
	}
}

func TestGetUser_ServerError(t *testing.T) {
	body := io.NopCloser(strings.NewReader(``))
	mock := &mockClient{
		response: &http.Response{StatusCode: 500, Body: body},
	}
	svc := NewAPIService("http://fake", mock)
	if _, err := svc.GetUser(1); err == nil {
		t.Error("500인데 에러 없음")
	}
}

func TestGetUser_NetworkError(t *testing.T) {
	mock := &mockClient{err: errors.New("connection refused")}
	svc := NewAPIService("http://fake", mock)
	if _, err := svc.GetUser(1); err == nil {
		t.Error("네트워크 에러인데 에러 없음")
	}
}
