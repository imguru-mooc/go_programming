# Go 언어 프로그래밍 5일차 — 표준 라이브러리, 테스트, 프로파일링, CGo, 종합 프로젝트

> **대상**: 1-4일차를 마친 C 개발자
> **목표**: 표준 라이브러리(HTTP/JSON)와 테스트·로깅·프로파일링 도구를 마스터하고, **CGo로 기존 C 코드와 연동**할 수 있으며, 최종적으로 동시성 기반 REST API 서버를 구축할 수 있다.
> **준비물**: 1-4일차 환경, `curl`, GCC  (CGo용)

---

## 📋 5일차 시간표

| 교시 | 주제 | 핵심 내용 |
|---|---|---|
| 1교시 | 표준 라이브러리 I | `net/http` 클라이언트/서버 |
| 2교시 | 표준 라이브러리 II | `encoding/json`, 데이터 직렬화 |
| 3교시 | 테스트와 벤치마크 심화 | 테이블 드리븐, mock, 커버리지 |
| 4교시 | 로깅과 디버깅 | `log/slog`, **Delve 디버깅 핸즈온** |
| 5교시 | 프로파일링 | `pprof`, CPU/메모리/고루틴 분석 |
| 6교시 | CGo | Go ↔ C 상호 호출 |
| 7교시 | 종합 실습 I | REST API 서버 — 구조 설계 + 핸들러 |
| 8교시 | 종합 실습 II | 서버 완성 + 배포 / 마무리 |

---

# 1교시. 표준 라이브러리 I — `net/http`

## 1.1 Go의 표준 라이브러리 철학

Go 표준 라이브러리의 한 줄 요약: **"Batteries Included."**

C에서 HTTP 서버를 만들려면 libmicrohttpd, mongoose 등 외부 라이브러리를 가져와야 했고, JSON 처리는 cJSON, jansson 등 또 다른 라이브러리를 써야 했습니다. Go는 **모두 표준 라이브러리에 내장**되어 있습니다.

| 기능 | C에서 흔한 선택 | Go 표준 라이브러리 |
|---|---|---|
| HTTP 서버 | libmicrohttpd, nginx-module | `net/http` |
| HTTP 클라이언트 | libcurl | `net/http` |
| JSON | cJSON, jansson | `encoding/json` |
| XML | libxml2 | `encoding/xml` |
| 암호화 | OpenSSL | `crypto/*` |
| 압축 | zlib | `compress/*` |
| 정규식 | PCRE | `regexp` |
| 데이터베이스 | libpq, mysql-client | `database/sql` (+ 드라이버) |

**외부 의존성을 최소화**하면서 운영 가능한 코드를 만들 수 있다는 점이 Go의 큰 매력입니다.

## 1.2 HTTP 클라이언트 — 가장 단순한 GET

```go
package main

import (
    "fmt"
    "io"
    "net/http"
)

func main() {
    resp, err := http.Get("https://api.github.com")
    if err != nil {
        fmt.Println("에러:", err)
        return
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        fmt.Println("읽기 실패:", err)
        return
    }

    fmt.Println("상태:", resp.StatusCode)
    fmt.Println("길이:", len(body), "바이트")
}
```

C로 같은 일을 libcurl로 해본 분이라면 이 짧음에 놀랄 겁니다. **`defer resp.Body.Close()`는 필수**입니다 — 잊으면 연결 누수가 발생합니다.

### 더 제어 필요할 때 — `http.Client`

```go
client := &http.Client{
    Timeout: 5 * time.Second,
}

req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
if err != nil {
    return err
}
req.Header.Set("User-Agent", "MyApp/1.0")
req.Header.Set("Authorization", "Bearer "+token)

resp, err := client.Do(req)
// ...
```

### Context와 결합 — 4일차 응용

```go
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()

req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
resp, err := http.DefaultClient.Do(req)
```

3초 후 자동 취소. 4일차의 Context가 자연스럽게 활용됩니다.

## 1.3 HTTP 서버 — 가장 단순한 것부터

```go
package main

import (
    "fmt"
    "net/http"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, %s!", r.URL.Query().Get("name"))
}

func main() {
    http.HandleFunc("/hello", helloHandler)
    fmt.Println("서버 시작: http://localhost:8080")
    http.ListenAndServe(":8080", nil)
}
```

실행:
```bash
go run server.go
# 다른 터미널에서
curl "http://localhost:8080/hello?name=Go"
# Hello, Go!
```

**이게 전부입니다.** 멀티스레딩, 비동기 I/O, 연결 관리 모두 알아서 처리됩니다.

### 핸들러의 시그니처

```go
func(w http.ResponseWriter, r *http.Request)
```

| 매개변수 | 의미 |
|---|---|
| `w` | 응답을 쓸 곳 (응답 본문, 헤더, 상태 코드) |
| `r` | 요청 정보 (URL, 헤더, 본문, 메서드, 쿼리) |

C에서 직접 `accept()` 하고 `recv()` 하고 `send()` 하던 일들이 모두 추상화되어 있습니다.

## 1.4 라우팅 — 메서드와 경로 분기

기본 `http.HandleFunc`은 경로만으로 라우팅합니다. 메서드 분기는 수동입니다.

### Go 1.22+ — 새 라우팅 문법

```go
mux := http.NewServeMux()

mux.HandleFunc("GET /users", listUsers)
mux.HandleFunc("POST /users", createUser)
mux.HandleFunc("GET /users/{id}", getUser)
mux.HandleFunc("PUT /users/{id}", updateUser)
mux.HandleFunc("DELETE /users/{id}", deleteUser)

http.ListenAndServe(":8080", mux)
```

핸들러 안에서 path parameter 추출:

```go
func getUser(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    fmt.Fprintf(w, "User ID: %s\n", id)
}
```

Go 1.22 이전엔 `gorilla/mux`, `chi` 같은 외부 라이브러리를 흔히 썼습니다. 이제는 표준 라이브러리만으로 충분합니다.

## 1.5 미들웨어 패턴

핸들러를 감싸는 함수로 공통 로직(로깅, 인증, 압축 등)을 처리합니다.

```go
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

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/api/secret", secretHandler)

    // 미들웨어 체이닝
    handler := logging(auth(mux))
    http.ListenAndServe(":8080", handler)
}
```

미들웨어는 **데코레이터 패턴**입니다. 함수가 일급 객체인 Go의 특성을 잘 활용합니다.

## 1.6 Graceful Shutdown — 운영 필수

운영 환경에서는 SIGTERM 시 진행 중인 요청을 마무리하고 종료해야 합니다.

```go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
)

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        time.Sleep(2 * time.Second)
        w.Write([]byte("OK"))
    })

    srv := &http.Server{
        Addr:    ":8080",
        Handler: mux,
    }

    // 서버를 고루틴에서 시작
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatal(err)
        }
    }()
    log.Println("서버 시작")

    // 종료 시그널 대기
    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
    <-sigs

    log.Println("종료 시작...")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        log.Println("강제 종료:", err)
    }
    log.Println("정상 종료")
}
```

`srv.Shutdown(ctx)`은:
- 새 연결 거부
- 활성 연결의 요청 처리 완료까지 대기
- ctx 만료되면 강제 종료

### ✅ 1교시 체크포인트

- [ ] `http.Get`과 `http.Client.Do`를 구분해 쓸 수 있는가?
- [ ] Go 1.22 새 라우팅 문법으로 RESTful 경로를 정의할 수 있는가?
- [ ] 미들웨어 체이닝을 구현할 수 있는가?
- [ ] `srv.Shutdown`으로 graceful shutdown을 구현할 수 있는가?

---

# 2교시. 표준 라이브러리 II — `encoding/json`과 직렬화

## 2.1 JSON 기본 — `Marshal` / `Unmarshal`

```go
import "encoding/json"

type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
    Age  int    `json:"age,omitempty"`
}

// Go → JSON
u := User{ID: 1, Name: "Alice", Age: 30}
data, _ := json.Marshal(u)
fmt.Println(string(data))
// {"id":1,"name":"Alice","age":30}

// JSON → Go
jsonStr := `{"id":2,"name":"Bob"}`
var u2 User
json.Unmarshal([]byte(jsonStr), &u2)
fmt.Printf("%+v\n", u2)
// {ID:2 Name:Bob Age:0}
```

C에서 cJSON으로 매번 `cJSON_GetObjectItem`, `cJSON_GetStringValue` 하던 번거로움이 완전히 사라집니다.

## 2.2 구조체 태그 — 직렬화 제어

`` `json:"필드명,옵션"` `` 형식의 태그로 동작을 제어합니다.

```go
type User struct {
    ID        int       `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email,omitempty"`     // 비어있으면 생략
    Password  string    `json:"-"`                    // 항상 제외
    CreatedAt time.Time `json:"created_at"`
    Internal  string    `json:"internal,omitempty"`
}
```

| 옵션 | 효과 |
|---|---|
| `"name"` | JSON 필드명을 `name`으로 |
| `"name,omitempty"` | zero value면 생략 |
| `"-"` | 항상 제외 (예: 비밀번호) |
| `",string"` | 숫자를 문자열로 인코딩 (JS의 큰 정수 문제) |

### 가시성 규칙

`json` 패키지는 **외부 패키지**이므로 reflection으로 필드를 봅니다. **소문자 필드는 안 보입니다.**

```go
type Bad struct {
    name string `json:"name"`  // ❌ 직렬화 안 됨 (소문자)
    Age  int    `json:"age"`   // ✅ OK
}
```

C 코드에서 옮길 때 자주 빠지는 실수입니다.

## 2.3 동적 JSON — `map[string]interface{}` 또는 `any`

구조가 정해지지 않은 JSON을 다룰 때.

```go
var data map[string]any
json.Unmarshal([]byte(`{"name":"Alice","age":30,"tags":["go","c"]}`), &data)

fmt.Println(data["name"])  // Alice
fmt.Println(data["age"])   // 30 (실제 타입은 float64!)
fmt.Println(data["tags"])  // [go c]
```

### 🔥 함정 — 모든 숫자는 `float64`로 디코딩

```go
n := data["age"].(float64)  // ✅
n := data["age"].(int)      // ❌ panic
```

이는 JSON 표준에 정수와 실수 구분이 없기 때문입니다. 정수가 필요하면 변환하거나, 구조체에 명시적 타입을 정의하세요.

## 2.4 커스텀 직렬화 — `Marshaler` / `Unmarshaler`

특수 형식이 필요할 때 인터페이스를 구현합니다.

```go
type Money int64  // cents 단위

