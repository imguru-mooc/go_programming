// 2.3 — 동적 JSON과 float64 함정 (전체 코드)
package main

import (
	"encoding/json"
	"fmt"
)

func main() {
	var data map[string]any
	raw := `{"name":"Alice","age":30,"tags":["go","c"]}`
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		panic(err)
	}

	fmt.Println(data["name"]) // Alice
	fmt.Println(data["age"])  // 30
	fmt.Println(data["tags"]) // [go c]

	// 실제 타입 확인
	fmt.Printf("age의 실제 타입: %T\n", data["age"]) // float64!

	// ✅ 올바른 단언
	n := data["age"].(float64)
	fmt.Println("정수 변환:", int(n))

	// ❌ 잘못된 단언은 panic — comma-ok로 안전하게 확인
	if _, ok := data["age"].(int); !ok {
		fmt.Println("data[\"age\"].(int)는 실패한다 (panic 방지: comma-ok 사용)")
	}

	// 중첩 접근: tags는 []any
	tags := data["tags"].([]any)
	for i, t := range tags {
		fmt.Printf("tags[%d] = %s (%T)\n", i, t, t)
	}
}
