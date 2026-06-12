// 2.2 — 구조체 태그로 직렬화 제어 (전체 코드)
package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email,omitempty"` // 비어있으면 생략
	Password  string    `json:"-"`               // 항상 제외
	CreatedAt time.Time `json:"created_at"`
	BigNum    int64     `json:"big_num,string"`  // 숫자를 문자열로
}

// 소문자 필드는 직렬화되지 않음을 보여주는 타입
type Bad struct {
	name string `json:"name"` // ❌ unexported — json 패키지가 못 봄
	Age  int    `json:"age"`  // ✅
}

func main() {
	u := User{
		ID:        1,
		Name:      "Alice",
		Email:     "", // omitempty로 생략됨
		Password:  "super-secret",
		CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		BigNum:    9007199254740993, // JS Number로 못 담는 큰 정수
	}
	data, _ := json.MarshalIndent(u, "", "  ")
	fmt.Println(string(data))
	// password 필드가 없고, email이 생략되고, big_num이 "문자열"임을 확인

	b := Bad{name: "보이지않음", Age: 10}
	_ = b.name // unused 경고 회피
	data2, _ := json.Marshal(b)
	fmt.Println(string(data2)) // {"age":10} — name 없음
}