func (m Money) MarshalJSON() ([]byte, error) {
    // 1234 → "12.34"
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

이제 `Money`는 자동으로 사용자 친화적 형식으로 직렬화됩니다.

## 2.5 스트리밍 디코딩 — 대용량 JSON

`json.Unmarshal`은 전체를 메모리에 올립니다. 큰 파일이나 스트림에선 `json.Decoder`를 씁니다.

```go
file, _ := os.Open("huge.json")
defer file.Close()

decoder := json.NewDecoder(file)

// 배열의 시작 토큰 읽기
decoder.Token()  // [

for decoder.More() {
    var item Item
    if err := decoder.Decode(&item); err != nil {
        log.Fatal(err)
    }
    process(item)
}
```

기가바이트급 JSON도 메모리 부담 없이 처리할 수 있습니다.

## 2.6 HTTP 핸들러에서 JSON 다루기 — 실전 패턴

```go
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

    // 요청 본문 디코딩
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "잘못된 요청", http.StatusBadRequest)
        return
    }
    defer r.Body.Close()

    // 유효성 검사
    if req.Name == "" {
        http.Error(w, "name 필수", http.StatusBadRequest)
        return
    }

    // 비즈니스 로직 (DB 저장 등 생략)
    resp := CreateUserResponse{
        ID:        42,
        Name:      req.Name,
        Email:     req.Email,
        CreatedAt: time.Now(),
    }

    // 응답 인코딩
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(resp)
}
```

이 패턴이 REST API 핸들러의 표준 구조입니다. 7교시에서 본격적으로 활용합니다.

## 2.7 다른 인코딩 — 빠른 소개

표준 라이브러리는 JSON 외에도 다양한 형식을 지원합니다.

```go
// CSV
import "encoding/csv"
w := csv.NewWriter(os.Stdout)
w.Write([]string{"name", "age"})
w.Write([]string{"Alice", "30"})
w.Flush()

// XML
import "encoding/xml"
type Person struct {
    XMLName xml.Name `xml:"person"`
    Name    string   `xml:"name"`
    Age     int      `xml:"age"`
}
xml.Marshal(p)

// Base64
import "encoding/base64"
enc := base64.StdEncoding.EncodeToString([]byte("Hello"))

// Gob (Go 전용 효율적 바이너리)
import "encoding/gob"
gob.NewEncoder(w).Encode(data)
```

### ✅ 2교시 체크포인트

- [ ] 구조체 태그로 JSON 필드를 제어할 수 있는가?
- [ ] 동적 JSON에서 숫자가 `float64`임을 기억하는가?
- [ ] 커스텀 `MarshalJSON`을 구현할 수 있는가?
- [ ] HTTP 핸들러의 표준 JSON 패턴을 작성할 수 있는가?

---

# 3교시. 테스트와 벤치마크 심화

## 3.1 Go 테스트의 철학

Go는 **테스트를 언어 기본**으로 다룹니다. 별도 프레임워크 설치 없이 `go test`만 알면 됩니다.

| 규칙 | 내용 |
|---|---|
| 파일 이름 | `*_test.go` (자동 인식) |
| 함수 이름 | `TestXxx`, `BenchmarkXxx`, `ExampleXxx`, `FuzzXxx` |
| 매개변수 | `*testing.T`, `*testing.B`, `*testing.F` |
| 위치 | 테스트 대상과 **같은 패키지** (또는 `_test` 패키지) |

## 3.2 기본 테스트

```go
// calc.go
package calc

func Add(a, b int) int {
    return a + b
}
```

```go
// calc_test.go
package calc

import "testing"

func TestAdd(t *testing.T) {
    got := Add(2, 3)
    want := 5
    if got != want {
        t.Errorf("Add(2,3) = %d; want %d", got, want)
    }
}
```

```bash
go test                # 현재 패키지
go test ./...          # 전체 프로젝트
go test -v             # 상세 출력
go test -run TestAdd   # 특정 테스트만
```

### `t.Error` vs `t.Fatal`

```go
t.Error("실패")  // 표시하고 계속
t.Errorf("...")
t.Fatal("실패")  // 표시하고 즉시 중단
t.Fatalf("...")
```

**`Fatal`은 이후 단계가 의미 없을 때만**(설정 실패 등) 사용. 보통은 `Error`로 여러 검증을 한 번에 수행합니다.

## 3.3 테이블 드리븐 테스트 — Go 관용구

같은 함수에 여러 입력을 테스트하는 표준 패턴입니다.

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name string
        a, b int
        want int
    }{
        {"양수+양수", 2, 3, 5},
        {"음수+음수", -2, -3, -5},
        {"부호 혼합", -5, 3, -2},
        {"0 포함", 0, 7, 7},
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            got := Add(tc.a, tc.b)
            if got != tc.want {
                t.Errorf("Add(%d,%d) = %d; want %d",
                    tc.a, tc.b, got, tc.want)
            }
        })
    }
}
```

### `t.Run`의 장점

```bash
go test -v
# === RUN   TestAdd
# === RUN   TestAdd/양수+양수
# --- PASS: TestAdd/양수+양수 (0.00s)
# === RUN   TestAdd/음수+음수
# --- PASS: TestAdd/음수+음수 (0.00s)
# ...

# 특정 서브테스트만
go test -run TestAdd/음수+음수
```

각 케이스가 별도 테스트로 인식되어, 실패 시점과 위치를 정확히 알 수 있습니다.

## 3.4 병렬 테스트 — `t.Parallel()`

CPU를 활용해 테스트를 빠르게.

```go
func TestExpensive(t *testing.T) {
    t.Parallel()  // 다른 Parallel 테스트와 동시 실행

    // ...
}
```

### 함정 — 테이블 드리븐에서 루프 변수 캡처 (Go 1.21 이하)

```go
for _, tc := range tests {
    tc := tc  // 1.22 미만: 반드시 재선언
    t.Run(tc.name, func(t *testing.T) {
        t.Parallel()
        // tc 사용
    })
}
```

Go 1.22+에서는 루프 변수 재선언이 불필요합니다(3일차 참고).

## 3.5 테스트 헬퍼 — `t.Helper()`

공통 검증 로직을 함수로 분리할 때.

```go
func assertEqual(t *testing.T, got, want int) {
    t.Helper()  // 에러 발생 시 줄 번호를 호출 측에 표시
    if got != want {
        t.Errorf("got %d, want %d", got, want)
    }
}

func TestSomething(t *testing.T) {
    assertEqual(t, Add(2, 3), 5)
    assertEqual(t, Add(0, 0), 0)
}
```

`t.Helper()` 없으면 실패 위치가 헬퍼 안으로 표시되어 디버깅이 어려워집니다.

## 3.6 Mock — 인터페이스 활용

Go에서는 별도 mock 라이브러리 없이 **인터페이스 + 직접 구현**으로 충분한 경우가 많습니다.

### 예: HTTP 클라이언트 의존성 분리

```go
// 인터페이스 정의 (1일차 인터페이스 활용)
type HTTPClient interface {
    Do(req *http.Request) (*http.Response, error)
}

// 실제 사용 코드
type APIService struct {
    client HTTPClient
}

func (s *APIService) GetUser(id int) (*User, error) {
    req, _ := http.NewRequest("GET", fmt.Sprintf("/users/%d", id), nil)
    resp, err := s.client.Do(req)
    // ...
}
```

### 테스트에서 mock 주입

```go
type mockClient struct {
    response *http.Response
    err      error
}

func (m *mockClient) Do(req *http.Request) (*http.Response, error) {
    return m.response, m.err
}

func TestGetUser(t *testing.T) {
    body := io.NopCloser(strings.NewReader(`{"id":1,"name":"Alice"}`))
    mock := &mockClient{
        response: &http.Response{
            StatusCode: 200,
            Body:       body,
        },
    }

    svc := &APIService{client: mock}
    user, err := svc.GetUser(1)

    if err != nil {
        t.Fatal(err)
    }
    if user.Name != "Alice" {
        t.Errorf("got %s, want Alice", user.Name)
    }
}
```

`testify/mock` 같은 라이브러리도 있지만, **간단한 경우 직접 작성이 더 명확**합니다.

## 3.7 `httptest` — HTTP 테스트의 표준 도구

```go
import "net/http/httptest"

func TestHelloHandler(t *testing.T) {
    req := httptest.NewRequest("GET", "/hello?name=Go", nil)
    w := httptest.NewRecorder()

    helloHandler(w, req)

    if w.Code != http.StatusOK {
        t.Errorf("status = %d", w.Code)
    }
    if !strings.Contains(w.Body.String(), "Hello, Go") {
        t.Errorf("body = %q", w.Body.String())
    }
}
```

`httptest.NewRecorder()`는 응답을 캡처하는 가짜 `http.ResponseWriter`입니다.

### 외부 API 모킹

```go
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    fmt.Fprint(w, `{"status":"ok"}`)
}))
defer server.Close()

// server.URL을 진짜 API처럼 사용
resp, _ := http.Get(server.URL + "/anything")
```

## 3.8 벤치마크

```go
func BenchmarkAdd(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Add(2, 3)
    }
}
```

`b.N`은 Go가 자동으로 조정합니다. 실행:

```bash
go test -bench=.
# BenchmarkAdd-8   1000000000   0.32 ns/op

go test -bench=. -benchmem
# BenchmarkAdd-8   1000000000   0.32 ns/op   0 B/op   0 allocs/op
```

| 컬럼 | 의미 |
|---|---|
| `1000000000` | 실행된 횟수 |
| `0.32 ns/op` | 1회당 평균 시간 |
| `0 B/op` | 1회당 할당 바이트 |
| `0 allocs/op` | 1회당 할당 횟수 |

### 벤치마크 비교 — `benchstat`

여러 번 실행해 통계적으로 비교:

```bash
go install golang.org/x/perf/cmd/benchstat@latest

go test -bench=. -count=10 > old.txt
# 코드 변경 후
go test -bench=. -count=10 > new.txt

benchstat old.txt new.txt
# 변화량과 통계적 유의성 출력
```

## 3.9 커버리지

```bash
# 커버리지 측정
go test -cover ./...

# 상세 보고서
go test -coverprofile=cover.out ./...
go tool cover -html=cover.out -o cover.html
xdg-open cover.html  # 브라우저로 확인
```

각 줄이 실행됐는지 색칠된 HTML 보고서가 나옵니다.

## 3.10 Example 테스트 — 문서 + 테스트 동시에

```go
func ExampleAdd() {
    fmt.Println(Add(2, 3))
    // Output: 5
}
```

- 함수 사용 예시를 코드로 작성
- `// Output:` 주석과 실제 출력이 일치해야 통과
- `go doc`에서 자동으로 사용 예시로 표시됨

C에는 없는 우아한 메커니즘입니다.

## 3.11 🧪 실습 코드: `calc` 패키지 완전판

> ⚠️ **점검 노트**: 함수 정의를 `calc_test.go`에 넣으면 3.2의 `calc.go`와 **중복 선언 컴파일 에러**(`Add redeclared in this block`)가 발생합니다. 함수는 `calc.go`에, 테스트는 `calc_test.go`에 분리합니다.

`calc.go`:

```go
package calc

import "fmt"

func Add(a, b int) int { return a + b }
func Sub(a, b int) int { return a - b }
func Mul(a, b int) int { return a * b }
func Div(a, b int) (int, error) {
    if b == 0 {
        return 0, fmt.Errorf("0으로 나눌 수 없음")
    }
    return a / b, nil
}
```

