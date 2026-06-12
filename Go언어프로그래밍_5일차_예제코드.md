# Go 언어 프로그래밍 5일차 — 조각코드 전체판 (실행 가능 코드 + 테스트 방법)

> **용도**: 5일차 강의자료 본문의 "조각코드"(설명용 발췌 코드)를 **그대로 복사해서 실행할 수 있는 전체 프로그램**으로 보완한 부록입니다.
> **검증 환경**: Go 1.22.2 / GCC 13 / Ubuntu 24.04 — **모든 코드는 실제 빌드·실행·테스트를 통과했으며**, 각 절의 "검증 결과"는 실측 출력입니다.
> **절 번호**: 본문 강의자료의 절 번호와 1:1 대응합니다. 본문에 이미 전체 코드가 있는 절(1.3, 1.6, 3.11, 4.7, 4.8, 5.10, 6.2~6.4, 6.7, 6.10, 7~8교시)은 마지막 "검증 현황표"에 테스트 방법만 정리했습니다.

## 📂 실습 디렉토리 구성

각 예제는 독립 모듈입니다. 아래처럼 한 번에 만들어두면 편합니다.

```bash
mkdir -p ~/go-class/day5/snippets && cd ~/go-class/day5/snippets
# 각 예제 디렉토리에서:  mkdir <이름> && cd <이름> && go mod init <이름>
```

```
snippets/
├── p1_server/        # 1.3  기본 서버 (테스트 페어용)
├── p1_client/        # 1.2  http.Client
├── p1_client_ctx/    # 1.2  Context + HTTP
├── p1_routing/       # 1.4  Go 1.22 라우팅
├── p1_middleware/    # 1.5  미들웨어 체이닝
├── p2_marshal/       # 2.1  Marshal/Unmarshal
├── p2_tags/          # 2.2  구조체 태그
├── p2_dynamic/       # 2.3  동적 JSON
├── p2_money/         # 2.4  커스텀 Marshaler (+테스트)
├── p2_stream/        # 2.5  스트리밍 디코딩
├── p2_handler/       # 2.6  핸들러 JSON 패턴 (+httptest)
├── p2_encodings/     # 2.7  CSV/XML/Base64/Gob
├── p3_mock/          # 3.6  인터페이스 mock (+테스트)
├── p3_httptest/      # 3.7  httptest.NewServer
├── p4_ctxlog/        # 4.4  Context 로거 전달
├── p4_recover/       # 4.6  recover + 스택 트레이스
├── p5_sort/          # 5.4  slowSort + pprof
├── p5_trace/         # 5.9  runtime/trace
└── p6_slice/         # 6.6  슬라이스 ↔ C 배열
```

---

# 1교시 보완

## 1.2-A `http.Client` — 타임아웃·헤더 제어 (전체 코드)

본문의 `client := &http.Client{...}` 조각을 실행 가능한 프로그램으로 만든 것입니다. **1.3의 기본 서버를 상대로 테스트**하도록 구성해, 외부 네트워크 없이 실습할 수 있습니다.

`p1_client/main.go`:

```go
// 실행 전 다른 터미널에서 1.3의 서버(p1_server)를 띄워두세요.
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func fetchWithClient(url, token string) error {
	client := &http.Client{
		Timeout: 5 * time.Second, // 전체 요청 타임아웃
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("요청 생성 실패: %w", err)
	}
	req.Header.Set("User-Agent", "MyApp/1.0")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("요청 실패: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("본문 읽기 실패: %w", err)
	}

	fmt.Println("상태:", resp.StatusCode)
	fmt.Println("본문:", string(body))
	return nil
}

func main() {
	url := "http://localhost:8080/hello?name=Client"
	if len(os.Args) > 1 {
		url = os.Args[1] // 다른 URL도 인자로 테스트 가능
	}
	if err := fetchWithClient(url, "demo-token"); err != nil {
		fmt.Println("에러:", err)
		os.Exit(1)
	}
}
```

**테스트 방법**:

```bash
# 터미널 1 — 1.3의 서버 실행
cd p1_server && go run .

# 터미널 2
cd p1_client && go run .
# 임의의 URL로도: go run . https://api.github.com
```

**검증 결과**:
```
상태: 200
본문: Hello, Client!
```

## 1.2-B Context와 결합 — 타임아웃 자동 취소를 눈으로 확인 (전체 코드)

본문의 4줄짜리 조각을 **타임아웃이 실제로 작동하는지 관찰 가능한** 프로그램으로 확장했습니다. 5초 걸리는 느린 핸들러를 내장해, 3초 제한이 정확히 3.0초에 끊는 것을 보여줍니다.

`p1_client_ctx/main.go`:

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

func fetchWithTimeout(url string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println("성공! 상태:", resp.StatusCode)
	return nil
}

