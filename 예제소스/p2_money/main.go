package main

import (
	"encoding/json"
	"fmt"
)

type Product struct {
	Name  string `json:"name"`
	Price Money  `json:"price"`
}

func main() {
	// 직렬화: 1234 cents → "12.34"
	p := Product{Name: "키보드", Price: 1234}
	data, _ := json.Marshal(p)
	fmt.Println(string(data))

	// 역직렬화: "99.99" → 9999 cents
	var p2 Product
	json.Unmarshal([]byte(`{"name":"마우스","price":"99.99"}`), &p2)
	fmt.Printf("%+v (cents=%d)\n", p2, int64(p2.Price))
}