`calc_test.go`:

```go
package calc

import (
    "fmt"
    "testing"
)

// 표 기반 테스트
func TestBasicOps(t *testing.T) {
    tests := []struct {
        name     string
        op       func(int, int) int
        a, b     int
        want     int
    }{
        {"Add 양수", Add, 2, 3, 5},
        {"Add 음수", Add, -1, -1, -2},
        {"Sub", Sub, 10, 4, 6},
        {"Mul", Mul, 6, 7, 42},
    }
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()
            if got := tc.op(tc.a, tc.b); got != tc.want {
                t.Errorf("got %d, want %d", got, tc.want)
            }
        })
    }
}

// 에러 테스트
func TestDivByZero(t *testing.T) {
    _, err := Div(10, 0)
    if err == nil {
        t.Error("0으로 나눴는데 에러가 없음")
    }
}

// 벤치마크
func BenchmarkAdd(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Add(2, 3)
    }
}

// Example
func ExampleAdd() {
    fmt.Println(Add(2, 3))
    // Output: 5
}
```

실행 (검증 결과: 전부 PASS):
```bash
go test -race -v ./calc
go test -bench=. ./calc
```

> 💡 `t.Parallel()`을 테이블 드리븐과 함께 쓰는 위 코드는 **Go 1.22+ 기준**입니다. 1.21 이하에서는 루프 안에 `tc := tc`를 추가하세요(3.4 참고).

### ✅ 3교시 체크포인트

- [ ] 테이블 드리븐 테스트 패턴을 자유롭게 쓸 수 있는가?
- [ ] `t.Parallel`과 `t.Helper`를 적절히 활용할 수 있는가?
- [ ] 인터페이스로 외부 의존성을 mock할 수 있는가?
- [ ] `httptest`로 HTTP 핸들러를 테스트할 수 있는가?
- [ ] 벤치마크 결과를 해석할 수 있는가?

---

# 4교시. 로깅과 디버깅 — `log/slog`

## 4.1 옛 `log` 패키지의 한계

```go
import "log"

log.Printf("user=%s action=%s", user, action)
```

문제:
- 평문 로그 — 파싱 불편
- 레벨 구분 없음 (DEBUG/INFO/WARN/ERROR)
- 구조화된 필드 없음
- 컨텍스트 통합 불편

## 4.2 `log/slog` — Go 1.21+ 구조화 로깅

Go 1.21부터 표준 라이브러리에 들어왔습니다. **structured logging**의 표준입니다.

```go
import "log/slog"

slog.Info("로그인",
    "user_id", 42,
    "ip", "192.168.1.1",
    "duration_ms", 123)
```

기본 출력 (텍스트):
```
time=2024-01-15T10:30:45.123Z level=INFO msg=로그인 user_id=42 ip=192.168.1.1 duration_ms=123
```

JSON 출력으로 전환:
```go
slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

slog.Info("로그인", "user_id", 42)
// {"time":"2024-01-15T...","level":"INFO","msg":"로그인","user_id":42}
```

운영 환경에서는 보통 JSON 핸들러를 씁니다. Elasticsearch, Loki, CloudWatch 등에 그대로 인입됩니다.

## 4.3 로그 레벨

```go
slog.Debug("디버그 정보")  // 보통 운영 환경에선 숨김
slog.Info("일반 정보")
slog.Warn("경고")
slog.Error("에러", "err", err)
```

레벨 필터링:
```go
opts := &slog.HandlerOptions{
    Level: slog.LevelWarn,  // Warn 이상만 출력
}
slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, opts)))
```

## 4.4 그룹과 컨텍스트

여러 로그에 공통 필드를 자동으로 붙이고 싶을 때.

```go
logger := slog.With(
    "service", "api",
    "version", "1.0.0",
)

logger.Info("요청 수신", "path", "/users")
// service=api version=1.0.0 path=/users msg="요청 수신"

userLogger := logger.With("user_id", 42)
userLogger.Info("프로필 조회")
// service=api version=1.0.0 user_id=42 msg="프로필 조회"
```

### Context와 결합

요청 ID 같은 컨텍스트 정보를 로그에 자동 포함:

```go
type ctxKey string
const loggerKey ctxKey = "logger"

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
    return context.WithValue(ctx, loggerKey, logger)
}

func LoggerFrom(ctx context.Context) *slog.Logger {
    if l, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
        return l
    }
    return slog.Default()
}

// 핸들러에서 - request ID 포함
func handler(w http.ResponseWriter, r *http.Request) {
    reqID := r.Header.Get("X-Request-ID")
    logger := slog.With("request_id", reqID)
    ctx := WithLogger(r.Context(), logger)

    doWork(ctx)
}

func doWork(ctx context.Context) {
    log := LoggerFrom(ctx)
    log.Info("작업 처리")  // 자동으로 request_id 포함
}
```

## 4.5 에러 로깅 패턴

```go
if err != nil {
    slog.Error("DB 쿼리 실패",
        "err", err,
        "query", query,
        "user_id", userID,
    )
    return err
}
```

**구조화된 필드**로 로그를 남기면, 나중에 `grep`이 아니라 **쿼리**로 분석할 수 있습니다.

```bash
# JSON 로그라면
jq 'select(.level=="ERROR" and .user_id==42)' logs.json
```

## 4.6 디버깅 기법

### `fmt.Printf` 기반 — 가장 기본

```go
fmt.Printf("DEBUG: x=%+v y=%#v\n", x, y)
```

| 동사 | 출력 |
|---|---|
| `%v` | 기본 표현 |
| `%+v` | 구조체 필드명 포함 |
| `%#v` | Go 구문 형태 |
| `%T` | 타입 |

### `runtime/debug` — 스택 트레이스

```go
import "runtime/debug"

defer func() {
    if r := recover(); r != nil {
        slog.Error("패닉 복구", "panic", r, "stack", string(debug.Stack()))
    }
}()
```

### Delve — Go 디버거

GDB 같은 역할의 Go 전용 디버거.

```bash
go install github.com/go-delve/delve/cmd/dlv@latest

dlv debug ./cmd/myapp
# (dlv) break main.main
# (dlv) continue
# (dlv) next
# (dlv) print someVar
```

VS Code, GoLand 등 IDE에서 그래픽 디버깅도 지원합니다. **바로 다음 4.7절에서 Delve를 처음부터 끝까지 따라하는 핸즈온 실습을 진행합니다.**

## 4.7 🔬 Go 디버깅 따라하기 — Delve 핸즈온

`fmt.Printf` 디버깅만으로는 한계가 있습니다. C에서 GDB를 쓰듯, Go에서는 **Delve(dlv)**를 씁니다. 이 절은 처음부터 끝까지 그대로 따라할 수 있는 실습입니다.

### 4.7.1 준비 — 설치와 디버그 대상 코드

```bash
# Delve 설치
go install github.com/go-delve/delve/cmd/dlv@latest

# PATH 확인 (~/go/bin이 PATH에 있어야 함)
dlv version
```

실습용 버그 프로그램을 만듭니다. **10% 할인을 의도했는데 결제 금액이 0원이 나오는 버그**가 숨어 있습니다.

```bash
mkdir -p ~/go-class/day5/dbg && cd ~/go-class/day5/dbg
go mod init dbg
```

`main.go`:

```go
package main

import "fmt"

type Item struct {
    Name  string
    Price int
    Qty   int
}

func total(items []Item) int {
    sum := 0
    for _, it := range items {
        sum += it.Price * it.Qty
    }
    return sum
}

func applyDiscount(sum int, rate float64) int {
    // 버그: 의도는 10% 할인인데 rate를 잘못 사용
    return sum - int(float64(sum)*rate*10)
}

func main() {
    items := []Item{
        {"키보드", 30000, 2},
        {"마우스", 15000, 1},
        {"모니터", 200000, 1},
    }
    sum := total(items)
    final := applyDiscount(sum, 0.1)
    fmt.Println("합계:", sum)
    fmt.Println("결제 금액:", final)
}
```

먼저 그냥 실행해서 증상을 확인합니다:

```bash
go run .
# 합계: 275000
# 결제 금액: 0        ← ❌ 10% 할인이면 247500이어야 하는데?
```

### 4.7.2 첫 디버깅 세션 — 중단점, 실행 제어, 변수 확인

```bash
dlv debug .
```

`dlv debug`는 최적화/인라이닝을 끄고(`-gcflags="all=-N -l"` 자동 적용) 빌드한 뒤 디버거 셸을 띄웁니다. 이제 GDB와 거의 같은 흐름으로 진행합니다:

```
(dlv) break main.applyDiscount        # ① 함수에 중단점
Breakpoint 1 set at 0x... for main.applyDiscount() ./main.go:19

(dlv) continue                        # ② 중단점까지 실행
> main.applyDiscount() ./main.go:19 (hits goroutine(1):1 total:1)

(dlv) args                            # ③ 함수 인자 확인
sum = 275000
rate = 0.1

(dlv) next                            # ④ 한 줄 실행 (GDB의 n)
(dlv) print int(float64(sum)*rate*10) # ⑤ 표현식 평가!
275000                                # ← 할인액이 합계 전체와 같다 = 버그 원인 발견

(dlv) print float64(sum)*rate
27500                                 # ← 이게 의도한 10% 할인액. "*10"이 범인
```

| GDB | Delve | 의미 |
|---|---|---|
| `break` / `b` | `break` / `b` | 중단점 설정 |
| `run` | `continue` / `c` | 실행/재개 |
| `next` / `n` | `next` / `n` | 한 줄 실행 (함수 위로) |
| `step` / `s` | `step` / `s` | 함수 안으로 진입 |
| `finish` | `stepout` / `so` | 현재 함수 끝까지 |
| `print` / `p` | `print` / `p` | 변수/표현식 출력 |
| `info args` | `args` | 함수 인자 |
| `info locals` | `locals` | 지역 변수 전체 |
| `backtrace` / `bt` | `bt` | 호출 스택 |
| `frame N` | `frame N` | 스택 프레임 이동 |
| `list` | `list` / `ls` | 현재 위치 소스 |
| `watch` | `watch` | 변수 감시점 |
| - | `restart` / `r` | 프로세스 재시작 |

버그를 고치고(`rate*10` → `rate`) 디버거 안에서 바로 재확인:

```
(dlv) restart                         # 코드 수정 후 rebuild + 재시작 (dlv debug 모드)
(dlv) continue
(dlv) args
sum = 275000
rate = 0.1
(dlv) stepout                         # 함수 끝까지 실행하고 반환값 확인
Values returned:
    ~r0: 247500                       # ✅ 의도한 결과
(dlv) quit
```

### 4.7.3 조건부 중단점 — "1000번째 반복에서만 멈춰줘"

루프 안 버그를 잡을 때 매번 `continue`를 칠 수는 없습니다.