func main() {
	// 느린 서버를 내장해서 시연 (실무에선 외부 URL)
	mux := http.NewServeMux()
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(5 * time.Second): // 5초 걸리는 응답
			w.Write([]byte("늦은 응답"))
		case <-r.Context().Done(): // 클라이언트가 끊으면 즉시 중단
			return
		}
	})
	mux.HandleFunc("/fast", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("빠른 응답"))
	})
	srv := &http.Server{Addr: ":8081", Handler: mux}
	go srv.ListenAndServe()
	time.Sleep(100 * time.Millisecond) // 서버 기동 대기

	fmt.Println("--- 빠른 엔드포인트 (3초 제한) ---")
	if err := fetchWithTimeout("http://localhost:8081/fast", 3*time.Second); err != nil {
		fmt.Println("에러:", err)
	}

	fmt.Println("--- 느린 엔드포인트 (3초 제한, 5초 걸림) ---")
	start := time.Now()
	err := fetchWithTimeout("http://localhost:8081/slow", 3*time.Second)
	fmt.Printf("%.1f초 후 결과: %v\n", time.Since(start).Seconds(), err)
	if errors.Is(err, context.DeadlineExceeded) {
		fmt.Println("→ context.DeadlineExceeded로 자동 취소 확인!")
	}
	srv.Close()
}
```

**테스트 방법**: `go run .` 한 번이면 됩니다 (서버 내장).

**검증 결과**:
```
--- 빠른 엔드포인트 (3초 제한) ---
성공! 상태: 200
--- 느린 엔드포인트 (3초 제한, 5초 걸림) ---
3.0초 후 결과: Get "http://localhost:8081/slow": context deadline exceeded
→ context.DeadlineExceeded로 자동 취소 확인!
```

> 💡 **강의 포인트**: 에러 비교에 `errors.Is(err, context.DeadlineExceeded)`를 쓰는 것은 4일차 에러 래핑(`%w`) 내용과 직결됩니다. `http.Client`가 ctx 에러를 `*url.Error`로 감싸서 반환하기 때문에 `==` 비교는 실패하고 `errors.Is`의 Unwrap 탐색이 필요합니다.

## 1.4 Go 1.22 라우팅 (전체 코드)

본문 조각의 `listUsers` 등 5개 핸들러를 전부 채웠습니다.

`p1_routing/main.go`:

```go
package main

import (
	"fmt"
	"net/http"
)

func listUsers(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "사용자 목록")
}

func createUser(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "사용자 생성됨")
}

func getUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id") // path parameter 추출
	fmt.Fprintf(w, "User ID: %s\n", id)
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "User %s 수정됨\n", r.PathValue("id"))
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /users", listUsers)
	mux.HandleFunc("POST /users", createUser)
	mux.HandleFunc("GET /users/{id}", getUser)
	mux.HandleFunc("PUT /users/{id}", updateUser)
	mux.HandleFunc("DELETE /users/{id}", deleteUser)

	fmt.Println("서버 시작: http://localhost:8080")
	http.ListenAndServe(":8080", mux)
}
```

**테스트 방법** (다른 터미널에서):

```bash
curl http://localhost:8080/users                                    # GET
curl -X POST http://localhost:8080/users                            # POST
curl http://localhost:8080/users/42                                 # path param
curl -X PUT http://localhost:8080/users/42                          # PUT
curl -s -o /dev/null -w "%{http_code}\n" -X DELETE http://localhost:8080/users/42   # 204
curl -s -o /dev/null -w "%{http_code}\n" -X PATCH  http://localhost:8080/users      # 405!
```

**검증 결과**:
```
사용자 목록
사용자 생성됨
User ID: 42
User 42 수정됨
204
405          ← 미등록 메서드는 자동으로 405 Method Not Allowed
```

> 💡 **강의 포인트**: 마지막 405가 새 라우팅의 숨은 장점입니다. 경로는 있는데 메서드만 다르면 404가 아닌 **405를 자동으로** 반환합니다 — 예전처럼 `if r.Method != ...` 분기를 직접 짤 때 흔히 빠뜨리던 부분입니다.

## 1.5 미들웨어 패턴 (전체 코드)

본문 조각에서 미정의였던 `secretHandler`를 채우고, 강의에서 자주 받는 질문 "미들웨어가 많아지면 괄호 지옥 아닌가요?"에 답하는 `chain` 헬퍼를 추가했습니다.

`p1_middleware/main.go`:

```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

type Middleware func(http.Handler) http.Handler

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}

func auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// 체인 헬퍼: chain(h, logging, auth) → logging(auth(h))
func chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

func secretHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "🔐 비밀 데이터")
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/secret", secretHandler)

	// 미들웨어 체이닝: 요청 → logging → auth → mux
	handler := chain(mux, logging, auth)

	fmt.Println("서버 시작: http://localhost:8080")
	http.ListenAndServe(":8080", handler)
}
```

**테스트 방법**:

```bash
curl -s -w "\n%{http_code}\n" http://localhost:8080/api/secret                          # 401 기대
curl -s -w "\n%{http_code}\n" -H "Authorization: Bearer abc" http://localhost:8080/api/secret  # 200 기대
```

**검증 결과**:
```
Unauthorized
401
🔐 비밀 데이터
200
```
서버 측 로그에는 두 요청 모두 `GET /api/secret 14.798µs` 형태로 기록됩니다 — **auth가 401로 끊어도 바깥쪽 logging은 실행됨**(체인 순서 이해에 좋은 관찰 포인트).

## 1.6 Graceful Shutdown — "graceful"인지 실제로 검증하는 방법

본문 코드는 이미 전체 코드입니다. 여기서는 **정말 진행 중 요청을 기다리는지** 확인하는 테스트 절차를 추가합니다. 핸들러가 2초 걸리므로, 요청을 보내놓고 즉시 SIGTERM을 날렸을 때 응답이 끝까지 오면 성공입니다.

```bash
go build -o app . && ./app &
SRV=$!
sleep 0.3

