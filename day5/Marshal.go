package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// 요청 구조체
type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// 응답 구조체
type CreateUserResponse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func createUser(w http.ResponseWriter, r *http.Request) {
	// 안전장치: POST 메서드만 허용합니다.
	if r.Method != http.MethodPost {
		http.Error(w, "허용되지 않은 메서드입니다. (Only POST allowed)", http.StatusMethodNotAllowed)
		return
	}

	var req CreateUserRequest

	// 요청 본문(r.Body) 디코딩
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "잘못된 요청 JSON 형식입니다.", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// 유효성 검사
	if req.Name == "" {
		http.Error(w, "name 필드는 필수입니다.", http.StatusBadRequest)
		return
	}

	// 비즈니스 로직 (DB 저장 등 생략)
	resp := CreateUserResponse{
		ID:        42, // Mock ID
		Name:      req.Name,
		Email:     req.Email,
		CreatedAt: time.Now(),
	}

	// 응답 인코딩하여 클라이언트에게 전송
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created 응답 코드 설정
	json.NewEncoder(w).Encode(resp)
}

func main() {
	mux := http.NewServeMux()

	// 라우터에 유저 생성 핸들러 매핑
	mux.HandleFunc("/api/users", createUser)

	fmt.Println("API 서버가 :8080 포트에서 시작되었습니다...")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal("서버 구동 실패:", err)
	}
}
/*
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

// Item 구조체 정의 (테스트용)
type Item struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// 각 아이템을 처리할 함수
func process(item Item) {
	fmt.Printf("[처리 중] ID: %d, Name: %s\n", item.ID, item.Name)
}

func main() {
	// 1. 대용량 JSON 파일 열기
	file, err := os.Open("huge.json")
	if err != nil {
		log.Fatalf("파일을 열 수 없습니다: %v\n(팁: 테스트용 huge.json 파일을 먼저 생성해주세요!)", err)
	}
	defer file.Close()

	// 2. 스트리밍 디코더 생성
	decoder := json.NewDecoder(file)

	// 3. 배열의 시작 토큰인 '['를 읽어서 지나갑니다.
	t, err := decoder.Token()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("시작 토큰 확인: %v\n", t) // 출력: [

	// 4. 배열의 끝(']')을 만날 때까지 반복해서 객체를 하나씩 읽습니다.
	count := 0
	for decoder.More() {
		var item Item
		// Decode는 현재 스트림 위치에서 JSON 객체 하나만 파싱하고 포인터를 다음으로 이동시킵니다.
		if err := decoder.Decode(&item); err != nil {
			log.Fatal("디코딩 실패:", err)
		}

		// 파싱된 아이템 처리
		process(item)
		count++
	}

	// 5. 배열의 끝 토큰인 ']'를 읽어서 마무리합니다.
	t, err = decoder.Token()
	if err != nil && err != io.EOF {
		log.Fatal(err)
	}
	fmt.Printf("끝 토큰 확인: %v\n", t) // 출력: ]
	fmt.Printf("총 %d개의 아이템을 메모리 낭비 없이 성공적으로 처리했습니다!\n", count)
}
*/

