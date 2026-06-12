package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateUser_OK(t *testing.T) {
	body := strings.NewReader(`{"name":"Alice","email":"alice@example.com"}`)
	req := httptest.NewRequest("POST", "/users", body)
	w := httptest.NewRecorder()

	createUser(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", w.Code)
	}
	var resp CreateUserResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Name != "Alice" || resp.ID != 42 {
		t.Errorf("resp = %+v", resp)
	}
}

func TestCreateUser_BadJSON(t *testing.T) {
	req := httptest.NewRequest("POST", "/users", strings.NewReader(`{잘못된}`))
	w := httptest.NewRecorder()
	createUser(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestCreateUser_MissingName(t *testing.T) {
	req := httptest.NewRequest("POST", "/users", strings.NewReader(`{"email":"x@x.com"}`))
	w := httptest.NewRecorder()
	createUser(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}