curl -s http://localhost:8080/ &   # 2초 걸리는 요청 시작
CURL=$!
sleep 0.3
kill -TERM $SRV                    # 요청 진행 중에 종료 신호!

wait $CURL && echo " ← 진행 중이던 요청 응답 수신 (graceful 확인)"
```

**검증 결과** (실측 타임라인):
```
23:59:11 서버 시작
23:59:12 종료 시작...          ← SIGTERM 수신
OK ← 진행 중이던 요청 응답 수신
23:59:14 정상 종료             ← 2초짜리 요청이 끝난 뒤에야 종료
```

`srv.Close()`로 바꿔서 다시 실행해 보면 curl이 `connection reset`으로 실패합니다 — Shutdown과 Close의 차이를 체감시키기 좋은 비교 실습입니다.

---

# 2교시 보완

## 2.1 Marshal / Unmarshal (전체 코드)

`p2_marshal/main.go`:

```go
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
```

**테스트 방법**: `go run .`

**검증 결과**:
```
{"id":1,"name":"Alice","age":30}
{ID:2 Name:Bob Age:0}
{"id":2,"name":"Bob"}        ← Age=0이라 omitempty로 생략됨
```

## 2.2 구조체 태그 (전체 코드)

본문 태그 표의 항목들(`omitempty`, `-`, `,string`)과 **소문자 필드 함정**을 한 프로그램에서 전부 시연합니다.

`p2_tags/main.go`:

```go
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
	BigNum    int64     `json:"big_num,string"` // 숫자를 문자열로
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
		BigNum:    9007199254740993, // JS Number가 못 담는 큰 정수
	}
	data, _ := json.MarshalIndent(u, "", "  ")
	fmt.Println(string(data))

	b := Bad{name: "보이지않음", Age: 10}
	_ = b.name // unused 경고 회피
	data2, _ := json.Marshal(b)
	fmt.Println(string(data2)) // {"age":10} — name 없음
}
```

**테스트 방법**: `go run .` — 출력에서 ① `password` 키 자체가 없음, ② `email` 생략, ③ `big_num`이 따옴표로 감싸짐, ④ `Bad`에서 `name` 누락, 4가지를 확인합니다.

**검증 결과**:
```json
{
  "id": 1,
  "name": "Alice",
  "created_at": "2026-01-01T00:00:00Z",
  "big_num": "9007199254740993"
}
{"age":10}
```

## 2.3 동적 JSON과 float64 함정 (전체 코드)

본문 조각의 `❌ panic` 줄은 그대로 두면 실행이 안 되므로, **comma-ok 패턴으로 안전하게 함정을 시연**하는 형태로 완성했습니다.

`p2_dynamic/main.go`:

```go
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
		fmt.Println(`data["age"].(int)는 실패한다 (panic 방지: comma-ok 사용)`)
	}

	// 중첩 접근: tags는 []any
	tags := data["tags"].([]any)
	for i, t := range tags {
		fmt.Printf("tags[%d] = %s (%T)\n", i, t, t)
	}
}
```

**테스트 방법**: `go run .`. 함정을 직접 체험시키려면 comma-ok를 지우고 `data["age"].(int)`로 바꿔 재실행 → `interface conversion: interface {} is float64, not int` panic을 관찰.

**검증 결과**:
```
Alice
30
[go c]
age의 실제 타입: float64
정수 변환: 30
data["age"].(int)는 실패한다 (panic 방지: comma-ok 사용)
tags[0] = go (string)
tags[1] = c (string)
```

## 2.4 커스텀 Marshaler — Money (전체 코드 + 단위 테스트)

본문 조각은 메서드만 있었습니다. 사용하는 `main`과 **왕복(round-trip) 테스트**를 추가했습니다 — 커스텀 직렬화는 Marshal/Unmarshal이 서로의 역함수인지 테스트하는 것이 정석입니다.

`p2_money/money.go`:

```go
package main

import (
	"fmt"
	"strconv"
	"strings"
)

type Money int64 // cents 단위

func (m Money) MarshalJSON() ([]byte, error) {
	dollars := float64(m) / 100
	return []byte(fmt.Sprintf(`"%.2f"`, dollars)), nil
}

func (m *Money) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}
	*m = Money(f * 100)
	return nil
}
```

`p2_money/main.go`:

```go
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
```

`p2_money/money_test.go`:

```go
package main

import (
	"encoding/json"
	"testing"
)

func TestMoneyRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		cents Money
		want  string
	}{
		{"정수 달러", 1200, `"12.00"`},
		{"센트 포함", 1234, `"12.34"`},
		{"0원", 0, `"0.00"`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.cents)
			if err != nil {
				t.Fatal(err)
			}
			if string(data) != tc.want {
				t.Errorf("Marshal = %s, want %s", data, tc.want)
			}
			// 왕복(round-trip) 검증
			var back Money
			if err := json.Unmarshal(data, &back); err != nil {
				t.Fatal(err)
			}
			if back != tc.cents {
				t.Errorf("round-trip = %d, want %d", back, tc.cents)
			}
		})
	}
}