```
(dlv) break main.go:14                # total()의 루프 내부 줄 번호
(dlv) condition 1 it.Price > 100000   # 중단점 1번에 조건 부여
(dlv) continue
> main.total() ./main.go:14
(dlv) print it
main.Item {Name: "모니터", Price: 200000, Qty: 1}   # 조건 맞는 순간만 정지
```

한 줄로도 가능합니다: `break main.go:14 if it.Qty > 1`

### 4.7.4 고루틴 디버깅 — Delve의 진짜 가치

C에서 pthread 디버깅의 악몽을 기억한다면, 이 부분이 Delve의 백미입니다.

```
(dlv) goroutines                      # 모든 고루틴 목록
* Goroutine 1 - ... main.main
  Goroutine 18 - ... main.worker (대기 중인 위치 표시)
  ...

(dlv) goroutines -with user           # 사용자 코드 고루틴만 필터
(dlv) goroutine 18                    # 18번 고루틴으로 전환
(dlv) bt                              # 그 고루틴의 스택 추적
(dlv) goroutine 18 bt                 # 전환 없이 한 번에
```

**고루틴 누수 의심 시**: `goroutines` 출력에서 같은 함수·같은 줄에 멈춘 고루틴이 비정상적으로 많다면 그곳이 누수 지점입니다(5교시 pprof의 `goroutine?debug=2`와 같은 정보를 대화형으로 보는 셈).

### 4.7.5 실패하는 테스트 디버깅 — `dlv test`

테스트가 깨졌을 때 `Printf`를 심지 말고 디버거를 붙이세요.

```bash
cd ~/go-class/day5/userapi

# 특정 테스트만 디버깅
dlv test ./internal/user -- -test.run TestMemoryStore_CRUD
```

```
(dlv) break user.(*MemoryStore).Update    # 메서드 중단점: 패키지.(*타입).메서드
(dlv) continue
(dlv) print u
(dlv) print s.users                       # map 내용도 그대로 출력됨
(dlv) locals
```

### 4.7.6 실행 중인 서버에 붙기 — `dlv attach`

운영 중(혹은 다른 터미널에서 실행 중)인 프로세스를 그대로 디버깅합니다.

```bash
# 터미널 1: 7~8교시 userapi 서버 실행
./bin/userapi

# 터미널 2: PID 찾아서 attach
pgrep userapi          # 예: 12345
dlv attach 12345
```

```
(dlv) break userapi/internal/httpserver.(*Server).createUser
(dlv) continue
```

```bash
# 터미널 3: 요청을 쏘면 터미널 2의 중단점에서 멈춤
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Debug","email":"debug@example.com"}'
```

```
(dlv) print u                  # 디코딩된 요청 본문 확인
(dlv) bt                       # net/http 내부부터 핸들러까지 전체 스택
(dlv) continue                 # 요청 계속 처리
(dlv) quit                     # ⚠️ attach 종료 시 "프로세스를 죽일까?" 물음 → n 선택
```

> ⚠️ **주의**: 중단점에 멈춘 동안 그 고루틴의 요청은 블록됩니다. 운영 트래픽에는 attach 대신 로깅/pprof를 우선하세요.

### 4.7.7 디버그 빌드와 운영 빌드

`dlv debug`/`dlv test`는 자동 처리하지만, 미리 빌드한 바이너리를 `dlv exec`로 디버깅할 때는 직접 플래그를 줘야 합니다:

```bash
# 디버깅용 빌드: 최적화(-N), 인라이닝(-l) 비활성화
go build -gcflags="all=-N -l" -o bin/userapi-debug ./cmd/userapi

dlv exec bin/userapi-debug
```

최적화된 바이너리를 디버깅하면 변수가 `optimized out`으로 보이거나 줄 번호가 어긋납니다 — C에서 `-O2` 빌드를 GDB로 볼 때와 같은 현상입니다. 반대로 운영 배포 바이너리는 보통 `-ldflags="-s -w"`로 디버그 심볼을 제거합니다(8교시 Makefile 참고). **그 바이너리는 Delve로 디버깅할 수 없으니** 디버깅용 빌드를 따로 두세요.

### 4.7.8 VS Code에서 그래픽 디버깅

Go 확장을 설치하면 내부적으로 Delve(DAP)를 사용합니다. `.vscode/launch.json`:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "userapi 디버그",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/userapi"
    },
    {
      "name": "현재 패키지 테스트 디버그",
      "type": "go",
      "request": "launch",
      "mode": "test",
      "program": "${workspaceFolder}/internal/user"
    },
    {
      "name": "실행 중 프로세스에 attach",
      "type": "go",
      "request": "attach",
      "mode": "local",
      "processId": 0
    }
  ]
}
```

줄 번호 왼쪽 클릭으로 중단점, F5 시작, F10 next, F11 step — CLI에서 익힌 개념이 그대로 매핑됩니다. 원격 디버깅은 서버에서 `dlv dap --listen=:2345` 또는 `dlv exec --headless --listen=:2345 ./bin/app`을 띄우고 IDE에서 접속합니다.

### 4.7.9 데이터 레이스 디버깅 — `-race`와 함께

디버거로 잡기 가장 어려운 버그가 race입니다. Go는 전용 도구가 있습니다.

`race_demo.go`:

```go
package main

import (
    "fmt"
    "sync"
)

func main() {
    counter := 0
    var wg sync.WaitGroup
    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            counter++ // ❌ race!
        }()
    }
    wg.Wait()
    fmt.Println("counter =", counter)
}
```

```bash
go run -race race_demo.go
# ==================
# WARNING: DATA RACE
# Read at 0x00c0000140a8 by goroutine 8:
#   main.main.func1()  race_demo.go:15 +0x...
# Previous write at 0x00c0000140a8 by goroutine 7:
#   ...
# ==================
# counter = 1000        ← 결과가 맞아 보여도 race는 race!
```

**디버깅 워크플로 정리**:

1. 증상 재현 → `go run -race` / `go test -race` 먼저 (race 여부 확인)
2. 로직 버그 → `dlv debug` + 중단점 + `print` 표현식 평가
3. 특정 입력에서만 발생 → 조건부 중단점
4. 동시성 흐름 문제 → `goroutines` / `goroutine N bt`
5. 실행 중 프로세스 → `dlv attach` (운영은 신중히)
6. 성능 문제 → 디버거가 아니라 5교시 **pprof**로

## 4.8 🧪 실습 코드: `logging_demo.go`

```go
package main

import (
    "context"
    "log/slog"
    "net/http"
    "os"
    "time"
)

func main() {
    // JSON 로거 설정
    opts := &slog.HandlerOptions{Level: slog.LevelInfo}
    slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, opts)))

    mux := http.NewServeMux()
    mux.HandleFunc("/", logMiddleware(handler))

    slog.Info("서버 시작", "addr", ":8080")
    http.ListenAndServe(":8080", mux)
}

func logMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        reqID := r.Header.Get("X-Request-ID")
        if reqID == "" {
            reqID = "auto-" + time.Now().Format("150405")
        }

        logger := slog.With(
            "request_id", reqID,
            "method", r.Method,
            "path", r.URL.Path,
        )
        ctx := context.WithValue(r.Context(), ctxLoggerKey, logger)

        logger.Info("요청 수신")
        next(w, r.WithContext(ctx))
        logger.Info("요청 완료", "duration_ms", time.Since(start).Milliseconds())
    }
}

type ctxKey string
const ctxLoggerKey ctxKey = "logger"

func handler(w http.ResponseWriter, r *http.Request) {
    logger := r.Context().Value(ctxLoggerKey).(*slog.Logger)
    logger.Info("처리 중", "step", "1")
    time.Sleep(100 * time.Millisecond)
    logger.Info("처리 중", "step", "2")
    w.Write([]byte("OK"))
}
```

```bash
go run logging_demo.go &
curl http://localhost:8080/
```

로그 출력 (JSON):
```json
{"time":"...","level":"INFO","msg":"요청 수신","request_id":"auto-103045","method":"GET","path":"/"}
{"time":"...","level":"INFO","msg":"처리 중","request_id":"auto-103045","method":"GET","path":"/","step":"1"}
{"time":"...","level":"INFO","msg":"처리 중","request_id":"auto-103045","method":"GET","path":"/","step":"2"}
{"time":"...","level":"INFO","msg":"요청 완료","request_id":"auto-103045","method":"GET","path":"/","duration_ms":100}
```

모든 로그가 `request_id`로 연결됨을 확인하세요.

### ✅ 4교시 체크포인트

- [ ] `slog`로 구조화 로깅을 작성할 수 있는가?
- [ ] JSON 핸들러와 텍스트 핸들러를 구분해 쓸 수 있는가?
- [ ] Context로 로거를 전달하는 패턴을 이해했는가?
- [ ] Delve로 중단점을 걸고 `print`/`args`/`locals`로 변수를 확인할 수 있는가?
- [ ] 조건부 중단점과 `goroutine N bt`를 활용할 수 있는가?
- [ ] `dlv test`로 실패하는 테스트를, `dlv attach`로 실행 중 프로세스를 디버깅할 수 있는가?
- [ ] 디버그 빌드(`-gcflags="all=-N -l"`)와 운영 빌드(`-ldflags="-s -w"`)의 차이를 아는가?

---

# 5교시. 프로파일링 — `pprof`로 성능 분석

## 5.1 프로파일링이란?

C에서 `gprof`, `perf`, Valgrind 같은 도구를 썼던 경험이 있을 겁니다. Go도 비슷한 도구를 **표준 라이브러리**로 제공합니다.

| 종류 | 측정 대상 |
|---|---|
| CPU | 어느 함수에 시간을 가장 많이 썼는가 |
| Heap | 어디서 메모리를 가장 많이 할당하는가 |
| Goroutine | 어떤 고루틴이 어디서 멈춰있는가 |
| Block | 채널/뮤텍스 대기 시간 |
| Mutex | 락 경합 |

## 5.2 가장 쉬운 방법 — HTTP `pprof` 엔드포인트

```go
import (
    _ "net/http/pprof"  // import side effect로 핸들러 등록
    "net/http"
)

func main() {
    // pprof는 /debug/pprof/ 경로에 자동 등록됨
    go http.ListenAndServe(":6060", nil)

    // 본 애플리케이션 로직
    runApplication()
}
```

이제 브라우저에서:

```
http://localhost:6060/debug/pprof/
```

### CLI로 분석

```bash
# CPU 프로파일 (30초 수집)
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# 힙
go tool pprof http://localhost:6060/debug/pprof/heap

# 고루틴
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

`pprof` 셸에서:
```
(pprof) top
(pprof) list 함수명
(pprof) web      # 그래프를 브라우저로
(pprof) png      # PNG 파일로 저장
```

`web` 명령은 graphviz가 필요합니다(`apt install graphviz`).

## 5.3 벤치마크와 통합

벤치마크 실행 시 프로파일 자동 생성:

```bash
go test -bench=. -cpuprofile=cpu.prof -memprofile=mem.prof

go tool pprof cpu.prof
# (pprof) top10
# (pprof) list MyFunction
```

