// 2.5 — json.Decoder로 스트리밍 디코딩 (전체 코드)
// 큰 JSON 배열 파일을 직접 만들어서, 전체를 메모리에 올리지 않고
// 항목 단위로 처리하는 것을 시연합니다.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type Item struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
}

// 시연용 huge.json 생성 (실무에선 이미 존재하는 대용량 파일)
func generateFile(path string, n int) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	f.WriteString("[\n")
	enc := json.NewEncoder(f)
	for i := 1; i <= n; i++ {
		if i > 1 {
			f.WriteString(",")
		}
		enc.Encode(Item{ID: i, Name: fmt.Sprintf("item-%d", i), Price: i * 100})
	}
	f.WriteString("]\n")
	return nil
}

func process(item Item) {
	if item.ID%2500 == 0 { // 너무 많이 찍지 않도록 샘플만 출력
		fmt.Printf("처리 중: %+v\n", item)
	}
}

func main() {
	const path = "huge.json"
	const n = 10000
	if err := generateFile(path, n); err != nil {
		log.Fatal(err)
	}
	info, _ := os.Stat(path)
	fmt.Printf("생성된 파일: %s (%d KB, 항목 %d개)\n", path, info.Size()/1024, n)

	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	// 배열 시작 토큰 '[' 소비
	if _, err := decoder.Token(); err != nil {
		log.Fatal(err)
	}

	count := 0
	for decoder.More() {
		var item Item
		if err := decoder.Decode(&item); err != nil {
			log.Fatal(err)
		}
		process(item)
		count++
	}

	// 배열 종료 토큰 ']' 소비
	if _, err := decoder.Token(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("총 처리 항목:", count)
}