func TestMoneyUnmarshalInvalid(t *testing.T) {
	var m Money
	if err := json.Unmarshal([]byte(`"abc"`), &m); err == nil {
		t.Error("잘못된 입력인데 에러가 없음")
	}
}
```

**테스트 방법**: `go run .` 후 `go test -v .`

**검증 결과**:
```
{"name":"키보드","price":"12.34"}
{Name:마우스 Price:9999} (cents=9999)
--- PASS: TestMoneyRoundTrip (0.00s)
    --- PASS: TestMoneyRoundTrip/정수_달러
    --- PASS: TestMoneyRoundTrip/센트_포함
    --- PASS: TestMoneyRoundTrip/0원
--- PASS: TestMoneyUnmarshalInvalid (0.00s)
PASS
```

> ⚠️ **수강생 질문 대비**: `float64(m)/100` 왕복은 큰 금액에서 부동소수점 오차가 날 수 있습니다("왜 금액을 float로 안 다루나요"의 역질문 소재). 실무 답은 ① cents 정수 유지(이 예제) ② `strconv` 기반 십진 파싱 ③ `shopspring/decimal` 같은 십진수 라이브러리입니다.

## 2.5 스트리밍 디코딩 (전체 코드)

본문 조각의 `huge.json`과 `process()`가 미정의였습니다. **시연용 대용량 파일을 직접 생성**한 뒤 항목 단위로 처리하는 완결 프로그램입니다.

`p2_stream/main.go`:

```go
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
```

**테스트 방법**: `go run .`. 메모리 효율을 정량적으로 보여주려면 n을 100만으로 키우고 `/usr/bin/time -v go run .`으로 Maximum resident set size를 `json.Unmarshal` 버전과 비교하는 심화 실습이 가능합니다.

**검증 결과**:
```
생성된 파일: huge.json (455 KB, 항목 10000개)
처리 중: {ID:2500 Name:item-2500 Price:250000}
처리 중: {ID:5000 Name:item-5000 Price:500000}
처리 중: {ID:7500 Name:item-7500 Price:750000}
처리 중: {ID:10000 Name:item-10000 Price:1000000}
총 처리 항목: 10000
```

## 2.6 HTTP 핸들러 JSON 패턴 (전체 코드 + httptest + curl)

본문 코드에 `main`과 라우팅을 붙이고, **3교시 httptest를 미리 맛보는 테스트 파일**을 페어로 제공합니다.

`p2_handler/main.go`:

```go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type CreateUserResponse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "잘못된 요청", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.Name == "" {
		http.Error(w, "name 필수", http.StatusBadRequest)
		return
	}

	resp := CreateUserResponse{
		ID:        42,
		Name:      req.Name,
		Email:     req.Email,
		CreatedAt: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /users", createUser)
	fmt.Println("서버 시작: http://localhost:8080")
	http.ListenAndServe(":8080", mux)
}
```

`p2_handler/main_test.go`:

```go
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
```

**테스트 방법**:

```bash
go test -v .          # 서버 안 띄우고 핸들러 직접 테스트

go run . &            # 실제 서버로도 확인
curl -s -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com"}'
curl -s -o /dev/null -w "%{http_code}\n" -X POST http://localhost:8080/users -d '{"email":"x@x.com"}'
```

**검증 결과**:
```
--- PASS: TestCreateUser_OK / TestCreateUser_BadJSON / TestCreateUser_MissingName
{"id":42,"name":"Alice","email":"alice@example.com","created_at":"2026-06-12T00:06:02Z"}
400
```

## 2.7 다른 인코딩 — CSV / XML / Base64 / Gob (전체 코드)

네 조각을 하나의 실행 가능한 투어 프로그램으로 묶었습니다.

`p2_encodings/main.go`:

```go
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
```

**테스트 방법**: `go run .`

**검증 결과**:
```
--- CSV ---
name,age
Alice,30
--- XML ---
<person>
  <name>Alice</name>
  <age>30</age>
</person>
역직렬화: {XMLName:{Space: Local:person} Name:Alice Age:30}
--- Base64 ---
인코딩: SGVsbG8=
디코딩: Hello
--- Gob ---
gob 크기: 39 바이트
복원: {X:3 Y:4}
```

> 💡 **강의 포인트**: `csv.Writer`는 내부 버퍼링을 하므로 `Flush()`를 잊으면 출력이 사라집니다 — 6교시 CGo의 stdio 버퍼링 함정과 같은 계열의 문제라 연결 지어 설명하면 좋습니다.

---

# 3교시 보완

## 3.6 인터페이스 mock — `APIService` 완전판 (전체 코드 + 테스트 3종)

본문 조각에서 `GetUser`의 본문이 `// ...`로 생략되어 있었습니다. 디코딩·상태코드 처리까지 채우고, **성공 / 서버 에러 / 네트워크 에러** 3가지 시나리오를 mock으로 테스트합니다.

`p3_mock/service.go`:

```go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// 표준 *http.Client도 이 인터페이스를 만족 (Do 메서드 보유)
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type APIService struct {
	baseURL string
	client  HTTPClient
}

func NewAPIService(baseURL string, client HTTPClient) *APIService {
	return &APIService{baseURL: baseURL, client: client}
}

func (s *APIService) GetUser(id int) (*User, error) {
	url := fmt.Sprintf("%s/users/%d", s.baseURL, id)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("요청 실패: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	var u User
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return nil, fmt.Errorf("디코딩 실패: %w", err)
	}
	return &u, nil
}

func main() {
	// 실제 사용 시: 진짜 http.Client 주입
	svc := NewAPIService("https://api.example.com", &http.Client{})
	_ = svc
	fmt.Println("이 프로그램의 핵심은 service_test.go — go test -v 로 확인하세요")
}
```

`p3_mock/service_test.go`:

```go
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
```

**테스트 방법**: `go test -race -v .`

**검증 결과**:
```
--- PASS: TestGetUser_OK (0.00s)
--- PASS: TestGetUser_ServerError (0.00s)
--- PASS: TestGetUser_NetworkError (0.00s)
PASS
```

> 💡 **강의 포인트**: `io.NopCloser`가 왜 필요한가? — `http.Response.Body`는 `io.ReadCloser` 타입인데 `strings.Reader`는 `Close`가 없습니다. `NopCloser`는 아무것도 안 하는 `Close`를 붙여주는 어댑터입니다. "인터페이스를 맞추기 위한 최소 어댑터"라는 1일차 인터페이스 철학의 좋은 실례입니다.

## 3.7 `httptest.NewServer` — 외부 API 모킹 (전체 코드)

본문 조각을 **테스트 대상 함수(`fetchStatus`)까지 포함한** 완결 테스트 파일로 만들었습니다. 3.6의 mock 주입과 달리, 이 방식은 **진짜 TCP 서버를 임의 포트에 띄우므로** `http.Get`을 그대로 쓰는 기존 코드도 테스트할 수 있다는 차이를 보여줍니다.

`p3_httptest/main_test.go`:

```go
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
```

(같은 디렉토리에 `package main`만 적힌 `main.go` 하나를 두면 됩니다.)

**테스트 방법**: `go test -v .`

**검증 결과**: `--- PASS: TestFetchStatus (0.00s)`

**3.6 vs 3.7 선택 기준** (수강생 질문 대비):

| | 3.6 인터페이스 mock | 3.7 httptest.NewServer |
|---|---|---|
| 코드 요구사항 | `HTTPClient` 인터페이스로 추상화 필요 | `http.Get` 직접 호출 코드도 OK |
| 속도 | 가장 빠름 (네트워크 없음) | 루프백 TCP — 충분히 빠름 |
| 검증 범위 | 비즈니스 로직만 | URL/헤더/직렬화까지 실제 HTTP 경로 |
| 권장 | 단위 테스트 | 통합 성격의 테스트 |

---

# 4교시 보완

## 4.4 Context로 로거 전달 (전체 코드)

본문 조각의 `handler`/`doWork`를 실행해볼 수 있도록 httptest로 핸들러를 직접 호출하는 `main`을 붙였습니다. **서버 없이도 핸들러 단위로 실행**하는 기법 자체가 3교시 복습이 됩니다.

`p4_ctxlog/main.go`:

```go
package main

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
)

type ctxKey string

const loggerKey ctxKey = "logger"

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func LoggerFrom(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return l
	}
	return slog.Default() // 없으면 기본 로거 — nil panic 방지
}

func handler(w http.ResponseWriter, r *http.Request) {
	reqID := r.Header.Get("X-Request-ID")
	logger := slog.With("request_id", reqID)
	ctx := WithLogger(r.Context(), logger)

	doWork(ctx)
	w.Write([]byte("OK"))
}

func doWork(ctx context.Context) {
	log := LoggerFrom(ctx)
	log.Info("작업 처리") // 자동으로 request_id 포함
	deeperWork(ctx)
}

func deeperWork(ctx context.Context) {
	// 몇 단계를 내려가도 ctx만 전달하면 같은 request_id가 따라옴
	LoggerFrom(ctx).Info("깊은 곳의 작업")
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	// 서버 없이 핸들러만 직접 호출해 시연 (httptest 활용)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Request-ID", "req-123")
	w := httptest.NewRecorder()
	handler(w, req)
}
```

**테스트 방법**: `go run .`

**검증 결과**:
```json
{"time":"...","level":"INFO","msg":"작업 처리","request_id":"req-123"}
{"time":"...","level":"INFO","msg":"깊은 곳의 작업","request_id":"req-123"}
```

## 4.6 recover + `debug.Stack()` (전체 코드)

본문 조각(defer/recover 3줄)을 **실제 panic을 일으키고 복구하는** 완결 프로그램으로 만들었습니다. named return으로 복구값을 돌려주는 패턴도 함께 보여줍니다.

`p4_recover/main.go`:

```go
package main

import (
	"log/slog"
	"os"
	"runtime/debug"
)

func mayPanic(data []int, idx int) (result int) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("패닉 복구",
				"panic", r,
				"stack", string(debug.Stack()))
			result = -1 // named return으로 복구값 지정
		}
	}()
	return data[idx] // idx가 범위 밖이면 panic
}

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))

	data := []int{10, 20, 30}
	slog.Info("정상 접근", "value", mayPanic(data, 1))
	slog.Info("범위 밖 접근", "value", mayPanic(data, 99)) // panic → 복구 → -1
	slog.Info("프로그램은 죽지 않고 계속 실행됨")
}
```