## 5.4 CPU 프로파일 분석 예시

```go
// 의도적으로 느린 함수
func slowSort(data []int) []int {
    // 버블 정렬 - O(n²)
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
```

```bash
go test -bench=BenchmarkSort -cpuprofile=cpu.prof
go tool pprof cpu.prof

(pprof) top
# Showing nodes accounting for 9000ms, 90.00% of 10000ms total
#       flat  flat%   sum%        cum   cum%
#     9000ms 90.00% 90.00%     9000ms 90.00%  main.slowSort
```

`slowSort`에 90%의 시간이 쓰이는 것이 명확히 보입니다.

## 5.5 메모리 프로파일

```bash
go tool pprof http://localhost:6060/debug/pprof/heap

(pprof) top
# 가장 많은 메모리 사용하는 함수들
```

### 누수 진단 방법

1. 시간 t1에 heap 스냅샷
2. 작업 수행
3. GC 후 시간 t2에 heap 스냅샷
4. 비교: 줄어들지 않은 부분이 누수 후보

```bash
curl http://localhost:6060/debug/pprof/heap > heap1.prof
# ... 작업 수행 후
curl http://localhost:6060/debug/pprof/heap > heap2.prof

go tool pprof -base heap1.prof heap2.prof
# 두 시점 사이의 차이만 표시
```

## 5.6 고루틴 누수 탐지

3일차에서 다룬 고루틴 누수를 실제로 잡는 방법:

```bash
curl http://localhost:6060/debug/pprof/goroutine?debug=2
```

각 고루틴의 스택 트레이스가 모두 출력됩니다. **같은 위치에서 멈춰있는 고루틴이 수천 개**라면 누수 의심.

## 5.7 `go test -race` 재강조

3일차에서 봤지만 다시 강조: **모든 테스트는 `-race`로 돌리세요**.

```bash
go test -race ./...
```

CI 파이프라인에 반드시 포함시킬 명령입니다.

## 5.8 flame graph

`pprof`의 그래픽 표현 중 가장 직관적인 형태.

```bash
go tool pprof -http=:8000 cpu.prof
# 브라우저에서 http://localhost:8000
# 좌측 메뉴: Top, Graph, Flame Graph, Source, Disassembly
```

Flame graph는 함수 호출 스택을 가로로 펼친 형태입니다. **넓은 막대 = 시간 많이 씀**.

## 5.9 `runtime/trace` — 더 정밀한 추적

```go
import "runtime/trace"

func main() {
    f, _ := os.Create("trace.out")
    defer f.Close()

    trace.Start(f)
    defer trace.Stop()

    // 측정할 코드
    runApplication()
}
```

```bash
go tool trace trace.out
# 브라우저 자동 열림
```

각 고루틴의 생명주기, GC 이벤트, syscall 등을 **시간순으로** 시각화합니다. 분산 트레이싱과는 다르지만, 단일 프로세스 내에서는 매우 강력합니다.

## 5.10 실전 — `pprof`로 메모리 누수 잡기

```go
// 누수 예제
package main

import (
    "log/slog"
    _ "net/http/pprof"
    "net/http"
    "time"
)

var leaked [][]byte

func leaker() {
    for {
        // 100KB씩 매초 누수
        leaked = append(leaked, make([]byte, 100_000))
        time.Sleep(time.Second)
    }
}

func main() {
    go http.ListenAndServe(":6060", nil)
    go leaker()

    slog.Info("실행 중 - http://localhost:6060/debug/pprof/")
    select {}  // 영원히 대기
}
```

```bash
go run leak.go &

# 1분 후
go tool pprof -http=:8000 http://localhost:6060/debug/pprof/heap

# 브라우저에서 Flame Graph 보면 leaker가 큰 막대로 표시
```

C 개발자에게 익숙한 `valgrind` 정도의 강력함을 표준 도구로 제공합니다.

### ✅ 5교시 체크포인트

- [ ] `net/http/pprof`를 통합할 수 있는가?
- [ ] CPU/Heap/Goroutine 프로파일을 구분해 쓸 수 있는가?
- [ ] `go tool pprof` 셸의 기본 명령을 알고 있는가?
- [ ] `flame graph`로 핫스팟을 식별할 수 있는가?

---

# 6교시. CGo — Go ↔ C 상호 호출

## 6.1 왜 CGo인가?

C 개발자에게 특히 중요한 주제입니다. 이미 검증된 C 라이브러리(OpenSSL, libsqlite3, OpenCV, FFmpeg, 사내 C 모듈 등)를 Go에서 그대로 쓸 수 있게 해주는 메커니즘이 **CGo**입니다.

**현실적 시나리오**:
- 회사에 10년 묵은 C 통신 모듈이 있음 → Go 서버에서 호출하고 싶음
- 하드웨어 SDK가 C로만 제공됨 → Go 앱에서 활용 필요
- C로 작성된 고성능 계산 라이브러리(BLAS 등) 사용

> **주의**: CGo는 강력하지만 **트레이드오프**가 있습니다. 가능하면 순수 Go로 두고, 정말 필요할 때만 쓰세요.

## 6.2 가장 단순한 CGo

```go
package main

/*
#include <stdio.h>
#include <stdlib.h>   // C.free에 반드시 필요!

void say_hello(const char *name) {
    printf("Hello from C, %s!\n", name);
    fflush(stdout);   // ⚠️ 아래 "출력 유실 함정" 참고
}
*/
import "C"

import "unsafe"

func main() {
    name := C.CString("Go")
    defer C.free(unsafe.Pointer(name))  // Go의 GC가 C 메모리 관리 안 함
    C.say_hello(name)
}
```

> ⚠️ **점검 노트 1**: `C.free`를 쓰려면 `#include <stdlib.h>`가 **반드시** 필요합니다. 빠뜨리면 `could not determine kind of name for C.free` 컴파일 에러. `unsafe.Pointer`를 쓰므로 `import "unsafe"`도 필요합니다.
>
> ⚠️ **점검 노트 2 — 출력 유실 함정 (실측 확인)**: C의 `printf`는 C 런타임의 stdio 버퍼를 씁니다. 터미널에서는 줄 단위로 즉시 출력되지만, **파이프/리다이렉트(`./demo | cat`, `./demo > log.txt`) 환경에서는 전체 버퍼링**되는데, Go 런타임이 종료할 때 C stdio 버퍼를 비우지 않아 **출력이 통째로 사라질 수 있습니다**. C 코드에서 `fflush(stdout)`을 호출하는 습관을 들이세요. CI 로그에서 "C 출력이 안 보여요"의 원인 1순위입니다.

실행:
```bash
go run main.go
# Hello from C, Go!
```

### 핵심 규칙

1. **주석 안에 C 코드** — `/* ... */` 안에 C 함수, `#include` 등
2. **`import "C"`** — 다른 import와 **반드시 분리된 단독 줄**, 주석 바로 다음
3. **`C.함수명`, `C.타입명`** — Go 코드에서 C 호출
4. **메모리 관리는 수동** — Go GC는 C가 할당한 메모리 추적 안 함

## 6.3 별도 `.c` 파일 사용

C 코드가 길어지면 별도 파일로 분리합니다.

`myadd.h`:
```c
#ifndef MYADD_H
#define MYADD_H
int my_add(int a, int b);
#endif
```

`myadd.c`:
```c
#include "myadd.h"

int my_add(int a, int b) {
    return a + b;
}
```

`main.go`:
```go
package main

/*
#include "myadd.h"
*/
import "C"

import "fmt"

func main() {
    result := C.my_add(C.int(3), C.int(4))
    fmt.Println("결과:", int(result))
}
```

`go build`가 자동으로 `.c` 파일을 컴파일해서 링크합니다. **별도 Makefile 불필요**.

## 6.4 외부 라이브러리 링크 — `#cgo` 지시어

```go
package main

/*
#cgo CFLAGS: -I/usr/include
#cgo LDFLAGS: -lm
#include <math.h>
*/
import "C"

import "fmt"

func main() {
    result := C.sqrt(C.double(2.0))
    fmt.Println(float64(result))  // 1.414...
}
```

| 지시어 | 의미 |
|---|---|
| `#cgo CFLAGS:` | C 컴파일 플래그 |
| `#cgo LDFLAGS:` | 링크 플래그 |
| `#cgo pkg-config:` | pkg-config 통합 |

`pkg-config` 활용:
```go
/*
#cgo pkg-config: openssl
#include <openssl/sha.h>
*/
import "C"
```

## 6.5 타입 변환표

| Go 타입 | C 타입 |
|---|---|
| `C.char` | `char` |
| `C.int` | `int` |
| `C.uint` | `unsigned int` |
| `C.long` | `long` |
| `C.size_t` | `size_t` |
| `C.float`, `C.double` | `float`, `double` |
| `*C.char` | `char *` (문자열) |
| `unsafe.Pointer` | `void *` |

### 문자열 변환

```go
// Go string → C string
cstr := C.CString("Hello")
defer C.free(unsafe.Pointer(cstr))

// C string → Go string
gostr := C.GoString(cstr)

// 길이 지정 변환
gostr := C.GoStringN(cstr, C.int(5))

// 바이트 슬라이스
goBytes := C.GoBytes(unsafe.Pointer(cArr), C.int(length))
```

**중요**: `C.CString`이 할당한 메모리는 **반드시 `C.free`로 해제**. 안 그러면 누수.

## 6.6 슬라이스 ↔ C 배열

> ⚠️ **점검 노트**: `import "C"`는 그룹 import에 넣으면 안 됩니다(6.2 핵심 규칙 2). cgo 전문(preamble) 주석은 **단독 `import "C"` 줄 바로 위**에 있어야 인식됩니다.

```go
/*
#include <stddef.h>
void process_bytes(const unsigned char *data, size_t len);
*/
import "C"

import "unsafe"

// Go 슬라이스 → C 배열로 전달
data := []byte{1, 2, 3, 4, 5}
C.process_bytes(
    (*C.uchar)(unsafe.Pointer(&data[0])),
    C.size_t(len(data)),
)
```

`&data[0]`은 슬라이스의 첫 원소 포인터. `unsafe.Pointer`로 변환 후 C 타입으로 캐스팅합니다.

**경고**: 슬라이스가 빈 슬라이스(`len(data) == 0`)일 때 `&data[0]`은 panic. 항상 확인.

## 6.7 콜백 — C에서 Go 함수 호출

C 라이브러리가 콜백 함수 포인터를 받는 경우.

> ⚠️ **점검 노트 (실측 확인)**: `//export`가 있는 Go 파일의 전문(preamble)에는 C 함수의 **선언만** 둘 수 있습니다. **정의(본문)를 넣으면 `multiple definition of 'run_with_callback'` 링크 에러**가 납니다 — cgo가 전문을 두 개의 생성 파일에 복사하기 때문입니다. 정의는 별도 `.c` 파일로 분리하세요.

`main.go`:

```go
package main

/*
extern void goCallback(int);     // Go 함수 선언 (//export로 노출됨)
void run_with_callback(void);    // C 함수는 "선언만"!
*/
import "C"

import "fmt"

//export goCallback
func goCallback(n C.int) {
    fmt.Println("Go에서 호출됨:", int(n))
}

func main() {
    C.run_with_callback()
}
```

`callback.c` (정의는 여기에):

```c
#include "_cgo_export.h"   // cgo가 자동 생성하는 헤더 — goCallback 선언 포함

void run_with_callback(void) {
    for (int i = 0; i < 3; i++) {
        goCallback(i);
    }
}
```

실행 (검증 결과):
```bash
go run .
# Go에서 호출됨: 0
# Go에서 호출됨: 1
# Go에서 호출됨: 2
```

**`//export 함수명`**으로 Go 함수를 C에 노출합니다.

## 6.8 CGo의 비용과 함정

### 함수 호출 오버헤드

| 호출 종류 | 비용 |
|---|---|
| Go → Go | ~1ns |
| Go → CGo → C | ~50~150ns |

**Go 함수 호출보다 100배 이상 느립니다.** 핫 루프에서 자주 호출하면 병목입니다. 가능한 한 **호출 한 번에 많은 데이터를 처리**하는 형태로 설계하세요.

### 메모리 관리

```go
// ❌ 위험 — Go가 GC하면 C가 보고 있는 메모리가 사라짐
p := unsafe.Pointer(&goSlice[0])
C.long_running_function(p)  // 도중에 GC 발생하면 위험
```

**규칙**:
- C로 전달된 Go 포인터는 **C 함수 호출 동안만 유효**
- C 코드가 포인터를 **저장**하려면 별도 복사 필요
- C에서 받은 메모리는 **반드시 C 방식으로 해제**

### 빌드 복잡도 증가

- CGo가 있으면 **크로스 컴파일이 어려워짐** (C 툴체인 필요)
- 정적 링크가 까다로워짐 — Docker `scratch` 이미지 사용 못 할 수 있음
- 빌드 속도 저하
- `CGO_ENABLED=0`으로 끄면 못 빌드됨

### 디버깅 어려움

- Go 디버거(Delve)가 C 코드 안에서는 제한적
- panic stack trace가 C 부분에서 끊김
- Race detector가 C 코드는 못 봄

## 6.9 대안 — 순수 Go 포팅 고려

CGo를 쓰기 전에 다음을 고민하세요.

1. **순수 Go 라이브러리가 있는가?** — 많은 C 라이브러리가 Go 포팅 존재
2. **별도 프로세스로 분리 가능?** — gRPC/REST로 통신
3. **호출 빈도가 적은가?** — 그렇다면 오버헤드 무시 가능
4. **C 코드 양이 적은가?** — 한 번 포팅해버리는 게 장기적으로 유리

### CGo가 명백히 옳은 경우

- 거대한 검증된 C 라이브러리(SQLite, OpenSSL 등) 그대로 사용
- 하드웨어 SDK가 C로만 제공
- 회사 자산이 C로 되어 있고 단기적으로 못 옮김

## 6.10 🧪 실습 코드: SHA-256 계산을 OpenSSL로

`hash.go`:

```go
package main

/*
#cgo pkg-config: openssl
#include <openssl/sha.h>
#include <string.h>

void compute_sha256(const unsigned char *data, size_t len, unsigned char *digest) {
    SHA256(data, len, digest);
}
*/
import "C"

import (
    "encoding/hex"
    "fmt"
    "unsafe"
)

func sha256OpenSSL(data []byte) string {
    digest := make([]byte, 32)  // SHA-256은 32바이트
    if len(data) == 0 {
        C.compute_sha256(nil, 0, (*C.uchar)(unsafe.Pointer(&digest[0])))
    } else {
        C.compute_sha256(
            (*C.uchar)(unsafe.Pointer(&data[0])),
            C.size_t(len(data)),
            (*C.uchar)(unsafe.Pointer(&digest[0])),
        )
    }
    return hex.EncodeToString(digest)
}

func main() {
    hash := sha256OpenSSL([]byte("Hello, CGo!"))
    fmt.Println("SHA-256:", hash)
}
```

빌드 & 실행:
```bash
# Ubuntu 기준 OpenSSL 설치
sudo apt install libssl-dev pkg-config

go run hash.go
# SHA-256: ...
```

물론 Go 표준 라이브러리 `crypto/sha256`이 더 단순합니다. 이 예제는 **CGo 메커니즘 이해용**입니다.

### ✅ 6교시 체크포인트

- [ ] CGo 기본 문법(주석 안 C 코드, `import "C"`)을 이해했는가?
- [ ] 외부 라이브러리를 `#cgo LDFLAGS`로 링크할 수 있는가?
- [ ] Go ↔ C 문자열/슬라이스 변환을 수행할 수 있는가?
- [ ] CGo의 오버헤드와 트레이드오프를 설명할 수 있는가?
- [ ] CGo 대신 순수 Go를 고려해야 할 상황을 판단할 수 있는가?

---

# 7교시. 종합 실습 I — REST API 서버 구축

5일간 배운 모든 것을 종합한 **사용자 관리 REST API 서버**를 만듭니다.

## 7.1 요구사항

- 엔드포인트: 사용자 CRUD
  - `POST   /users` — 생성
  - `GET    /users` — 목록 (페이지네이션)
  - `GET    /users/{id}` — 조회
  - `PUT    /users/{id}` — 수정
  - `DELETE /users/{id}` — 삭제
- 인메모리 저장소 (동시성 안전)
- JSON 요청/응답
- 구조화 로깅 (`slog`)
- Graceful shutdown
- `/healthz`, `/debug/pprof/`
- 단위 테스트 + HTTP 핸들러 테스트

## 7.2 Step 1 — 프로젝트 셋업

```bash
mkdir -p ~/go-class/day5/userapi
cd ~/go-class/day5/userapi
go mod init userapi

mkdir -p cmd/userapi
mkdir -p internal/user
mkdir -p internal/httpserver
```

최종 구조:
```
userapi/
├── go.mod
├── Makefile
├── cmd/userapi/main.go
└── internal/
    ├── user/
    │   ├── user.go
    │   ├── store.go
    │   └── store_test.go
    └── httpserver/
        ├── handlers.go
        ├── handlers_test.go
        └── middleware.go
```

## 7.3 Step 2 — 도메인 모델

`internal/user/user.go`:

```go
package user

import (
    "errors"
    "time"
)

var (
    ErrNotFound     = errors.New("사용자를 찾을 수 없음")
    ErrAlreadyExists = errors.New("이미 존재함")
    ErrInvalid      = errors.New("잘못된 입력")
)

type User struct {
    ID        int       `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

func (u *User) Validate() error {
    if u.Name == "" {
        return errors.New("name 필수")
    }
    if u.Email == "" {
        return errors.New("email 필수")
    }
    return nil
}
```

## 7.4 Step 3 — 저장소 (인터페이스 + 인메모리 구현)

`internal/user/store.go`:

```go
package user

import (
    "sort"
    "sync"
    "time"
)

// Store는 사용자 저장소 추상화. 다른 구현(DB 등)으로 교체 가능
type Store interface {
    Create(u *User) error
    Get(id int) (*User, error)
    List(offset, limit int) ([]*User, int, error)
    Update(u *User) error
    Delete(id int) error
}

// 인메모리 구현
type MemoryStore struct {
    mu     sync.RWMutex
    users  map[int]*User
    nextID int
}

func NewMemoryStore() *MemoryStore {
    return &MemoryStore{
        users:  make(map[int]*User),
        nextID: 1,
    }
}

func (s *MemoryStore) Create(u *User) error {
    if err := u.Validate(); err != nil {
        return err
    }
    s.mu.Lock()
    defer s.mu.Unlock()

    // 이메일 중복 체크
    for _, existing := range s.users {
        if existing.Email == u.Email {
            return ErrAlreadyExists
        }
    }

    u.ID = s.nextID
    s.nextID++
    now := time.Now()
    u.CreatedAt = now
    u.UpdatedAt = now
    s.users[u.ID] = u
    return nil
}

func (s *MemoryStore) Get(id int) (*User, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    u, ok := s.users[id]
    if !ok {
        return nil, ErrNotFound
    }
    // 복사본 반환 - 외부 수정 방지
    cp := *u
    return &cp, nil
}

func (s *MemoryStore) List(offset, limit int) ([]*User, int, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()

    total := len(s.users)
    all := make([]*User, 0, total)
    for _, u := range s.users {
        cp := *u
        all = append(all, &cp)
    }
    // ID 기준 정렬 (간단하게)
    sort.Slice(all, func(i, j int) bool { return all[i].ID < all[j].ID })

    if offset > total {
        return []*User{}, total, nil
    }
    end := offset + limit
    if end > total {
        end = total
    }
    return all[offset:end], total, nil
}

func (s *MemoryStore) Update(u *User) error {
    if err := u.Validate(); err != nil {
        return err
    }
    s.mu.Lock()
    defer s.mu.Unlock()

    existing, ok := s.users[u.ID]
    if !ok {
        return ErrNotFound
    }
    existing.Name = u.Name
    existing.Email = u.Email
    existing.UpdatedAt = time.Now()
    return nil
}

func (s *MemoryStore) Delete(id int) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    if _, ok := s.users[id]; !ok {
        return ErrNotFound
    }
    delete(s.users, id)
    return nil
}
```

> ✅ **점검 노트**: `List`에서 `sort.Slice`를 쓰므로 import에 `"sort"`가 반드시 필요합니다(위 코드에 반영됨). 누락 시 `undefined: sort` 컴파일 에러.

## 7.5 Step 4 — 저장소 테스트

`internal/user/store_test.go`:

```go
package user

import (
    "errors"
    "fmt"
    "testing"
)

func TestMemoryStore_CRUD(t *testing.T) {
    s := NewMemoryStore()

    // Create
    u := &User{Name: "Alice", Email: "alice@example.com"}
    if err := s.Create(u); err != nil {
        t.Fatal(err)
    }
    if u.ID == 0 {
        t.Error("ID 미할당")
    }

    // Get
    got, err := s.Get(u.ID)
    if err != nil {
        t.Fatal(err)
    }
    if got.Name != "Alice" {
        t.Errorf("name = %s", got.Name)
    }

    // Update
    got.Name = "Alice Updated"
    if err := s.Update(got); err != nil {
        t.Fatal(err)
    }
    got2, _ := s.Get(u.ID)
    if got2.Name != "Alice Updated" {
        t.Errorf("update 안 됨")
    }

    // Delete
    if err := s.Delete(u.ID); err != nil {
        t.Fatal(err)
    }
    _, err = s.Get(u.ID)
    if !errors.Is(err, ErrNotFound) {
        t.Errorf("삭제 후에도 조회됨")
    }
}

