// 2.7 — CSV / XML / Base64 / Gob (전체 코드)
package main

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"encoding/gob"
	"encoding/xml"
	"fmt"
	"os"
)

type Person struct {
	XMLName xml.Name `xml:"person"`
	Name    string   `xml:"name"`
	Age     int      `xml:"age"`
}

func main() {
	// --- CSV ---
	fmt.Println("--- CSV ---")
	w := csv.NewWriter(os.Stdout)
	w.Write([]string{"name", "age"})
	w.Write([]string{"Alice", "30"})
	w.Flush()
	if err := w.Error(); err != nil { // Flush 후 에러 확인 습관
		panic(err)
	}

	// --- XML ---
	fmt.Println("--- XML ---")
	p := Person{Name: "Alice", Age: 30}
	out, _ := xml.MarshalIndent(p, "", "  ")
	fmt.Println(string(out))

	// XML → Go
	var p2 Person
	xml.Unmarshal(out, &p2)
	fmt.Printf("역직렬화: %+v\n", p2)

	// --- Base64 ---
	fmt.Println("--- Base64 ---")
	enc := base64.StdEncoding.EncodeToString([]byte("Hello"))
	fmt.Println("인코딩:", enc)
	dec, _ := base64.StdEncoding.DecodeString(enc)
	fmt.Println("디코딩:", string(dec))

	// --- Gob (Go 전용 바이너리) ---
	fmt.Println("--- Gob ---")
	var buf bytes.Buffer
	type Point struct{ X, Y int }
	gob.NewEncoder(&buf).Encode(Point{3, 4})
	fmt.Println("gob 크기:", buf.Len(), "바이트")
	var pt Point
	gob.NewDecoder(&buf).Decode(&pt)
	fmt.Printf("복원: %+v\n", pt)
}