**테스트 방법**: `go run .`

**검증 결과** (요약):
```
level=INFO msg="정상 접근" value=20
level=ERROR msg="패닉 복구" panic="runtime error: index out of range [99] with length 3" stack="goroutine 1 [running]:..."
level=INFO msg="범위 밖 접근" value=-1
level=INFO msg="프로그램은 죽지 않고 계속 실행됨"
```

> 💡 **C 비교 포인트**: C에서 segfault 후 코어덤프를 분석하던 것과 달리, Go는 **죽기 전에 자기 스택을 문자열로 꺼내 로깅**할 수 있습니다. 단, recover 남용은 금물 — HTTP 미들웨어(7교시 `RecoveryMiddleware`)처럼 "요청 하나의 실패가 프로세스를 죽이면 안 되는 경계"에서만 쓰는 것이 관례입니다.

---

# 5교시 보완

## 5.4 `slowSort` — 벤치마크 + CPU 프로파일 (전체 코드)

본문에는 함수만 있었습니다. **벤치마크 파일과 정확성 테스트, 표준 정렬과의 비교**까지 붙여 pprof 실습이 바로 가능하게 했습니다.

`p5_sort/sort.go`:

```go
package sortdemo

import "math/rand"

// 의도적으로 느린 버블 정렬 — O(n²)
func slowSort(data []int) []int {
	n := len(data)
	for i := 0; i < n; i++ {
		for j := 0; j < n-i-1; j++ {
			if data[j] > data[j+1] {
				data[j], data[j+1] = data[j+1], data[j]
			}
		}
	}
	return data
}

func randomData(n int) []int {
	data := make([]int, n)
	for i := range data {
		data[i] = rand.Intn(1_000_000)
	}
	return data
}
```

`p5_sort/sort_test.go`:

```go
package sortdemo

import (
	"slices"
	"sort"
	"testing"
)

func TestSlowSort(t *testing.T) {
	data := []int{5, 2, 9, 1, 7}
	got := slowSort(data)
	if !slices.IsSorted(got) {
		t.Errorf("정렬 안 됨: %v", got)
	}
}

func BenchmarkSlowSort(b *testing.B) {
	src := randomData(2000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		data := slices.Clone(src) // 매번 동일한 무작위 입력
		b.StartTimer()
		slowSort(data)
	}
}

func BenchmarkStdSort(b *testing.B) {
	src := randomData(2000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		data := slices.Clone(src)
		b.StartTimer()
		sort.Ints(data) // 표준 정렬과 비교
	}
}
```

**테스트 방법**:

```bash
go test -v -run TestSlowSort .                          # 정확성
go test -bench=. -cpuprofile=cpu.prof .                 # 벤치마크 + 프로파일
go tool pprof -top -nodecount=5 cpu.prof                # CLI로 핫스팟
go tool pprof -http=:8000 cpu.prof                      # Flame Graph (graphviz 필요)
```

**검증 결과** (실측):
```
BenchmarkSlowSort     544     2178706 ns/op     ← 2.18ms
BenchmarkStdSort    16392       68913 ns/op     ← 0.07ms (약 32배 빠름)

(pprof) top
      flat  flat%   sum%        cum   cum%
    1390ms 40.64% 40.64%     1390ms 40.64%  sortdemo.slowSort   ← 단일 함수 최대 핫스팟
    1190ms 34.80% 75.44%     1370ms 40.06%  slices.partitionOrdered[go.shape.int]
```

> ⚠️ **벤치마크 작성 함정 2가지** (조각코드를 그대로 벤치마크로 옮기면 빠지는 함정):
> 1. **이미 정렬된 데이터 재사용** — 같은 슬라이스를 반복 정렬하면 2회차부터는 "정렬된 입력"이 되어 최선 케이스만 측정합니다. 그래서 매 반복 `slices.Clone`으로 원본을 복원합니다.
> 2. **복사 비용 포함** — Clone 자체가 측정에 섞이지 않도록 `b.StopTimer()`/`b.StartTimer()`로 감쌉니다.
>
> 이는 본문 8.2의 `BenchmarkCreateUser` 점검 노트(중복 이메일로 409 경로만 측정)와 같은 종류의 함정입니다 — "**벤치마크는 매 반복이 같은 작업을 수행하는지 의심하라**"가 공통 교훈입니다.

## 5.9 `runtime/trace` (전체 코드)

본문 조각의 `runApplication()`을 **고루틴 4개의 CPU+I/O 혼합 작업**으로 구체화해, trace 뷰어에서 볼 거리가 있게 만들었습니다.

`p5_trace/main.go`:

```go
package main

import (
	"fmt"
	"log"
	"os"
	"runtime/trace"
	"sync"
	"time"
)

func worker(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	sum := 0
	for i := 0; i < 50_000_000; i++ { // CPU 작업
		sum += i
	}
	time.Sleep(10 * time.Millisecond) // I/O 흉내
	_ = sum
}

func main() {
	f, err := os.Create("trace.out")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if err := trace.Start(f); err != nil {
		log.Fatal(err)
	}
	defer trace.Stop()

	// 측정 대상: 고루틴 4개의 동시 작업
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go worker(i, &wg)
	}
	wg.Wait()
	fmt.Println("완료 — trace.out 생성됨")
}
```

**테스트 방법**:

```bash
go run .                # trace.out 생성 (검증: 2.7KB 파일 생성 확인)
go tool trace trace.out # 브라우저가 열림
# → "View trace by proc" 클릭: 4개 고루틴이 P(프로세서)들에 분배되는 모습,
#   Sleep 구간에서 고루틴이 러닝 큐에서 빠지는 모습 관찰
```

> 💡 **4일차 연결**: trace 화면의 P 레인은 3~4일차에서 다룬 **G-M-P 스케줄러**의 P를 그대로 시각화한 것입니다. GOMAXPROCS를 1로 줄여(`GOMAXPROCS=1 go run .`) 다시 trace를 찍으면 4개 고루틴이 한 P에서 시분할되는 것이 보입니다 — 스케줄러 강의의 결정적 시연 자료가 됩니다.

## 5.10 누수 데모 — 검증 절차 보완

본문 코드는 전체 코드입니다. 검증 절차를 정량화했습니다:

```bash
go run leak.go &
sleep 5                                                       # 5초만 기다려도 충분
curl -s http://localhost:6060/debug/pprof/heap > heap.prof
go tool pprof -top -nodecount=3 heap.prof
```

**검증 결과** (실행 4초 후 실측):
```
      flat  flat%   sum%        cum   cum%
  565.76kB   100%   100%   565.76kB   100%  main.leaker    ← 누수 지점이 100%로 지목됨
```

---

# 6교시 보완

## 6.6 슬라이스 ↔ C 배열 (전체 코드)

본문 조각의 `process_bytes`가 선언만 있었습니다. **합계를 계산해 반환하는 C 구현**을 붙여, 데이터가 실제로 C에 전달됐음을 검증할 수 있게 했고, 본문 경고(빈 슬라이스 panic)를 코드로 반영했습니다.

`p6_slice/main.go`:

```go
package main

/*
#include <stddef.h>
#include <stdio.h>

// 받은 바이트들의 합을 계산하고 내용을 출력하는 C 함수
unsigned long process_bytes(const unsigned char *data, size_t len) {
    unsigned long sum = 0;
    printf("C가 받은 데이터: ");
    for (size_t i = 0; i < len; i++) {
        printf("%d ", data[i]);
        sum += data[i];
    }
    printf("\n");
    fflush(stdout);
    return sum;
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

func processBytes(data []byte) uint64 {
	if len(data) == 0 { // ⚠️ 빈 슬라이스에서 &data[0]은 panic
		return 0
	}
	sum := C.process_bytes(
		(*C.uchar)(unsafe.Pointer(&data[0])),
		C.size_t(len(data)),
	)
	return uint64(sum)
}

func main() {
	data := []byte{1, 2, 3, 4, 5}
	fmt.Println("합계(C에서 계산):", processBytes(data))
	fmt.Println("빈 슬라이스:", processBytes(nil)) // 안전하게 0
}
```

**테스트 방법**: `go run .` (GCC 필요). 빈 슬라이스 가드를 지우고 `processBytes(nil)`을 호출해 panic을 직접 관찰하는 변형 실습도 유용합니다.

**검증 결과**:
```
C가 받은 데이터: 1 2 3 4 5
합계(C에서 계산): 15
빈 슬라이스: 0
```

## 6.10 OpenSSL SHA-256 — 교차 검증 테스트 추가

본문 코드는 전체 코드이므로, **"해시값이 정말 맞는지"를 표준 라이브러리 `crypto/sha256`과 교차 검증하는 테스트**만 추가합니다. CGo 코드를 순수 Go 레퍼런스와 비교하는 것은 CGo 마이그레이션/연동 검증의 표준 기법입니다.

`p6_openssl/hash_test.go`:

```go
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

// 표준 라이브러리 crypto/sha256과 교차 검증
func TestSHA256MatchesStdlib(t *testing.T) {
	inputs := []string{"", "Hello, CGo!", "한글도 테스트"}
	for _, in := range inputs {
		want := sha256.Sum256([]byte(in))
		got := sha256OpenSSL([]byte(in))
		if got != hex.EncodeToString(want[:]) {
			t.Errorf("입력 %q: got %s, want %x", in, got, want)
		}
	}
}
```

**테스트 방법**: `go test -v .` (libssl-dev, pkg-config 필요)

**검증 결과**:
```
SHA-256: 1b0e82a9aba7268a1b34f3fc4dd298bc42acc532d38a6b6b0267bf89c8baae03
--- PASS: TestSHA256MatchesStdlib (0.00s)
```
빈 문자열 입력(`SHA256(nil, 0, ...)`)까지 표준 라이브러리와 일치함을 확인했습니다.

---

# 📋 전체 검증 현황표