func TestMemoryStore_Duplicate(t *testing.T) {
    s := NewMemoryStore()
    s.Create(&User{Name: "A", Email: "a@a.com"})
    err := s.Create(&User{Name: "B", Email: "a@a.com"})
    if !errors.Is(err, ErrAlreadyExists) {
        t.Errorf("중복 허용됨")
    }
}

func TestMemoryStore_Concurrent(t *testing.T) {
    s := NewMemoryStore()
    const N = 100

    done := make(chan struct{})
    for i := 0; i < N; i++ {
        go func(idx int) {
            u := &User{
                Name:  "User",
                Email: fmt.Sprintf("user%d@example.com", idx),
            }
            s.Create(u)
            done <- struct{}{}
        }(i)
    }
    for i := 0; i < N; i++ {
        <-done
    }

    _, total, _ := s.List(0, 1000)
    if total != N {
        t.Errorf("동시 생성 개수 불일치: got %d, want %d", total, N)
    }
}
```

> ✅ **점검 노트**: `TestMemoryStore_Concurrent`에서 `fmt.Sprintf`를 쓰므로 `"fmt"` import가 필요합니다(위 코드에 반영됨).

테스트 실행 (검증 결과: `ok userapi/internal/user`):
```bash
go test -race ./internal/user
```

`-race`로 race 없음 확인.

## 7.6 Step 5 — HTTP 핸들러

`internal/httpserver/handlers.go`:

```go
package httpserver

import (
    "encoding/json"
    "errors"
    "log/slog"
    "net/http"
    "strconv"

    "userapi/internal/user"
)

type Server struct {
    Store user.Store
}

func New(store user.Store) *Server {
    return &Server{Store: store}
}

func (s *Server) Routes() http.Handler {
    mux := http.NewServeMux()
    mux.HandleFunc("GET /healthz", s.healthz)
    mux.HandleFunc("POST /users", s.createUser)
    mux.HandleFunc("GET /users", s.listUsers)
    mux.HandleFunc("GET /users/{id}", s.getUser)
    mux.HandleFunc("PUT /users/{id}", s.updateUser)
    mux.HandleFunc("DELETE /users/{id}", s.deleteUser)
    return mux
}

func (s *Server) healthz(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("ok"))
}

// 유틸 - JSON 응답
func writeJSON(w http.ResponseWriter, status int, body any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(body)
}

// 유틸 - 에러 응답
func writeError(w http.ResponseWriter, status int, msg string) {
    writeJSON(w, status, map[string]string{"error": msg})
}

func (s *Server) createUser(w http.ResponseWriter, r *http.Request) {
    var u user.User
    if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
        writeError(w, http.StatusBadRequest, "잘못된 JSON")
        return
    }
    defer r.Body.Close()

    if err := s.Store.Create(&u); err != nil {
        switch {
        case errors.Is(err, user.ErrAlreadyExists):
            writeError(w, http.StatusConflict, err.Error())
        default:
            writeError(w, http.StatusBadRequest, err.Error())
        }
        return
    }

    slog.Info("user created", "user_id", u.ID, "name", u.Name)
    writeJSON(w, http.StatusCreated, u)
}

func (s *Server) listUsers(w http.ResponseWriter, r *http.Request) {
    offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
    limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
    if limit == 0 {
        limit = 20
    }

    users, total, err := s.Store.List(offset, limit)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }

    writeJSON(w, http.StatusOK, map[string]any{
        "users":  users,
        "total":  total,
        "offset": offset,
        "limit":  limit,
    })
}

func (s *Server) getUser(w http.ResponseWriter, r *http.Request) {
    id, err := strconv.Atoi(r.PathValue("id"))
    if err != nil {
        writeError(w, http.StatusBadRequest, "잘못된 id")
        return
    }

    u, err := s.Store.Get(id)
    if err != nil {
        if errors.Is(err, user.ErrNotFound) {
            writeError(w, http.StatusNotFound, err.Error())
        } else {
            writeError(w, http.StatusInternalServerError, err.Error())
        }
        return
    }
    writeJSON(w, http.StatusOK, u)
}

func (s *Server) updateUser(w http.ResponseWriter, r *http.Request) {
    id, err := strconv.Atoi(r.PathValue("id"))
    if err != nil {
        writeError(w, http.StatusBadRequest, "잘못된 id")
        return
    }

    var u user.User
    if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
        writeError(w, http.StatusBadRequest, "잘못된 JSON")
        return
    }
    u.ID = id

    if err := s.Store.Update(&u); err != nil {
        if errors.Is(err, user.ErrNotFound) {
            writeError(w, http.StatusNotFound, err.Error())
        } else {
            writeError(w, http.StatusBadRequest, err.Error())
        }
        return
    }

    slog.Info("user updated", "user_id", id)
    writeJSON(w, http.StatusOK, &u)
}

func (s *Server) deleteUser(w http.ResponseWriter, r *http.Request) {
    id, err := strconv.Atoi(r.PathValue("id"))
    if err != nil {
        writeError(w, http.StatusBadRequest, "잘못된 id")
        return
    }
    if err := s.Store.Delete(id); err != nil {
        if errors.Is(err, user.ErrNotFound) {
            writeError(w, http.StatusNotFound, err.Error())
        } else {
            writeError(w, http.StatusInternalServerError, err.Error())
        }
        return
    }
    slog.Info("user deleted", "user_id", id)
    w.WriteHeader(http.StatusNoContent)
}
```

## 7.7 Step 6 — 미들웨어

`internal/httpserver/middleware.go`:

```go
package httpserver

import (
    "log/slog"
    "net/http"
    "time"
)

type statusRecorder struct {
    http.ResponseWriter
    status int
}

func (sr *statusRecorder) WriteHeader(code int) {
    sr.status = code
    sr.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        sr := &statusRecorder{ResponseWriter: w, status: 200}
        next.ServeHTTP(sr, r)
        slog.Info("http",
            "method", r.Method,
            "path", r.URL.Path,
            "status", sr.status,
            "duration_ms", time.Since(start).Milliseconds(),
        )
    })
}

func RecoveryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if rec := recover(); rec != nil {
                slog.Error("패닉", "panic", rec)
                http.Error(w, "internal server error", http.StatusInternalServerError)
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```

7교시는 여기까지. 8교시에서 `main.go`와 종합 테스트로 마무리합니다.

### ✅ 7교시 체크포인트

- [ ] 도메인/저장소/HTTP 핸들러를 패키지로 분리할 수 있는가?
- [ ] `Store` 인터페이스로 의존성을 추상화할 수 있는가?
- [ ] Go 1.22 라우팅 + `PathValue`를 활용할 수 있는가?
- [ ] 미들웨어로 로깅과 panic 복구를 구현할 수 있는가?

---

# 8교시. 종합 실습 II — 서버 완성 + 마무리

## 8.1 Step 7 — `main.go` (서버 부트스트랩)

`cmd/userapi/main.go`:

```go
package main

import (
    "context"
    "errors"
    "log/slog"
    "net/http"
    _ "net/http/pprof"
    "os"
    "os/signal"
    "syscall"
    "time"

    "userapi/internal/httpserver"
    "userapi/internal/user"
)

var (
    Version   = "dev"
    BuildTime = "unknown"
)

func main() {
    // 로거 설정 - JSON 형식
    opts := &slog.HandlerOptions{Level: slog.LevelInfo}
    slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, opts)))

    slog.Info("starting userapi", "version", Version, "built", BuildTime)

    // 저장소 + 서버 조립
    store := user.NewMemoryStore()
    server := httpserver.New(store)

    // 미들웨어 체이닝
    handler := httpserver.LoggingMiddleware(
        httpserver.RecoveryMiddleware(
            server.Routes(),
        ),
    )

    srv := &http.Server{
        Addr:         ":8080",
        Handler:      handler,
        ReadTimeout:  10 * time.Second,
        WriteTimeout: 10 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    // pprof 별도 포트
    go func() {
        slog.Info("pprof 시작", "addr", ":6060")
        http.ListenAndServe(":6060", nil)
    }()

    // 본 서버 시작
    go func() {
        slog.Info("서버 시작", "addr", srv.Addr)
        if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
            slog.Error("ListenAndServe", "err", err)
            os.Exit(1)
        }
    }()

    // graceful shutdown
    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
    <-sigs

    slog.Info("종료 신호 수신")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        slog.Error("Shutdown 에러", "err", err)
    }
    slog.Info("정상 종료")
}
```

## 8.2 Step 8 — HTTP 핸들러 테스트

`internal/httpserver/handlers_test.go`:

```go
package httpserver

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"

    "userapi/internal/user"
)

func newTestServer() *Server {
    return New(user.NewMemoryStore())
}

func TestCreateUser(t *testing.T) {
    srv := newTestServer()
    handler := srv.Routes()

    body := strings.NewReader(`{"name":"Alice","email":"alice@example.com"}`)
    req := httptest.NewRequest("POST", "/users", body)
    w := httptest.NewRecorder()

    handler.ServeHTTP(w, req)

    if w.Code != http.StatusCreated {
        t.Fatalf("status = %d", w.Code)
    }

    var got user.User
    if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
        t.Fatal(err)
    }
    if got.Name != "Alice" {
        t.Errorf("name = %s", got.Name)
    }
    if got.ID == 0 {
        t.Error("ID 미할당")
    }
}

func TestGetUser_NotFound(t *testing.T) {
    srv := newTestServer()
    handler := srv.Routes()

    req := httptest.NewRequest("GET", "/users/999", nil)
    w := httptest.NewRecorder()

    handler.ServeHTTP(w, req)

    if w.Code != http.StatusNotFound {
        t.Errorf("status = %d, want 404", w.Code)
    }
}

func TestUserLifecycle(t *testing.T) {
    srv := newTestServer()
    handler := srv.Routes()

    // 1. 생성
    body := strings.NewReader(`{"name":"Bob","email":"bob@example.com"}`)
    req := httptest.NewRequest("POST", "/users", body)
    w := httptest.NewRecorder()
    handler.ServeHTTP(w, req)
    if w.Code != http.StatusCreated {
        t.Fatalf("create status = %d", w.Code)
    }

    var created user.User
    json.NewDecoder(w.Body).Decode(&created)

    // 2. 조회
    req = httptest.NewRequest("GET", "/users/1", nil)
    w = httptest.NewRecorder()
    handler.ServeHTTP(w, req)
    if w.Code != http.StatusOK {
        t.Errorf("get status = %d", w.Code)
    }

    // 3. 수정
    body = strings.NewReader(`{"name":"Bob Updated","email":"bob@example.com"}`)
    req = httptest.NewRequest("PUT", "/users/1", body)
    w = httptest.NewRecorder()
    handler.ServeHTTP(w, req)
    if w.Code != http.StatusOK {
        t.Errorf("update status = %d", w.Code)
    }

    // 4. 삭제
    req = httptest.NewRequest("DELETE", "/users/1", nil)
    w = httptest.NewRecorder()
    handler.ServeHTTP(w, req)
    if w.Code != http.StatusNoContent {
        t.Errorf("delete status = %d", w.Code)
    }

    // 5. 삭제 후 조회 - 404
    req = httptest.NewRequest("GET", "/users/1", nil)
    w = httptest.NewRecorder()
    handler.ServeHTTP(w, req)
    if w.Code != http.StatusNotFound {
        t.Errorf("after delete get status = %d", w.Code)
    }
}

