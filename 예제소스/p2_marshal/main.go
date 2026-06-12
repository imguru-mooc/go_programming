// 2.1 — Marshal / Unmarshal (전체 코드)
package main

import (
	"encoding/json"
	"fmt"
	"log"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age,omitempty"`
}

func main() {
	// Go → JSON
	u := User{ID: 1, Name: "Alice", Age: 30}
	data, err := json.Marshal(u)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(data))

	// JSON → Go
	jsonStr := `{"id":2,"name":"Bob"}`
	var u2 User
	if err := json.Unmarshal([]byte(jsonStr), &u2); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", u2)

	// omitempty 동작 확인: Age가 zero value(0)면 출력에서 생략
	data2, _ := json.Marshal(u2)
	fmt.Println(string(data2))
}