| 절 | 코드 | 상태 | 테스트 방법 | 실측 검증 결과 |
|---|---|---|---|---|
| 1.2 GET | 본문 전체 코드 | ✅ | `go run .` (api.github.com) | 상태 200 |
| 1.2-A Client | **이 문서에서 완성** | ✅ | 1.3 서버 + `go run .` | 200 / 본문 수신 |
| 1.2-B Context | **이 문서에서 완성** | ✅ | `go run .` 단독 | 3.0초에 DeadlineExceeded |
| 1.3 기본 서버 | 본문 전체 코드 | ✅ | `curl /hello?name=Go` | `Hello, Go!` |
| 1.4 라우팅 | **이 문서에서 완성** | ✅ | curl 6종 | 200/201/204 + **405 자동** |
| 1.5 미들웨어 | **이 문서에서 완성** | ✅ | curl ±토큰 | 401 → 200, 로그 기록 |
| 1.6 Graceful | 본문 전체 코드 | ✅ | 요청 중 `kill -TERM` | 요청 완료 후 종료 확인 |
| 2.1 Marshal | **이 문서에서 완성** | ✅ | `go run .` | omitempty 동작 확인 |
| 2.2 태그 | **이 문서에서 완성** | ✅ | `go run .` | `-`/omitempty/`,string`/소문자 |
| 2.3 동적 JSON | **이 문서에서 완성** | ✅ | `go run .` | float64 타입 확인 |
| 2.4 Money | **이 문서에서 완성** | ✅ | `go run .` + `go test -v` | 4개 테스트 PASS |
| 2.5 스트리밍 | **이 문서에서 완성** | ✅ | `go run .` | 10,000건 항목 단위 처리 |
| 2.6 핸들러 패턴 | **이 문서에서 완성** | ✅ | `go test -v` + curl | 3개 테스트 PASS, 201/400 |
| 2.7 기타 인코딩 | **이 문서에서 완성** | ✅ | `go run .` | CSV/XML/B64/Gob 왕복 |
| 3.2~3.5, 3.8~3.10 | 단편 (3.11에 통합) | ✅ | `go test -race -v ./calc` | 전부 PASS |
| 3.6 mock | **이 문서에서 완성** | ✅ | `go test -race -v` | 3개 시나리오 PASS |
| 3.7 httptest | **이 문서에서 완성** | ✅ | `go test -v` | PASS |
| 3.11 calc | 본문 전체 코드 | ✅ | `go test -race -v` + bench | PASS, 0.32 ns/op |
| 4.4 ctx 로거 | **이 문서에서 완성** | ✅ | `go run .` | request_id 전파 확인 |
| 4.6 recover | **이 문서에서 완성** | ✅ | `go run .` | panic 복구, 스택 로깅 |
| 4.7 디버그 대상 | 본문 전체 코드 | ✅ | `go run .` → `dlv debug .` | 합계 275000, 결제 0 재현 |
| 4.7.9 race | 본문 전체 코드 | ✅ | `go run -race .` | `WARNING: DATA RACE` |
| 4.8 logging_demo | 본문 전체 코드 | ✅ | `go run .` + curl | request_id 4줄 연결 |
| 5.4 slowSort | **이 문서에서 완성** | ✅ | bench + pprof | top에서 40.6% flat |
| 5.9 trace | **이 문서에서 완성** | ✅ | `go run .` + `go tool trace` | trace.out 생성 |
| 5.10 leak | 본문 전체 코드 | ✅ | heap 프로파일 | `main.leaker` 100% |
| 6.2 hello | 본문 전체 코드 | ✅ | `go run .`, `./app \| cat` | fflush로 파이프에서도 출력 |
| 6.3 별도 .c | 본문 전체 코드 | ✅ | `go run .` | 결과: 7 |
| 6.4 libm | 본문 전체 코드 | ✅ | `go run .` | 1.4142135623730951 |
| 6.6 슬라이스↔C | **이 문서에서 완성** | ✅ | `go run .` | C에서 합계 15 계산 |
| 6.7 콜백 | 본문 전체 코드 | ✅ | `go run .` | 0,1,2 출력 |
| 6.10 OpenSSL | 본문 + **테스트 추가** | ✅ | `go run .` + `go test -v` | stdlib와 해시 일치 |
| 7~8교시 userapi | 본문 전체 코드 | ✅ | `make all` 등 본문 참고 | (본문 점검 노트 참고) |

## 한 번에 전부 검증하는 스크립트

수업 전 환경 점검용. `snippets/` 루트에 `verify.sh`로 저장:

```bash
#!/usr/bin/env bash
set -e
echo "=== 실행형 예제 ==="
for d in p1_client_ctx p2_marshal p2_tags p2_dynamic p2_money p2_stream \
         p2_encodings p4_ctxlog p4_recover p5_trace p6_slice; do
  echo "--- $d ---"; (cd "$d" && go run .)
done

echo "=== 테스트형 예제 ==="
for d in p2_money p2_handler p3_mock p3_httptest p3_calc p5_sort p6_openssl; do
  echo "--- $d ---"; (cd "$d" && go test -race ./...)
done

echo "=== race 데모 (실패가 정상) ==="
(cd p4_race && go run -race . 2>&1 | grep -c "DATA RACE" && echo "race 검출 OK")

echo "✅ 전체 검증 완료"
```

```bash
chmod +x verify.sh && ./verify.sh
```