/*
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
)

// Money는 cents 단위를 나타내는 커스텀 타입입니다. (예: 1234 = $12.34)
type Money int64

// Product는 Money 타입을 사용하는 예시 구조체입니다.
type Product struct {
	Name  string `json:"name"`
	Price Money  `json:"price"`
}

// MarshalJSON: Go 구조체 ➡️ JSON 문자열
// 내부 정수형 cents 데이터를 소수점이 있는 문자열 형태로 변환합니다. (예: 1234 ➡️ "12.34")
func (m Money) MarshalJSON() ([]byte, error) {
	dollars := float64(m) / 100
	return []byte(fmt.Sprintf(`"%.2f"`, dollars)), nil
}

// UnmarshalJSON: JSON 문자열 ➡️ Go 구조체
// 소수점 문자열을 읽어와 다시 정수형 cents 데이터로 변환합니다. (예: "12.34" ➡️ 1234)
func (m *Money) UnmarshalJSON(data []byte) error {
	// 앞뒤 큰따옴표(") 제거
	s := strings.Trim(string(data), `"`)

	// 문자열을 float64로 파싱
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}

	// ⚠️ 중요 부동소수점 오차 방지: float에 100을 곱할 때 미세한 오차(예: 19.98999...)가
	// 발생할 수 있으므로 반올림(math.Round)을 해준 뒤 정수로 캐스팅해야 안전합니다.
	*m = Money(math.Round(f * 100))
	return nil
}

func main() {
	// ==========================================
	// 1. 테스트: Go 구조체를 JSON으로 변환 (Marshal)
	// ==========================================
	p1 := Product{
		Name:  "Premium Coffee Beans",
		Price: 1234, // $12.34를 뜻함
	}

	jsonData, err := json.Marshal(p1)
	if err != nil {
		log.Fatal("소형 변환 실패:", err)
	}

	fmt.Println("--- 1. 구조체 ➡️ JSON 결과 ---")
	fmt.Println(string(jsonData))
	// 출력: {"name":"Premium Coffee Beans","price":"12.34"}

	fmt.Println() // 줄바꿈

	// ==========================================
	// 2. 테스트: JSON을 Go 구조체로 변환 (Unmarshal)
	// ==========================================
	jsonStr := `{"name":"Mechanical Keyboard","price":"99.95"}`

	var p2 Product
	err = json.Unmarshal([]byte(jsonStr), &p2)
	if err != nil {
		log.Fatal("구조체 변환 실패:", err)
	}

	fmt.Println("--- 2. JSON ➡️ 구조체 결과 ---")
	fmt.Printf("상품명: %s\n", p2.Name)
	fmt.Printf("정수형 Cents 값: %d (실제 금액: $%.2f)\n", p2.Price, float64(p2.Price)/100)
	// 출력:
	// 상품명: Mechanical Keyboard
	// 정수형 Cents 값: 9995 (실제 금액: $99.95)
}
*/
/*
package main

import (
	"encoding/json"
	"fmt"
	"reflect"
)

func main() {
	// JSON 문자열 준비
	jsonStr := `{"name":"Alice","age":30,"tags":["go","c"]}`

	// 결과를 담을 map 선언 (Go 1.18 미만 버전이라면 any 대신 interface{} 사용)
	var data map[string]any

	// JSON ➡️ Map 변환
	err := json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		fmt.Println("파싱 에러:", err)
		return
	}

	fmt.Println("--- 1. 데이터 출력 ---")
	fmt.Println("name:", data["name"]) // Alice
	fmt.Println("age:", data["age"])   // 30
	fmt.Println("tags:", data["tags"]) // [go c]

	fmt.Println("\n--- 2. 실제 내부 타입 확인 ---")
	// reflect.TypeOf()를 사용해 Go가 실제로 어떤 타입으로 저장했는지 확인합니다.
	fmt.Println("name 타입:", reflect.TypeOf(data["name"])) // string
	fmt.Println("age 타입: ", reflect.TypeOf(data["age"]))  // float64 (주의!)
	fmt.Println("tags 타입:", reflect.TypeOf(data["tags"])) // []interface{} 또는 []any

	fmt.Println("\n--- 3. 안전하게 데이터 꺼내 쓰기 (Type Assertion) ---")
	// age를 안전하게 int로 바꾸어 연산에 사용하려면 아래와 같이 타입 인터셉트가 필요합니다.
	if ageFloat, ok := data["age"].(float64); ok {
		ageInt := int(ageFloat) // float64를 int로 변환
		fmt.Printf("내년 Alice의 나이는 %d살입니다.\n", ageInt+1)
	} else {
		fmt.Println("age가 숫자가 아닙니다.")
	}
}
*/

/*
package main

import (
	"encoding/json"
	"fmt"
)

// User 구조체 정의
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age,omitempty"` // 값이 0(기본값)이면 JSON 변환 시 제외됨
}

func main() {
	// 1. Go 구조체 ➡️ JSON 문자열 (직렬화 / Marshal)
	u := User{ID: 1, Name: "Alice", Age: 30}
	data, _ := json.Marshal(u)

	fmt.Println("--- 1. Go 구조체를 JSON으로 변환 ---")
	fmt.Println(string(data))
	// 출력: {"id":1,"name":"Alice","age":30}

	fmt.Println() // 줄바꿈

	// 2. JSON 문자열 ➡️ Go 구조체 (역직렬화 / Unmarshal)
	jsonStr := `{"id":2,"name":"Bob"}`
	var u2 User
	_ = json.Unmarshal([]byte(jsonStr), &u2)

	fmt.Println("--- 2. JSON을 Go 구조체로 변환 ---")
	fmt.Printf("%+v\n", u2)
	// 출력: {ID:2 Name:Bob Age:0}
}
*/