// 벤치마크 - 생성 throughput
// ⚠️ 점검 노트: 같은 이메일을 반복하면 이메일 중복 체크에 걸려
// 두 번째 반복부터 409 Conflict 경로만 측정하게 됨 (실측 확인).
// 반복마다 고유 이메일을 만들어야 "생성" 성능을 측정할 수 있다.
func BenchmarkCreateUser(b *testing.B) {
    srv := newTestServer()
    handler := srv.Routes()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        payload := fmt.Sprintf(`{"name":"X","email":"x%d@x.com"}`, i)
        body := bytes.NewReader([]byte(payload))
        req := httptest.NewRequest("POST", "/users", body)
        w := httptest.NewRecorder()
        handler.ServeHTTP(w, req)
    }
}
```

테스트 실행:
```bash
go test -race -v ./...
go test -bench=. ./internal/httpserver
```

## 8.3 Step 9 — Makefile

```makefile
BINARY := userapi
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -ldflags="-s -w -X 'main.Version=$(VERSION)' -X 'main.BuildTime=$(BUILD_TIME)'"

.PHONY: all build test fmt vet run clean cover bench

all: fmt vet test build

build:
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/$(BINARY)

test:
	go test -race -v ./...

cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "→ coverage.html 열어보세요"

bench:
	go test -bench=. -benchmem ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

run: build
	./bin/$(BINARY)

clean:
	rm -rf bin/ coverage.*
```

## 8.4 Step 10 — 실행 및 종합 테스트

```bash
make all
make run
```

다른 터미널에서 API 테스트:

```bash
# 헬스 체크
curl http://localhost:8080/healthz

# 사용자 생성
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com"}'

# 목록
curl http://localhost:8080/users

# 조회
curl http://localhost:8080/users/1

# 수정
curl -X PUT http://localhost:8080/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice Updated","email":"alice@example.com"}'

# 삭제
curl -X DELETE http://localhost:8080/users/1

# pprof 확인
curl http://localhost:6060/debug/pprof/

# 부하 테스트 (간단)
for i in {1..100}; do
  curl -s -X POST http://localhost:8080/users \
    -H "Content-Type: application/json" \
    -d "{\"name\":\"User$i\",\"email\":\"u$i@x.com\"}" > /dev/null
done

curl http://localhost:8080/users?limit=5
```

## 8.5 Step 11 — 도커 이미지 (보너스)

`Dockerfile`:

```dockerfile
# 빌드 스테이지
FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/userapi ./cmd/userapi

# 실행 스테이지 - 정적 바이너리 + 빈 이미지
FROM scratch
COPY --from=builder /out/userapi /userapi
EXPOSE 8080 6060
ENTRYPOINT ["/userapi"]
```

빌드:
```bash
docker build -t userapi:latest .
docker run -p 8080:8080 -p 6060:6060 userapi:latest
```

**`FROM scratch`**가 가능한 이유는 Go가 정적 링크 바이너리를 만들기 때문입니다. 이미지 크기 ~10MB.

## 8.6 5일차 마무리 — 무엇을 더 배워야 하나?

5일 동안 Go 언어의 **80%**를 다뤘습니다. 실무에서 추가로 만나게 될 주제들:

### 데이터베이스
- `database/sql` 표준 인터페이스
- `pgx` (PostgreSQL), `go-sql-driver/mysql` (MySQL)
- 마이그레이션: `golang-migrate/migrate`
- ORM: `gorm`, `ent`, `sqlc` (대부분 ORM 회피 권장)

### 설정 관리
- 환경 변수: `os.Getenv`, `kelseyhightower/envconfig`
- 설정 파일: YAML(`yaml.v3`), TOML
- `flag` 패키지 (CLI 옵션)

### 인증/인가
- JWT: `golang-jwt/jwt`
- OAuth2: `golang.org/x/oauth2`
- 세션 관리

### 분산 추적
- OpenTelemetry: 표준 분산 추적/메트릭
- Prometheus: 메트릭 수집

### gRPC
- 마이크로서비스 간 통신
- `google.golang.org/grpc`

### 메시징
- Kafka, NATS, RabbitMQ 클라이언트

### 인기 프레임워크 (필요할 때)
- Gin, Echo, Fiber — HTTP 프레임워크
- 다만 표준 라이브러리만으로 시작하는 게 권장됨

## 8.7 🎯 최종 도전 과제 (선택)

이 5일 과정을 졸업하는 도전 과제. 시간이 될 때 천천히:

1. **DB 연동** — 인메모리 저장소를 PostgreSQL로 교체 (`pgx` 사용)
2. **인증** — JWT 토큰 기반 인증 미들웨어 추가
3. **OpenAPI** — Swagger UI로 API 문서화
4. **메트릭** — Prometheus 메트릭 노출 (`/metrics`)
5. **CI/CD** — GitHub Actions로 테스트 + 빌드 + 도커 이미지 푸시
6. **부하 테스트** — `wrk` 또는 `vegeta`로 부하 측정 후 pprof로 병목 분석

---

# 🎓 5일 과정 전체 마무리

## 5일간 다룬 것

| 일차 | 주제 |
|---|---|
| **1일차** | Go 기초 — 문법, 자료구조, 인터페이스, 첫 CLI |
| **2일차** | 패키지, 모듈, 빌드 — 본격적인 Go 프로젝트 |
| **3일차** | 동시성 — 고루틴, 채널, sync |
| **4일차** | 동시성 심화 — Context, Worker Pool, Pipeline |
| **5일차** | 실전 — HTTP, JSON, 테스트, 로깅, 프로파일링, CGo, REST API |

## C 개발자로서 Go를 배운 의미

지금까지 5일간 본 Go의 강점을 종합해봅시다.

| 영역 | C의 어려움 | Go의 해결 |
|---|---|---|
| 메모리 관리 | malloc/free 누수, double free | GC로 안전 |
| 의존성 관리 | pkg-config, LD_LIBRARY_PATH | go mod, 단일 바이너리 |
| 빌드 시스템 | Makefile, autotools 미로 | go build 한 줄 |
| 동시성 | pthread, mutex 지옥 | 고루틴, 채널, ctx |
| 에러 처리 | errno, return -1 산만 | 명시적 error 반환 |
| 테스팅 | 외부 프레임워크 필요 | testing 표준 |
| HTTP/JSON | libcurl, cJSON, libmicrohttpd | 표준 라이브러리 |
| 크로스 컴파일 | 별도 툴체인 | GOOS=linux GOARCH=arm64 |
| 프로파일링 | gprof, valgrind 등 별도 | pprof 표준 |

**Go가 C를 대체하는 언어**라고 보긴 어렵습니다. 시스템 프로그래밍, 임베디드, 커널 영역은 여전히 C/C++/Rust의 영역입니다. **Go는 그 위 계층의 서비스/도구/네트워크 소프트웨어**를 빠르고 안정적으로 만드는 데 특화되어 있습니다.

C 개발자에게 Go는 **"확장 도구"**입니다. 모든 걸 Go로 바꾸지 마시고, **적재적소에 활용**하세요. 그리고 정말 필요하면 CGo로 두 세계를 연결할 수 있습니다.

## 추가 학습 로드맵

### 단기 (1~3개월)
1. 사내 작은 도구를 Go로 만들어보기 (CLI, 자동화 스크립트)
2. 5일차 REST API 예제를 DB 연동까지 확장
3. 표준 라이브러리 문서 훑어보기: `pkg.go.dev/std`

### 중기 (3~6개월)
1. 실제 운영되는 마이크로서비스를 Go로 구축
2. Kubernetes operator나 클라우드 네이티브 도구 분석
3. 『Go 100 Mistakes』 책 정독

### 장기 (6개월+)
1. 오픈소스 Go 프로젝트 기여
2. 고성능 시스템(>10K RPS) 최적화 경험
3. Go 컴파일러/런타임 내부 탐구

## 핵심 격언 모음

5일간 등장한 격언들로 마무리합니다.

> "Less is exponentially more." — **Rob Pike** (Go의 단순함 철학)

> "Don't communicate by sharing memory; share memory by communicating." — **Rob Pike** (채널 우선 동시성)

> "Don't just check errors, handle them gracefully." — **Dave Cheney**

> "Clear is better than clever." — **Go Proverbs**

> "A little copying is better than a little dependency." — **Go Proverbs**

> "The bigger the interface, the weaker the abstraction." — **Rob Pike**

> "Concurrency is not parallelism." — **Rob Pike**

이 격언들은 단순한 문구가 아니라, **5일간 우리가 코드로 체험한 원칙**입니다.

---

## 📚 마지막 참고 자료

### 공식 문서
- [Go 공식 사이트](https://go.dev/) — 시작점
- [pkg.go.dev](https://pkg.go.dev/) — 표준 라이브러리 + 외부 패키지
- [Go 블로그](https://go.dev/blog/) — 새 기능 / 모범 사례

### 책 (강력 추천)
- **『The Go Programming Language』** (Donovan & Kernighan) — 한국어판 있음. C 출신자 필독
- **『Go 100 Mistakes and How to Avoid Them』** (Teiva Harsanyi) — 실무 함정 모음
- **『Concurrency in Go』** (Katherine Cox-Buday) — 동시성 깊이

### 영상/강연
- [Rob Pike - Go Concurrency Patterns](https://www.youtube.com/watch?v=f6kdp27TYZs)
- [Bryan Cantrill - Is It Time to Rewrite the OS in Rust?](https://www.youtube.com/watch?v=HgtRAbE1nBM) (Go와 Rust 비교 관점)

### 코드 읽기
- [Standard Library Source](https://github.com/golang/go/tree/master/src)
- [Awesome Go](https://github.com/avelino/awesome-go) — 큐레이션된 패키지 목록

### 커뮤니티
- [r/golang](https://reddit.com/r/golang)
- [Gophers Slack](https://gophers.slack.com/)
- 한국 Go 커뮤니티: GDG, GoKR

---

**축하합니다 — 5일 과정을 모두 마쳤습니다.** 🎉

이제 여러분은 Go 언어로 실무 프로젝트를 시작할 준비가 되었습니다. 코드는 자주 쓰고, 자주 읽고, 자주 리팩토링하면서 늘어갑니다. 한국에서도 Go를 쓰는 회사가 빠르게 늘고 있으니, **사내 또는 사이드 프로젝트**에서 작게라도 적용해보세요.

질문이 생기면 표준 라이브러리 소스를 직접 읽는 습관을 들이시길 권합니다. Go 표준 라이브러리는 **Go 코드의 모범 예시**이자 가장 좋은 교과서입니다.

> *"Programs must be written for people to read, and only incidentally for machines to execute."* — Harold Abelson
