# Go 언어 프로그래밍 2일차 — 에러 처리, 모듈, 패키지, 빌드 시스템

> **대상**: 1일차를 마친 C 개발자
> **목표**: 실무에서 Go 프로젝트를 운영하는 데 필요한 **에러 처리 관용구**, **모듈/패키지 시스템**, **빌드 도구**를 익혀, 멀티 패키지 프로젝트를 처음부터 끝까지 구성할 수 있다.
> **준비물**: 1일차 환경(`go`, 에디터, 터미널), Git(권장)

---

## 📋 2일차 시간표

| 교시 | 주제 | 핵심 내용 |
|---|---|---|
| 1교시 | Go 에러 처리 패턴 I | `error` 인터페이스, `errno` 비교 |
| 2교시 | Go 에러 처리 패턴 II | 에러 래핑, `panic`/`recover` |
| 3교시 | Go 모듈 시스템 I | `go mod`, C 라이브러리 관리 비교 |
| 4교시 | Go 모듈 시스템 II | 의존성, `go.sum`, 버전 전략 |
| 5교시 | 패키지 설계 및 관리 | 가시성, 헤더 파일 비교 |
| 6교시 | Go 빌드 시스템 I | `go build`, `go install`, 크로스 컴파일 |
| 7교시 | Go 빌드 시스템 II | Makefile 활용 |
| 8교시 | 실습 | 멀티 패키지 프로젝트 구성 |

---

# 1교시. Go 에러 처리 패턴 I — `error` 인터페이스

## 1.1 왜 에러 처리가 중요한가?

1일차에서 Go의 다중 반환값을 잠깐 봤습니다. 사실 **Go 코드의 30~50%는 에러 처리 코드**입니다.

```go
data, err := os.ReadFile("config.yaml")
if err != nil {
    return err
}
```

이 패턴이 끝없이 반복됩니다. 처음엔 "너무 장황하다"고 느낄 수 있지만, Go 설계자들은 **에러를 숨기지 않고 명시적으로 다루는 것**이 안전한 시스템 프로그래밍의 핵심이라고 봅니다.

## 1.2 C의 에러 처리 — `errno`의 한계

C에서 에러를 다루는 전통적인 방식을 떠올려봅시다.

```c
#include <stdio.h>
#include <errno.h>
#include <string.h>

int main(void) {
    FILE *fp = fopen("none.txt", "r");
    if (fp == NULL) {
        printf("에러: %s (errno=%d)\n", strerror(errno), errno);
        return 1;
    }
    // ...
}
```

C 에러 처리의 문제점은 명확합니다.

| 문제 | 설명 |
|---|---|
| **전역 변수 의존** | `errno`는 스레드 로컬이지만 본질적으로 전역. 함수 호출 사이에 덮어쓰여질 수 있음 |
| **반환값과 분리** | 어떤 함수는 -1, 어떤 함수는 NULL, 어떤 함수는 양수 → 일관성 없음 |
| **놓치기 쉽다** | `fopen`의 반환값 체크 안 해도 컴파일러가 경고하지 않음 |
| **타입 정보 없음** | `errno`는 그저 정수. 컨텍스트 정보 없음 |
| **에러 추적 불가** | 어디서 발생했는지 추적 어려움 |

## 1.3 Go의 `error` 인터페이스

Go에서 에러는 그냥 **인터페이스**입니다. 내장 정의는 이렇게 한 줄로 끝납니다.

```go
type error interface {
    Error() string
}
```

`Error() string` 메서드를 가진 모든 타입이 에러로 쓰일 수 있습니다. 1일차에서 배운 Duck Typing이 여기에 활용됩니다.

### 가장 기본적인 사용 — `errors.New`, `fmt.Errorf`

```go
package main

import (
    "errors"
    "fmt"
)

func sqrt(x float64) (float64, error) {
    if x < 0 {
        return 0, errors.New("음수의 제곱근을 계산할 수 없습니다")
    }
    // 실제 계산...
    return 0, nil
}

func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, fmt.Errorf("0으로 나눌 수 없습니다 (a=%v)", a)
    }
    return a / b, nil
}

func main() {
    if _, err := sqrt(-4); err != nil {
        fmt.Println("에러:", err)
    }

    if _, err := divide(10, 0); err != nil {
        fmt.Println("에러:", err)
    }
}
```

- `errors.New(string)`: 단순한 에러 생성
- `fmt.Errorf(format, args...)`: 포맷 문자열을 활용한 에러 생성 (printf 스타일)

### `nil`은 "에러 없음"

```go
return 결과값, nil  // 성공
return 0,    에러   // 실패
```

C에서 `errno == 0`이 정상이듯, Go에서는 `err == nil`이 정상입니다.

## 1.4 관용구(idiom): `if err != nil`

```go
result, err := someFunction()
if err != nil {
    // 에러 처리
    return err  // 또는 적절한 대응
}
// 정상 흐름 계속
```

이 패턴을 **얼리 리턴(early return)**이라고도 부릅니다. 정상 경로를 들여쓰기 깊은 곳에 두지 않는 것이 Go 스타일입니다.

### ❌ 안 좋은 예 — C 스타일 중첩

```go
result, err := doSomething()
if err == nil {
    other, err2 := doOther()
    if err2 == nil {
        // 깊은 중첩 — 가독성 나쁨
    }
}
```

### ✅ 좋은 예 — 얼리 리턴

```go
result, err := doSomething()
if err != nil {
    return err
}

other, err := doOther()
if err != nil {
    return err
}

// 정상 경로는 들여쓰기 없이
```

## 1.5 사용자 정의 에러 타입

내장 함수만으로 부족할 때, 직접 에러 타입을 만들 수 있습니다.

```go
type ValidationError struct {
    Field   string
    Message string
}

// error 인터페이스 충족 — Error() string만 만들면 됨
func (e *ValidationError) Error() string {
    return fmt.Sprintf("필드 '%s' 검증 실패: %s", e.Field, e.Message)
}

func validateAge(age int) error {
    if age < 0 {
        return &ValidationError{Field: "age", Message: "음수일 수 없습니다"}
    }
    if age > 150 {
        return &ValidationError{Field: "age", Message: "150세를 넘을 수 없습니다"}
    }
    return nil
}
```

이러면 호출자가 **에러의 종류를 구분**할 수 있습니다. 자세한 방법은 2교시에서 다룹니다.

## 1.6 `errno`와 Go 에러의 비교 정리

| 측면 | C `errno` | Go `error` |
|---|---|---|
| 전달 방식 | 전역 변수 | 명시적 반환값 |
| 무시 가능 여부 | 쉽게 무시됨 | 무시하면 변수 미사용으로 경고 (linter) |
| 타입 정보 | 정수만 | 임의의 타입 |
| 컨텍스트 | 없음 | 메시지, 필드 등 자유롭게 |
| 스레드 안전성 | 스레드 로컬 의존 | 본질적으로 안전 (값 자체) |

### 🧪 실습 코드: `errors_basic.go`

```go
package main

import (
    "errors"
    "fmt"
)

type DivisionError struct {
    Dividend float64
    Divisor  float64
}

func (e *DivisionError) Error() string {
    return fmt.Sprintf("나눗셈 오류: %v / %v", e.Dividend, e.Divisor)
}

func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, &DivisionError{Dividend: a, Divisor: b}
    }
    return a / b, nil
}

func main() {
    // 케이스 1: errors.New
    err1 := errors.New("단순 에러")
    fmt.Println(err1)

    // 케이스 2: fmt.Errorf
    err2 := fmt.Errorf("값 %d는 허용 범위(%d~%d)를 벗어남", 200, 0, 100)
    fmt.Println(err2)

    // 케이스 3: 사용자 정의 에러
    _, err := divide(10, 0)
    if err != nil {
        fmt.Println(err)
    }

    // 케이스 4: 정상 동작
    result, err := divide(10, 2)
    if err != nil {
        fmt.Println("에러:", err)
        return
    }
    fmt.Println("결과:", result)
}
```

### ✅ 1교시 체크포인트

- [ ] `error`가 인터페이스라는 사실을 이해했는가?
- [ ] `if err != nil` 관용구를 자연스럽게 쓸 수 있는가?
- [ ] 사용자 정의 에러 타입을 만들 수 있는가?
- [ ] C `errno`의 한계를 설명할 수 있는가?

---

# 2교시. Go 에러 처리 패턴 II — 래핑, panic, recover

## 2.1 에러 래핑(Wrapping) — Go 1.13+

호출 체인을 따라 에러가 올라올 때, 컨텍스트를 누적하면 디버깅이 쉬워집니다. C에서는 보통 로그 메시지를 별도로 남기지만, Go는 **에러 자체에 컨텍스트를 추가**하는 메커니즘이 있습니다.

### `fmt.Errorf` + `%w` 동사

```go
import (
    "fmt"
    "os"
)

func readConfig(path string) error {
    _, err := os.Open(path)
    if err != nil {
        // %w로 원본 에러를 래핑
        return fmt.Errorf("설정 파일 열기 실패 (%s): %w", path, err)
    }
    return nil
}

func main() {
    err := readConfig("none.yaml")
    if err != nil {
        fmt.Println(err)
        // 출력:
        // 설정 파일 열기 실패 (none.yaml): open none.yaml: no such file or directory
    }
}
```

- `%w`: 에러를 래핑(wrap). 원본 에러 정보를 유지한 채 메시지를 추가
- `%v`: 단순 출력 (원본 에러 정보 손실)

차이를 비교해봅시다.

```go
return fmt.Errorf("X 실패: %v", err)  // 원본 에러 정보 손실 — 단순 문자열
return fmt.Errorf("X 실패: %w", err)  // 원본 에러 정보 유지 — 추후 분석 가능
```

## 2.2 `errors.Is` — 특정 에러인지 확인

래핑된 에러 사슬에서 **특정 에러**를 찾아 분기 처리합니다.

```go
import (
    "errors"
    "fmt"
    "os"
)

func main() {
    _, err := os.Open("/nonexistent/file")
    if err != nil {
        if errors.Is(err, os.ErrNotExist) {
            fmt.Println("파일이 없습니다")
        } else if errors.Is(err, os.ErrPermission) {
            fmt.Println("권한 없음")
        } else {
            fmt.Println("기타 에러:", err)
        }
    }
}
```

C 스타일이라면 `errno == ENOENT`와 비슷한 비교지만, **래핑되어도 동작**한다는 차이가 있습니다.

```go
// 깊이 래핑되어도 OK
err := fmt.Errorf("처리 중: %w",
       fmt.Errorf("로딩 중: %w",
       fmt.Errorf("열기: %w", os.ErrNotExist)))

errors.Is(err, os.ErrNotExist)  // true
```

## 2.3 `errors.As` — 특정 타입으로 추출

사용자 정의 에러 타입에서 **필드 값**을 꺼내려면 `errors.As`를 씁니다.

```go
type ValidationError struct {
    Field string
    Msg   string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("%s: %s", e.Field, e.Msg)
}

func main() {
    var err error = fmt.Errorf("처리 실패: %w",
                   &ValidationError{Field: "email", Msg: "형식 오류"})

    var ve *ValidationError
    if errors.As(err, &ve) {
        fmt.Println("문제 필드:", ve.Field)
        fmt.Println("이유:", ve.Msg)
    }
}
```

| 함수 | 용도 |
|---|---|
| `errors.Is(err, target)` | 특정 **값**(sentinel)과 같은지 비교 |
| `errors.As(err, &target)` | 특정 **타입**으로 형변환 (필드 접근용) |

## 2.4 Sentinel Error 패턴

표준 라이브러리에는 미리 정의된 에러 값들이 있습니다.

```go
import "io"

if errors.Is(err, io.EOF) {
    // 입력 끝
}
```

내 패키지에서도 만들 수 있습니다.

```go
package mypackage

import "errors"

var (
    ErrNotFound      = errors.New("찾을 수 없음")
    ErrAlreadyExists = errors.New("이미 존재함")
    ErrUnauthorized  = errors.New("인증 실패")
)
```

호출자는 이렇게 비교합니다.

```go
if errors.Is(err, mypackage.ErrNotFound) {
    // 404 페이지로
}
```

## 2.5 `panic`과 `recover` — 마지막 수단

지금까지 본 패턴은 모두 **예측 가능한 에러**용입니다. 정말로 비정상적인 상황(예: 배열 범위 초과, nil 포인터 역참조, 가정 위반)에서는 `panic`이 발생합니다.

```go
func main() {
    arr := []int{1, 2, 3}
    fmt.Println(arr[10])  // panic: runtime error: index out of range
}
```

### C의 `setjmp/longjmp`와 비교

```c
// C — 거의 안 씀, 매우 위험
#include <setjmp.h>
jmp_buf env;

void deep_function(void) {
    longjmp(env, 1);  // env로 점프
}

int main(void) {
    if (setjmp(env) == 0) {
        deep_function();
    } else {
        printf("점프 후 복귀\n");
    }
}
```

```go
// Go — 관리된 방식
func mayPanic() {
    panic("심각한 오류!")
}

func safeCall() {
    defer func() {
        if r := recover(); r != nil {
            fmt.Println("패닉 복구:", r)
        }
    }()
    mayPanic()
    fmt.Println("이 라인은 실행 안 됨")
}

func main() {
    safeCall()
    fmt.Println("프로그램 계속")
}
```

### 동작 원리

1. `panic("...")`이 호출되면 현재 함수 실행 중단
2. `defer`된 함수들이 역순으로 실행
3. 호출 스택을 거슬러 올라가며 panic 전파
4. 도중에 `recover()`를 만나면 panic 중단
5. `recover()`가 없으면 프로그램 종료 + 스택 트레이스 출력

### 🔥 중요: panic은 일반적인 에러 처리에 쓰지 말 것!

```go
// ❌ 잘못된 사용
func parseAge(s string) int {
    n, err := strconv.Atoi(s)
    if err != nil {
        panic(err)  // 너무 강력 — 호출자가 처리할 기회 박탈
    }
    return n
}

// ✅ 올바른 사용
func parseAge(s string) (int, error) {
    return strconv.Atoi(s)
}
```

**`panic` 사용 가이드라인**:
- 프로그램 초기화 중 치명적 오류 (예: `regexp.MustCompile`)
- 절대로 일어나서는 안 되는 가정 위반 (불변량 깨짐)
- 라이브러리 사용자 대상 API에서는 거의 쓰지 말 것

### `defer`의 활용

`recover`만 보면 단순해 보이지만, `defer`는 Go에서 매우 자주 쓰입니다.

```go
f, err := os.Open("data.txt")
if err != nil {
    return err
}
defer f.Close()  // 함수 종료 시 자동 호출

// 파일 사용...
```

C에서 `goto cleanup` 패턴이나 try-finally(C에는 없음) 흉내를 위해 매크로를 만들던 일을 `defer` 한 줄로 해결합니다.

**다중 defer는 LIFO(스택) 순서로 실행됩니다**:

```go
defer fmt.Println("1")
defer fmt.Println("2")
defer fmt.Println("3")
// 출력 순서: 3, 2, 1
```

### 🧪 실습 코드: `errors_advanced.go`

```go
package main

import (
    "errors"
    "fmt"
)

var ErrInsufficientFunds = errors.New("잔액 부족")

type WithdrawError struct {
    AccountID string
    Requested float64
    Available float64
}

func (e *WithdrawError) Error() string {
    return fmt.Sprintf("계좌 %s에서 출금 실패: 요청=%.2f, 잔액=%.2f",
        e.AccountID, e.Requested, e.Available)
}

// Unwrap을 구현하면 errors.Is가 sentinel 비교 가능
func (e *WithdrawError) Unwrap() error {
    return ErrInsufficientFunds
}

func withdraw(accountID string, amount float64, balance float64) error {
    if amount > balance {
        return &WithdrawError{
            AccountID: accountID,
            Requested: amount,
            Available: balance,
        }
    }
    return nil
}

func safeDivide(a, b int) (result int, err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("패닉 복구: %v", r)
        }
    }()
    return a / b, nil // b=0이면 panic 발생
}

func main() {
    // 1. 사용자 정의 에러 + 래핑
    err := withdraw("ACC-001", 5000, 1000)
    if err != nil {
        fmt.Println(err)

        // sentinel 검사
        if errors.Is(err, ErrInsufficientFunds) {
            fmt.Println("→ 잔액 부족으로 분류")
        }

        // 타입 추출
        var we *WithdrawError
        if errors.As(err, &we) {
            fmt.Printf("→ 부족액: %.2f\n", we.Requested-we.Available)
        }
    }

    // 2. panic / recover
    fmt.Println()
    if _, err := safeDivide(10, 0); err != nil {
        fmt.Println("safeDivide:", err)
    }
    if result, err := safeDivide(10, 2); err == nil {
        fmt.Println("safeDivide:", result)
    }
}
```

### ✅ 2교시 체크포인트

- [ ] `%w`로 에러 래핑을 할 수 있는가?
- [ ] `errors.Is`와 `errors.As`를 구분해서 쓸 수 있는가?
- [ ] `panic`/`recover`를 언제 써야 하는지 판단할 수 있는가?
- [ ] `defer`의 LIFO 동작을 이해했는가?

---

# 3교시. Go 모듈 시스템 I — `go mod`

## 3.1 C에서의 의존성 관리는 얼마나 고통스러운가?

C 프로젝트에서 외부 라이브러리(예: libcurl, OpenSSL)를 쓰려면 일반적으로 이런 작업이 필요합니다.

```bash
# 1) 시스템 패키지로 설치 (배포판마다 명령 다름)
sudo apt install libcurl4-openssl-dev   # Ubuntu
sudo yum install libcurl-devel          # CentOS

# 2) Makefile에 헤더 경로, 라이브러리 경로 설정
CFLAGS  += $(shell pkg-config --cflags libcurl)
LDFLAGS += $(shell pkg-config --libs libcurl)

# 3) 빌드 후 실행 시 .so 라이브러리 경로 문제
ldd ./myapp
export LD_LIBRARY_PATH=/usr/local/lib:$LD_LIBRARY_PATH
```

문제점:
- **버전 고정 어려움**: 시스템에 설치된 한 가지 버전만 사용 가능
- **재현성 부족**: 다른 사람 컴퓨터에서 같은 빌드를 보장하기 힘듦
- **의존성 추적 부재**: 어떤 버전의 무엇이 필요한지 코드에 적혀있지 않음
- **충돌**: 프로젝트 A는 OpenSSL 1.0, B는 1.1 필요한 경우 지옥

## 3.2 Go 모듈의 등장

Go는 1.11부터 **모듈(modules)** 시스템을 도입했습니다(2018년). 핵심 아이디어:

- **프로젝트 단위로 의존성 명시** (`go.mod`)
- **버전 잠금**으로 재현 가능한 빌드 (`go.sum`)
- **시스템 전역 설치 불필요** — 프로젝트별로 독립
- **자동 다운로드** — `go get`이 알아서 가져옴

### `go.mod` 파일의 구조

```text
module github.com/user/myproject

go 1.24

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/stretchr/testify v1.8.4
)
```

- **module**: 이 모듈의 경로 (보통 git 저장소 URL과 일치)
- **go**: 사용하는 Go 버전
- **require**: 의존하는 외부 모듈과 버전

C로 비유하면 `pkg-config --modversion`을 자동으로 기록해두는 셈입니다.

## 3.3 첫 모듈 만들기 — 따라하기

### Step 1. 빈 디렉터리에서 시작

```bash
mkdir -p ~/go-class/day2/hello-mod
cd ~/go-class/day2/hello-mod

# 모듈 초기화 — go.mod 파일 생성
go mod init github.com/myname/hellomod
```

> **모듈 이름은 보통 `github.com/계정/리포지토리` 형식**입니다. 비공개라도 미래에 공개 가능성을 고려해 이 컨벤션을 따릅니다.

확인:
```bash
cat go.mod
```

```text
module github.com/myname/hellomod

go 1.22
```

### Step 2. 외부 라이브러리 사용

`main.go`:
```go
package main

import (
    "fmt"
    "github.com/google/uuid"
)

func main() {
    id := uuid.New()
    fmt.Println("새 UUID:", id)
}
```

### Step 3. 의존성 다운로드

```bash
go mod tidy
```

이 명령은:
- 코드에서 `import`된 외부 패키지를 분석
- `go.mod`에 `require` 자동 추가
- 다운로드 후 캐시 저장 (`~/go/pkg/mod`)
- 사용하지 않는 의존성은 제거

`go.mod`를 다시 보면:

```text
module github.com/myname/hellomod

go 1.22

require github.com/google/uuid v1.6.0
```

### Step 4. 실행

```bash
go run main.go
# 새 UUID: 5d2f8a3c-...
```

C에서 라이브러리를 쓸 때처럼 `pkg-config`, `LD_LIBRARY_PATH`를 만질 필요가 전혀 없습니다.

## 3.4 모듈 캐시의 위치

다운로드된 라이브러리는 어디 있을까요?

```bash
ls ~/go/pkg/mod/github.com/google/
# uuid@v1.6.0/

go env GOMODCACHE
# /home/사용자/go/pkg/mod
```

**프로젝트별로 복사되지 않습니다.** 한 번 다운로드하면 모든 프로젝트가 공유합니다. C의 시스템 전역 설치와 비슷해 보이지만, **버전별로 디렉터리가 분리**되어 충돌이 없습니다.

```bash
ls ~/go/pkg/mod/github.com/google/uuid@*
# uuid@v1.3.0/  uuid@v1.6.0/  uuid@v1.7.0/
```

## 3.5 모듈 경로의 의미

`module github.com/myname/hellomod`는 단순한 이름이 아닙니다.

다른 프로젝트가 우리 모듈을 import한다고 가정해봅시다.

```go
import "github.com/myname/hellomod/utils"
```

Go 툴체인은 이렇게 동작합니다:
1. `github.com/myname/hellomod`를 보고 GitHub 저장소를 추론
2. `git clone https://github.com/myname/hellomod`
3. 그 안의 `utils` 하위 디렉터리에서 패키지 찾기

**즉, 모듈 경로 = 다운로드 가능한 위치**가 됩니다. 별도의 패키지 레지스트리(npm, pip 같은)가 필요 없습니다.

## 3.6 C 라이브러리 관리 vs Go 모듈 정리

| 항목 | C | Go |
|---|---|---|
| 설치 위치 | 시스템 전역 (`/usr/lib`) | 프로젝트별 명시 (`go.mod`) + 공유 캐시 |
| 버전 관리 | 시스템 패키지 하나 | 프로젝트마다 독립 |
| 설치 방식 | OS 패키지 매니저 | `go get`, `go mod tidy` |
| 헤더/라이브러리 분리 | 분리됨 (`.h` + `.so`) | 통합 (소스+바이너리 한 번에) |
| 의존성 명시 | Makefile/CMake에 수동 | `go.mod`에 자동 |
| 빌드 재현성 | 어려움 | `go.sum`이 보장 |
| 사설 저장소 | 별도 환경 변수 | `GOPRIVATE` 환경 변수 |

### ✅ 3교시 체크포인트

- [ ] `go mod init`으로 새 모듈을 만들 수 있는가?
- [ ] `go mod tidy`의 역할을 설명할 수 있는가?
- [ ] `go.mod` 파일의 각 항목 의미를 이해했는가?
- [ ] C의 의존성 관리와의 차이를 설명할 수 있는가?

---

# 4교시. Go 모듈 시스템 II — 의존성과 버전 관리

## 4.1 `go.sum` — 무결성 보증서

`go mod tidy` 후 디렉터리를 보면 `go.sum`이라는 파일도 생겨있습니다.

```bash
cat go.sum
```

```text
github.com/google/uuid v1.6.0 h1:NIvaJDMOsjHA8n1jAhLSgzrAzy1Hgr+hNrb57e+94F0=
github.com/google/uuid v1.6.0/go.mod h1:TIyPZe4MgqvfeYDBFedMoGGpEw/LqOeaOT+nhxU+yHo=
```

각 의존성의 **SHA-256 해시**가 저장됩니다. 누군가 라이브러리를 몰래 변조해도 빌드가 거부됩니다.

> **C에는 이런 게 없습니다.** 시스템 패키지 매니저가 서명 검증을 하긴 하지만, 프로젝트 자체에 잠금 기능은 없습니다.

**`go.sum`은 반드시 Git에 커밋하세요.** 협업자와 CI에서 동일한 빌드를 보장하는 핵심입니다.

## 4.2 의존성 추가하기

### 방법 ①: 코드에서 import 후 `go mod tidy`

```go
import "github.com/sirupsen/logrus"
```

```bash
go mod tidy  # 자동으로 추가
```

### 방법 ②: `go get` 직접 실행

```bash
go get github.com/sirupsen/logrus
go get github.com/sirupsen/logrus@v1.9.3   # 특정 버전
go get github.com/sirupsen/logrus@latest   # 최신 버전
```

### 의존성 제거

```bash
# 1) 코드에서 import 삭제
# 2) tidy 실행
go mod tidy
```

## 4.3 Semantic Versioning (SemVer)

Go 모듈은 **시맨틱 버저닝**을 따릅니다.

```
v MAJOR . MINOR . PATCH
v   1   .   2   .   3
```

| 부분 | 의미 |
|---|---|
| MAJOR | 호환성을 깨는 변경 |
| MINOR | 호환성 유지 + 기능 추가 |
| PATCH | 호환성 유지 + 버그 수정 |

### Go의 독특한 규칙: v2 이상은 모듈 경로에 명시

C 라이브러리는 보통 `liblib1.so.2`처럼 파일명에 버전을 넣지만, Go는 **모듈 경로** 자체에 포함시킵니다.

```text
// v1.x.x
import "github.com/foo/bar"

// v2.x.x — 경로에 /v2 붙임
import "github.com/foo/bar/v2"
```

이러면 한 프로그램이 같은 모듈의 v1과 v2를 **동시에** 쓸 수도 있습니다. C에서 OpenSSL 1.x와 3.x를 동시 사용하는 게 어려운 것과 대조적입니다.

## 4.4 버전 업그레이드 / 다운그레이드

```bash
# 패치 버전 업데이트 (v1.9.3 → v1.9.5)
go get -u=patch github.com/sirupsen/logrus

# 마이너 + 패치 업데이트 (호환성 유지 범위)
go get -u github.com/sirupsen/logrus

# 특정 버전으로 고정
go get github.com/sirupsen/logrus@v1.9.0

# 모든 의존성 마이너 업데이트
go get -u ./...
```

## 4.5 `replace` — 로컬 개발 / 포크 사용

라이브러리에 버그가 있어서 직접 수정해야 한다면, `replace` 지시어를 씁니다.

```text
// go.mod
module github.com/myname/myapp

require github.com/foo/bar v1.2.3

replace github.com/foo/bar => ../my-fork-of-bar
// 또는 다른 저장소로
replace github.com/foo/bar => github.com/myname/bar-fork v1.2.4
```

C에서는 이런 일을 하려면 라이브러리를 직접 빌드해서 LD_LIBRARY_PATH로 우선순위 조작을 하거나, 시스템 라이브러리를 덮어써야 했습니다.

## 4.6 `exclude` — 특정 버전 제외

```text
// go.mod
exclude github.com/foo/bar v1.4.0  // 알려진 버그 버전
```

## 4.7 `vendor` 디렉터리 — 오프라인 빌드

```bash
go mod vendor
```

`vendor/` 디렉터리에 모든 의존성 소스가 복사됩니다. 인터넷이 없는 환경(폐쇄망)에서 빌드할 때 유용합니다. C 프로젝트에서 `third_party/` 디렉터리에 소스를 함께 두던 관행과 같습니다.

```bash
# vendor가 있으면 자동으로 vendor 사용
go build

# 명시적으로
go build -mod=vendor
```

## 4.8 의존성 점검 명령어

```bash
# 의존성 트리 보기
go mod graph

# 누가 이 패키지를 끌어왔는가?
go mod why github.com/some/package

# 사용 가능한 버전 확인
go list -m -versions github.com/sirupsen/logrus

# 알려진 취약점 검사 (govulncheck 설치 필요)
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

## 4.9 실전 워크플로 — Step by Step

### 시나리오: HTTP 클라이언트 모듈 사용

```bash
# 1) 새 프로젝트
mkdir -p ~/go-class/day2/http-demo
cd ~/go-class/day2/http-demo
go mod init httpdemo
```

`main.go`:
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

    body, _ := io.ReadAll(resp.Body)
    fmt.Println("응답 길이:", len(body), "바이트")
    fmt.Println("상태 코드:", resp.StatusCode)
}
```

```bash
go run main.go
```

여기까지는 외부 라이브러리 없이 표준 라이브러리만 썼습니다. `go.mod`는 변경되지 않습니다.

이제 외부 로깅 라이브러리를 추가해봅시다.

```go
import (
    "github.com/sirupsen/logrus"
    // ...
)

func main() {
    log := logrus.New()
    log.SetFormatter(&logrus.JSONFormatter{})

    resp, err := http.Get("https://api.github.com")
    if err != nil {
        log.WithError(err).Fatal("요청 실패")
    }
    defer resp.Body.Close()

    log.WithFields(logrus.Fields{
        "status": resp.StatusCode,
    }).Info("응답 수신")
}
```

```bash
go mod tidy   # logrus 자동 추가
go run main.go
```

`go.mod` 확인:
```text
module httpdemo

go 1.22

require github.com/sirupsen/logrus v1.9.3

require golang.org/x/sys v0.0.0-... // indirect
```

`// indirect`는 직접 import하진 않았지만 의존성의 의존성으로 들어온 패키지입니다.

### ✅ 4교시 체크포인트

- [ ] `go.sum`의 역할을 설명할 수 있는가?
- [ ] `go get`으로 버전을 지정해 설치할 수 있는가?
- [ ] `replace` 지시어로 로컬 포크를 쓸 수 있는가?
- [ ] `go mod vendor`가 언제 유용한지 알고 있는가?

---

# 5교시. 패키지 설계 및 관리

## 5.1 패키지란?

C에서 코드를 구조화하려면 `.h`/`.c` 파일을 나누고 `#include`로 연결합니다. Go는 **패키지(package)**라는 명확한 단위를 제공합니다.

**한 디렉터리 = 한 패키지** — 이 규칙은 절대적입니다.

```
myapp/
├── go.mod
├── main.go              ← package main
└── greeter/
    ├── english.go       ← package greeter
    └── korean.go        ← package greeter
```

`greeter/` 디렉터리 안의 모든 `.go` 파일은 같은 `package greeter` 선언을 공유합니다.

## 5.2 가시성(Visibility) 규칙 — 대문자/소문자

C는 `static` 키워드로 파일 스코프 제한을 했습니다.

```c
// C
static int internal_helper(void) { ... }  // 같은 파일에서만 보임
int public_function(void) { ... }         // 전역
```

Go는 더 단순합니다. **이름의 첫 글자가 대문자면 외부 공개(exported), 소문자면 비공개(unexported)**.

```go
package greeter

// 외부에 노출됨 (대문자)
func Hello() string {
    return greeting() + ", world!"
}

// 패키지 내부 전용 (소문자)
func greeting() string {
    return "Hello"
}

// 구조체 필드도 같은 규칙
type Config struct {
    Name    string  // 공개
    secret  string  // 비공개
}
```

### 사용 예

```go
package main

import (
    "fmt"
    "github.com/myname/myapp/greeter"
)

func main() {
    fmt.Println(greeter.Hello())     // ✅ OK
    // fmt.Println(greeter.greeting()) // ❌ 컴파일 에러 (비공개)

    cfg := greeter.Config{
        Name: "Go",         // ✅ OK
        // secret: "abc",  // ❌ 컴파일 에러
    }
    _ = cfg
}
```

| 측면 | C | Go |
|---|---|---|
| 제어 방식 | `static`, `extern` 키워드 | 이름 대소문자 |
| 단위 | 파일 단위 | 패키지(디렉터리) 단위 |
| 헤더 분리 | 필요 (`.h`) | 불필요 |
| 명시성 | 코드에 키워드 | 식별자에 내장 |

## 5.3 좋은 패키지 이름

| 좋음 | 나쁨 | 이유 |
|---|---|---|
| `http` | `httputil_v2` | 짧고 명확 |
| `json` | `json_parser` | 동사 / 군더더기 제거 |
| `user` | `userlib` | "lib", "pkg" 접미사 불필요 |

> **컨벤션**: 패키지 이름은 **짧은 소문자 한 단어**. 언더바, 카멜케이스 피함.

## 5.4 `internal/` — 강제 비공개

`internal/` 이라는 특수 디렉터리는 **부모 모듈 외부에서 import할 수 없습니다**.

```
github.com/myname/myapp/
├── go.mod
├── main.go
├── internal/
│   └── auth/
│       └── token.go   ← 외부 모듈에서 import 불가
└── pkg/
    └── api/
        └── client.go  ← 외부에 공개되는 API
```

다른 모듈이 `import "github.com/myname/myapp/internal/auth"`를 시도하면 컴파일 거부됩니다. C에는 이런 메커니즘이 없어서 헤더를 공개해 두면 누구나 가져다 쓸 수 있었지만, Go는 **공개 API와 내부 구현을 디렉터리 구조로 분리**합니다.

## 5.5 init 함수 — 패키지 초기화

C에서 `__attribute__((constructor))`로 main 전 호출되는 함수를 만들 수 있는데, Go에서는 `init()`이 그 역할입니다.

```go
package config

var settings map[string]string

func init() {
    settings = make(map[string]string)
    settings["env"] = "development"
}
```

- 패키지가 import될 때 **자동 호출**
- 한 패키지에 여러 `init()` 가능 (파일별로 따로 둘 수 있음)
- 인자도 반환값도 없음

> **남용 주의**: `init()`은 디버깅이 어렵습니다. 가능하면 명시적인 초기화 함수(예: `NewClient()`)를 쓰는 편이 권장됩니다.

## 5.6 추천 디렉터리 구조

작은 프로젝트:
```
myapp/
├── go.mod
├── main.go
└── README.md
```

중간 규모:
```
myapp/
├── go.mod
├── cmd/
│   └── myapp/
│       └── main.go      ← 실행 진입점
├── internal/
│   ├── config/
│   └── server/
└── pkg/                 ← 외부에서 import 가능한 라이브러리
    └── client/
```

- `cmd/`: 실행 가능한 프로그램들 (여러 개 가능)
- `internal/`: 이 프로젝트 전용
- `pkg/`: 다른 프로젝트가 가져다 쓸 수 있는 부분

> **주의**: 이건 강제 규칙이 아니라 **커뮤니티 컨벤션**입니다. 작은 프로젝트에 무리해서 적용할 필요는 없습니다.

## 5.7 🧪 실습 — 다중 패키지 미니 프로젝트

### Step 1. 구조 만들기

```bash
mkdir -p ~/go-class/day2/multipkg
cd ~/go-class/day2/multipkg
go mod init github.com/myname/multipkg

mkdir -p mathx
mkdir -p stringx
```

### Step 2. `mathx/mathx.go`

```go
package mathx

// 두 수의 최대공약수 (외부 공개)
func GCD(a, b int) int {
    a, b = abs(a), abs(b)
    for b != 0 {
        a, b = b, a%b
    }
    return a
}

// 비공개 헬퍼
func abs(n int) int {
    if n < 0 {
        return -n
    }
    return n
}
```

### Step 3. `stringx/stringx.go`

```go
package stringx

// 문자열 뒤집기
func Reverse(s string) string {
    r := []rune(s)
    for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
        r[i], r[j] = r[j], r[i]
    }
    return string(r)
}
```

### Step 4. `main.go`

```go
package main

import (
    "fmt"
    "github.com/myname/multipkg/mathx"
    "github.com/myname/multipkg/stringx"
)

func main() {
    fmt.Println("GCD(24, 36) =", mathx.GCD(24, 36))
    fmt.Println("Reverse('안녕 Go') =", stringx.Reverse("안녕 Go"))
}
```

### Step 5. 실행

```bash
go run main.go
# GCD(24, 36) = 12
# Reverse('안녕 Go') = oG 녕안
```

### Step 6. 비공개 함수 접근 시도 (실패 확인)

```go
// main.go에 추가
fmt.Println(mathx.abs(-5))
```

```bash
go run main.go
# ./main.go:11:21: cannot refer to unexported name mathx.abs
```

→ 컴파일 거부됨을 확인. C의 `static` 함수와 같은 효과지만 **이름만 보고도 구분**됩니다.

### ✅ 5교시 체크포인트

- [ ] 한 디렉터리 = 한 패키지 규칙을 이해했는가?
- [ ] 이름 첫 글자 대소문자로 공개/비공개를 제어할 수 있는가?
- [ ] `internal/` 디렉터리의 효과를 설명할 수 있는가?
- [ ] 다른 패키지의 함수를 import해 사용할 수 있는가?

---

# 6교시. Go 빌드 시스템 I — `go build`, `go install`

## 6.1 C의 컴파일 과정 복습

C 프로그램 하나를 빌드하려면 보통 4단계를 거칩니다.

```text
[main.c] → preprocess → [main.i]
                            ↓
                         compile
                            ↓
                       [main.s]   (어셈블리)
                            ↓
                         assemble
                            ↓
                       [main.o]   (오브젝트)
                            ↓
                         link  ← stdlib.o, libcurl.so, ...
                            ↓
                       [main]    (실행 파일)
```

대규모 프로젝트에서는 Makefile, CMake 등 빌드 시스템이 이 과정을 관리합니다. 각 `.o` 파일을 일일이 만들고 링크하는 것은 분명히 빠르지 않습니다.

## 6.2 Go의 컴파일은 왜 빠른가?

Go의 빌드 시간이 C/C++에 비해 매우 빠른 이유:

1. **헤더 파일 없음** — 텍스트 인클루드를 재처리할 필요 없음
2. **의존성 그래프 단순** — 패키지 단위로 한 번만 컴파일, 결과 캐싱
3. **단일 패스 컴파일러** — 전방 선언이 필요 없는 설계
4. **링커 빌트인** — 외부 ld 호출 없음

Google 사내에서 "C++로 1시간 걸리던 빌드가 Go로 1분"이라는 일화가 자주 언급됩니다.

## 6.3 `go build`

```bash
cd ~/go-class/day2/multipkg
go build
```

기본 동작:
- 현재 디렉터리의 `main` 패키지를 빌드
- 디렉터리 이름과 같은 실행 파일 생성 (`multipkg`)
- 빌드된 모든 의존 패키지는 `$GOCACHE`에 캐싱

```bash
ls
# go.mod  go.sum  main.go  multipkg*  mathx/  stringx/

./multipkg
# GCD(24, 36) = 12
# ...
```

### 출력 파일 이름 지정

```bash
go build -o myapp
# ./myapp
```

### 다른 디렉터리 빌드

```bash
go build ./cmd/server      # 특정 패키지
go build ./...             # 모든 패키지
```

### 상세 출력

```bash
go build -v ./...   # 컴파일되는 패키지 목록 표시
go build -x ./...   # 내부에서 실행되는 명령어까지 표시
```

## 6.4 `go install`

```bash
go install
```

- 빌드 결과를 **`$GOPATH/bin`**에 설치 (`~/go/bin`)
- PATH에 `$GOPATH/bin`이 있으면 어디서든 실행 가능

```bash
which multipkg
# /home/사용자/go/bin/multipkg
```

C로 치면 `make && sudo make install`을 한 번의 명령으로 끝낸 셈입니다. 단, **시스템 디렉터리가 아닌 사용자 디렉터리**에 설치되어 sudo가 필요 없습니다.

### 원격에서 직접 설치

```bash
go install github.com/some/tool@latest
```

GitHub 등의 저장소에서 도구를 받아 컴파일하고 `$GOPATH/bin`에 설치합니다. C의 `apt install` 같은 효과지만, **소스 빌드를 거치므로 플랫폼에 맞춰 최적화**됩니다.

## 6.5 빌드 플래그 — 자주 쓰는 것들

### 정적 링크 / 동적 링크 제어

```bash
# 완전 정적 링크 (CGO 비활성화)
CGO_ENABLED=0 go build -o myapp

# 결과 확인
ldd ./myapp
# not a dynamic executable    ← 완전 독립 실행 파일
```

Linux Docker 이미지에서 `FROM scratch`(빈 베이스)로 시작할 수 있는 이유입니다.

### 디버그 정보 제거 — 바이너리 크기 축소

```bash
go build -ldflags="-s -w" -o myapp
# -s: 심볼 테이블 제거
# -w: DWARF 디버그 정보 제거
```

크기가 30~40% 줄어듭니다. 다만 stack trace 시 함수명만 남고 라인 정보는 사라집니다.

### 버전 정보 주입

```go
// main.go
package main

var (
    Version   = "dev"
    BuildTime = "unknown"
)

func main() {
    fmt.Printf("v%s built at %s\n", Version, BuildTime)
}
```

```bash
go build -ldflags="-X 'main.Version=1.0.0' -X 'main.BuildTime=$(date)'" -o myapp
./myapp
# v1.0.0 built at Mon Jan 15 14:23:00 UTC 2024
```

C에서 `-D VERSION=...` 매크로로 하던 일이 더 깔끔하게 됩니다.

## 6.6 크로스 컴파일 — 다른 OS/아키텍처 빌드

이 부분은 Go의 압도적 장점입니다.

```bash
# Linux ARM64
GOOS=linux GOARCH=arm64 go build -o myapp-linux-arm64

# Windows x64
GOOS=windows GOARCH=amd64 go build -o myapp.exe

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o myapp-mac

# Raspberry Pi (ARM7)
GOOS=linux GOARCH=arm GOARM=7 go build -o myapp-rpi
```

**툴체인 추가 설치 없이** 한 줄로 됩니다. C에서 ARM 크로스 컴파일 환경을 세팅하던 고통과 비교해보세요.

지원되는 조합 목록:
```bash
go tool dist list
# aix/ppc64
# android/386
# darwin/amd64
# darwin/arm64
# linux/amd64
# linux/arm64
# windows/amd64
# ...
```

## 6.7 빌드 캐시

```bash
go env GOCACHE
# /home/사용자/.cache/go-build
```

빌드 결과가 여기에 캐싱되어, 변경되지 않은 패키지는 재컴파일되지 않습니다.

```bash
go clean -cache     # 캐시 모두 삭제
go build           # 다시 전체 빌드 → 첫 빌드는 느림
go build           # 두 번째 빌드 → 거의 즉시 (캐시 적중)
```

### ✅ 6교시 체크포인트

- [ ] `go build`와 `go install`의 차이를 설명할 수 있는가?
- [ ] 크로스 컴파일을 환경 변수로 지정할 수 있는가?
- [ ] `-ldflags`로 버전 정보를 주입할 수 있는가?
- [ ] 정적 링크 바이너리를 만들 수 있는가?

---

# 7교시. Go 빌드 시스템 II — Makefile 활용

## 7.1 왜 Makefile이 또 필요한가?

Go는 `go build` 한 줄로 충분한데, 왜 Makefile을 또 만들까요?

이유는 **반복 작업의 자동화** 때문입니다.

- 빌드 + 테스트 + 린터 + 포맷 한 번에
- 환경 변수 세팅 자동화
- 여러 플랫폼용 바이너리 일괄 생성
- 도커 이미지 빌드와 연계
- CI/CD에서 일관된 명령 진입점

## 7.2 C Makefile과의 차이

C Makefile은 보통 다음을 다룹니다.

```makefile
CC = gcc
CFLAGS = -Wall -O2
LDFLAGS = -lm

myapp: main.o util.o
	$(CC) $(LDFLAGS) -o myapp main.o util.o

main.o: main.c util.h
	$(CC) $(CFLAGS) -c main.c

util.o: util.c util.h
	$(CC) $(CFLAGS) -c util.c

clean:
	rm -f *.o myapp
```

Go에서는 **소스 파일 간 의존성을 Make에 적을 필요가 없습니다.** Go 컴파일러가 알아서 합니다. 그래서 Go Makefile은 훨씬 단순합니다.

## 7.3 표준 Go Makefile 템플릿

```makefile
# === 변수 ===
BINARY      := myapp
VERSION     := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME  := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS     := -ldflags="-s -w -X 'main.Version=$(VERSION)' -X 'main.BuildTime=$(BUILD_TIME)'"

# === 기본 타깃 ===
.PHONY: all
all: fmt vet test build

# === 빌드 ===
.PHONY: build
build:
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/$(BINARY)

# === 크로스 컴파일 ===
.PHONY: build-all
build-all: build-linux build-mac build-windows

.PHONY: build-linux
build-linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY)-linux-amd64 ./cmd/$(BINARY)

.PHONY: build-mac
build-mac:
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY)-darwin-arm64 ./cmd/$(BINARY)

.PHONY: build-windows
build-windows:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY)-windows-amd64.exe ./cmd/$(BINARY)

# === 테스트 ===
.PHONY: test
test:
	go test -v -race -cover ./...

.PHONY: bench
bench:
	go test -bench=. -benchmem ./...

# === 코드 품질 ===
.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: lint
lint:
	@which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run ./...

# === 의존성 ===
.PHONY: tidy
tidy:
	go mod tidy

.PHONY: vendor
vendor:
	go mod vendor

# === 실행 / 설치 ===
.PHONY: run
run:
	go run ./cmd/$(BINARY)

.PHONY: install
install:
	go install $(LDFLAGS) ./cmd/$(BINARY)

# === 정리 ===
.PHONY: clean
clean:
	rm -rf bin/
	go clean -cache -testcache

# === 도움말 ===
.PHONY: help
help:
	@echo "사용 가능한 타깃:"
	@echo "  make build       - 바이너리 빌드"
	@echo "  make build-all   - 모든 플랫폼용 빌드"
	@echo "  make test        - 테스트 실행"
	@echo "  make fmt         - 코드 포맷"
	@echo "  make vet         - 정적 분석"
	@echo "  make lint        - 린터 실행"
	@echo "  make clean       - 빌드 산출물 정리"
```

## 7.4 핵심 포인트

### `.PHONY`의 중요성

```makefile
.PHONY: build
build:
	go build ...
```

`.PHONY`로 선언하면 **같은 이름의 파일이 있어도 항상 실행**됩니다. C Makefile에서도 `make clean`이 동작하려면 `clean`을 PHONY로 지정하던 것과 같은 이유입니다.

### `@` 접두사로 명령어 출력 숨기기

```makefile
fmt:
	@echo "포맷팅 중..."
	go fmt ./...
```

`@`가 붙은 줄은 명령어 자체가 출력되지 않습니다. 깔끔한 로그를 위해 유용합니다.

### 변수 평가 시점

```makefile
VERSION := $(shell git describe --tags)   # := 즉시 평가 (한 번만)
VERSION  = $(shell git describe --tags)   #  = 매번 평가 (느림)
```

Go Makefile에서는 보통 `:=`를 씁니다.

## 7.5 자주 쓰는 패턴들

### 도커 이미지 빌드 통합

```makefile
DOCKER_IMAGE := myname/myapp
DOCKER_TAG   := $(VERSION)

.PHONY: docker
docker:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_IMAGE):latest

.PHONY: docker-push
docker-push: docker
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
	docker push $(DOCKER_IMAGE):latest
```

### 테스트 커버리지 보고서

```makefile
.PHONY: cover
cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "커버리지 보고서: coverage.html"
```

### 개발용 핫 리로드 (외부 도구 사용)

```makefile
.PHONY: dev
dev:
	@which air > /dev/null || go install github.com/cosmtrek/air@latest
	air
```

## 7.6 C Makefile과의 차이 정리

| 항목 | C Makefile | Go Makefile |
|---|---|---|
| 의존성 추적 | 일일이 명시 (.h 파일 등) | Go가 알아서 처리 |
| 부분 빌드 | `.o` 단위 캐시 수동 | `go build`가 자동 캐싱 |
| 빌드 단계 | 다단계 (cpp → cc → as → ld) | 단일 명령 |
| 주 용도 | 빌드 자체 | 빌드 + 테스트 + 배포 자동화 |
| 복잡도 | 높음 | 낮음 |

### ✅ 7교시 체크포인트

- [ ] Go Makefile의 일반적인 타깃들을 이해했는가?
- [ ] `.PHONY`의 역할을 설명할 수 있는가?
- [ ] `make build-all`로 멀티 플랫폼 빌드를 할 수 있는가?
- [ ] C Makefile과의 단순함 차이를 체감했는가?

---

# 8교시. 실습 — 멀티 패키지 프로젝트 구성

지금까지 배운 모든 것을 종합해, **간단한 계산기 CLI**를 멀티 패키지 + Makefile로 만듭니다.

## 8.1 요구사항

- 명령: `calc add 3 5`, `calc mul 4 7`, `calc div 10 3`
- 지원 연산: add, sub, mul, div
- 에러 처리: 0으로 나누기, 잘못된 인자 개수, 숫자 파싱 실패
- 멀티 패키지: `cmd/calc`, `internal/calculator`, `pkg/parser`
- Makefile로 빌드 / 테스트 / 설치

## 8.2 디렉터리 구조 만들기

```bash
mkdir -p ~/go-class/day2/calc
cd ~/go-class/day2/calc

go mod init github.com/myname/calc

mkdir -p cmd/calc
mkdir -p internal/calculator
mkdir -p pkg/parser
```

최종 구조:
```
calc/
├── go.mod
├── Makefile
├── cmd/
│   └── calc/
│       └── main.go
├── internal/
│   └── calculator/
│       ├── calculator.go
│       └── calculator_test.go
└── pkg/
    └── parser/
        ├── parser.go
        └── parser_test.go
```

## 8.3 Step 1 — `pkg/parser/parser.go`

명령줄 인자를 파싱하는 작은 패키지.

```go
package parser

import (
    "fmt"
    "strconv"
)

// Args는 파싱된 명령 인자를 담는다
type Args struct {
    Op string
    A  float64
    B  float64
}

// Parse는 ["add", "3", "5"] 같은 슬라이스를 받아 Args로 변환한다
func Parse(args []string) (*Args, error) {
    if len(args) != 3 {
        return nil, fmt.Errorf("인자 3개가 필요합니다 (받은 개수: %d)", len(args))
    }

    a, err := strconv.ParseFloat(args[1], 64)
    if err != nil {
        return nil, fmt.Errorf("첫 번째 피연산자 파싱 실패 (%q): %w", args[1], err)
    }

    b, err := strconv.ParseFloat(args[2], 64)
    if err != nil {
        return nil, fmt.Errorf("두 번째 피연산자 파싱 실패 (%q): %w", args[2], err)
    }

    return &Args{Op: args[0], A: a, B: b}, nil
}
```

## 8.4 Step 2 — `pkg/parser/parser_test.go`

```go
package parser

import "testing"

func TestParse_Valid(t *testing.T) {
    args, err := Parse([]string{"add", "3", "5"})
    if err != nil {
        t.Fatalf("예상치 못한 에러: %v", err)
    }
    if args.Op != "add" || args.A != 3 || args.B != 5 {
        t.Errorf("파싱 결과 불일치: %+v", args)
    }
}

func TestParse_WrongCount(t *testing.T) {
    _, err := Parse([]string{"add", "3"})
    if err == nil {
        t.Error("에러가 발생해야 함")
    }
}

func TestParse_InvalidNumber(t *testing.T) {
    _, err := Parse([]string{"add", "abc", "5"})
    if err == nil {
        t.Error("에러가 발생해야 함")
    }
}
```

테스트 실행:
```bash
go test ./pkg/parser
# ok  github.com/myname/calc/pkg/parser  0.002s
```

> `_test.go` 파일은 Go에서 **자동 인식**됩니다. 별도 빌드 설정 없이 `go test`로 실행됩니다.

## 8.5 Step 3 — `internal/calculator/calculator.go`

핵심 연산 로직. **`internal/`에 두어** 외부 모듈이 못 쓰게 합니다.

```go
package calculator

import (
    "errors"
    "fmt"
)

// ErrDivByZero는 0으로 나눌 때 반환된다
var ErrDivByZero = errors.New("0으로 나눌 수 없습니다")

// Calculate는 연산자와 두 피연산자를 받아 결과를 반환한다
func Calculate(op string, a, b float64) (float64, error) {
    switch op {
    case "add":
        return a + b, nil
    case "sub":
        return a - b, nil
    case "mul":
        return a * b, nil
    case "div":
        if b == 0 {
            return 0, fmt.Errorf("a=%v, b=%v: %w", a, b, ErrDivByZero)
        }
        return a / b, nil
    default:
        return 0, fmt.Errorf("지원하지 않는 연산: %q (지원: add, sub, mul, div)", op)
    }
}
```

## 8.6 Step 4 — `internal/calculator/calculator_test.go`

```go
package calculator

import (
    "errors"
    "testing"
)

func TestCalculate_BasicOps(t *testing.T) {
    tests := []struct {
        name     string
        op       string
        a, b     float64
        expected float64
    }{
        {"add", "add", 3, 5, 8},
        {"sub", "sub", 10, 4, 6},
        {"mul", "mul", 6, 7, 42},
        {"div", "div", 10, 2, 5},
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            got, err := Calculate(tc.op, tc.a, tc.b)
            if err != nil {
                t.Fatalf("예상치 못한 에러: %v", err)
            }
            if got != tc.expected {
                t.Errorf("기대값 %v, 받은 값 %v", tc.expected, got)
            }
        })
    }
}

func TestCalculate_DivByZero(t *testing.T) {
    _, err := Calculate("div", 10, 0)
    if !errors.Is(err, ErrDivByZero) {
        t.Errorf("ErrDivByZero를 기대했으나 %v를 받음", err)
    }
}

func TestCalculate_UnknownOp(t *testing.T) {
    _, err := Calculate("pow", 2, 3)
    if err == nil {
        t.Error("에러가 발생해야 함")
    }
}
```

> 이 코드의 `t.Run`은 **테이블 드리븐 테스트** 패턴입니다. Go 커뮤니티의 표준 관용구이니 익혀두세요.

## 8.7 Step 5 — `cmd/calc/main.go`

진입점. 두 패키지를 import해서 조합합니다.

```go
package main

import (
    "errors"
    "fmt"
    "os"

    "github.com/myname/calc/internal/calculator"
    "github.com/myname/calc/pkg/parser"
)

var Version = "dev"

func main() {
    if len(os.Args) < 2 {
        usage()
        os.Exit(1)
    }

    // --version 처리
    if os.Args[1] == "--version" || os.Args[1] == "-v" {
        fmt.Printf("calc %s\n", Version)
        return
    }

    args, err := parser.Parse(os.Args[1:])
    if err != nil {
        fmt.Fprintln(os.Stderr, "에러:", err)
        os.Exit(1)
    }

    result, err := calculator.Calculate(args.Op, args.A, args.B)
    if err != nil {
        if errors.Is(err, calculator.ErrDivByZero) {
            fmt.Fprintln(os.Stderr, "❌ 0으로 나눌 수 없습니다")
        } else {
            fmt.Fprintln(os.Stderr, "에러:", err)
        }
        os.Exit(1)
    }

    fmt.Printf("%g %s %g = %g\n", args.A, symbol(args.Op), args.B, result)
}

func symbol(op string) string {
    switch op {
    case "add":
        return "+"
    case "sub":
        return "-"
    case "mul":
        return "×"
    case "div":
        return "÷"
    }
    return op
}

func usage() {
    fmt.Println("사용법: calc <op> <a> <b>")
    fmt.Println("  op: add | sub | mul | div")
    fmt.Println("예시: calc add 3 5")
    fmt.Println()
    fmt.Println("옵션:")
    fmt.Println("  -v, --version  버전 표시")
}
```

## 8.8 Step 6 — Makefile

프로젝트 루트에 `Makefile` 생성:

```makefile
BINARY     := calc
VERSION    := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS    := -ldflags="-s -w -X 'main.Version=$(VERSION)'"

.PHONY: all
all: fmt vet test build

.PHONY: build
build:
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/$(BINARY)

.PHONY: build-all
build-all:
	GOOS=linux   GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY)-linux-amd64 ./cmd/$(BINARY)
	GOOS=darwin  GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY)-darwin-arm64 ./cmd/$(BINARY)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY)-windows-amd64.exe ./cmd/$(BINARY)

.PHONY: test
test:
	go test -v -race -cover ./...

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: install
install:
	go install $(LDFLAGS) ./cmd/$(BINARY)

.PHONY: run
run:
	go run ./cmd/$(BINARY) add 3 5

.PHONY: clean
clean:
	rm -rf bin/

.PHONY: help
help:
	@echo "타깃: all build build-all test fmt vet tidy install run clean"
```

## 8.9 Step 7 — 빌드 및 테스트

```bash
# 1) 코드 포맷
make fmt

# 2) 정적 분석
make vet

# 3) 테스트 실행
make test
```

기대 출력:
```text
=== RUN   TestCalculate_BasicOps
=== RUN   TestCalculate_BasicOps/add
=== RUN   TestCalculate_BasicOps/sub
=== RUN   TestCalculate_BasicOps/mul
=== RUN   TestCalculate_BasicOps/div
--- PASS: TestCalculate_BasicOps (0.00s)
    --- PASS: TestCalculate_BasicOps/add
    ...
PASS
ok  github.com/myname/calc/internal/calculator  0.003s
```

```bash
# 4) 빌드
make build

ls bin/
# calc

# 5) 실행
./bin/calc add 3 5
# 3 + 5 = 8

./bin/calc div 10 3
# 10 ÷ 3 = 3.3333333333333335

./bin/calc div 10 0
# ❌ 0으로 나눌 수 없습니다

./bin/calc pow 2 3
# 에러: 지원하지 않는 연산: "pow" (지원: add, sub, mul, div)

./bin/calc --version
# calc dev
```

## 8.10 Step 8 — 멀티 플랫폼 빌드

```bash
make build-all

ls -la bin/
# -rwxr-xr-x  calc-darwin-arm64
# -rwxr-xr-x  calc-linux-amd64
# -rwxr-xr-x  calc-windows-amd64.exe
```

C에서 이런 일을 하려면 각 플랫폼별 크로스 컴파일러 설치, 표준 라이브러리 크로스 빌드, 환경 변수 설정 등이 필요했죠. Go는 한 명령으로 끝납니다.

## 8.11 Step 9 — 시스템에 설치

```bash
make install
# go install -ldflags="..." ./cmd/calc

# 어디서든 실행 가능
calc mul 4 7
# 4 × 7 = 28

which calc
# /home/사용자/go/bin/calc
```

## 8.12 디렉터리 최종 점검

```bash
tree
```

```text
.
├── Makefile
├── bin
│   ├── calc
│   ├── calc-darwin-arm64
│   ├── calc-linux-amd64
│   └── calc-windows-amd64.exe
├── cmd
│   └── calc
│       └── main.go
├── go.mod
├── internal
│   └── calculator
│       ├── calculator.go
│       └── calculator_test.go
└── pkg
    └── parser
        ├── parser.go
        └── parser_test.go
```

이제 진정한 "Go 프로젝트" 형태가 갖춰졌습니다.

### 🎯 도전 과제 (선택)

이 프로젝트를 발전시켜보세요.

1. **`mod` (나머지) 연산 추가** — `calc mod 10 3`이 `1` 출력
2. **계산 이력 저장** — 환경 변수 `CALC_HISTORY=on` 시 결과를 `~/.calc-history`에 기록
3. **외부 라이브러리 사용** — `github.com/spf13/cobra`로 CLI를 더 풍부하게 (서브커맨드, 도움말 자동 생성)
4. **벤치마크 테스트** — `BenchmarkCalculate`를 작성하고 `make bench` 타깃 추가

### ✅ 8교시 체크포인트

- [ ] 멀티 패키지 프로젝트의 구조를 직접 설계할 수 있는가?
- [ ] `internal/`과 `pkg/`의 차이를 이해했는가?
- [ ] 테이블 드리븐 테스트 패턴을 쓸 수 있는가?
- [ ] Makefile로 빌드 / 테스트 / 배포 자동화 흐름을 만들 수 있는가?

---

# 🎓 2일차 마무리

## 오늘 배운 것

1. **에러 처리 I**: `error` 인터페이스, `if err != nil` 관용구, 사용자 정의 에러
2. **에러 처리 II**: `%w` 래핑, `errors.Is/As`, `panic`/`recover`, `defer`
3. **모듈 시스템 I**: `go mod init`, `go mod tidy`, `go.mod` 구조
4. **모듈 시스템 II**: `go.sum`, SemVer, `replace`/`exclude`/`vendor`
5. **패키지 설계**: 대소문자 가시성, `internal/`, 디렉터리 구조
6. **빌드 시스템 I**: `go build`, `go install`, 크로스 컴파일, `-ldflags`
7. **빌드 시스템 II**: Go용 Makefile 패턴
8. **실전**: 계산기 CLI를 멀티 패키지로 구축, 테스트, 빌드 자동화

## 한 줄 요약

> **"Go의 진짜 강점은 언어 자체보다 도구 생태계에 있다."** — 오늘 배운 `go mod`, `go test`, `go build`는 C에서 수십 개의 외부 도구를 조합해야 했던 일을 통합 제공합니다.

## 복습 과제

다음 시간 전에 다음을 직접 해보세요.

1. **간단한 텍스트 처리 도구**를 만들어 보세요.
   - 패키지 구조: `cmd/textutil`, `internal/processor`, `pkg/fileio`
   - 기능: 파일 읽고 단어 개수 / 줄 개수 / 문자 개수 출력 (`wc` 흉내)
   - 에러 처리: 파일 없음, 권한 없음을 `errors.Is`로 구분
   - Makefile 포함

2. **사용자 정의 에러 타입** 3가지 이상 정의해보고 `errors.As`로 추출하는 예제 작성

3. **`go mod why`**, **`go mod graph`** 명령을 좋아하는 프로젝트에서 실행해 의존성 트리 분석

## 다음 시간 예고 — 3일차

드디어 Go의 **시그니처 기능, 동시성**에 들어갑니다.

- **런타임 스케줄러**: M:N 스케줄링, 고루틴 vs OS 스레드
- **고루틴 심화**: 스택 관리, Context Switching 비용
- **Channel**: Unbuffered, Buffered, 동기화 패턴
- **sync 패키지**: `Mutex`, `RWMutex`, `WaitGroup`, `Once`
- **실습**: Producer-Consumer 패턴, Mutex vs Channel 비교

---

## 📚 참고 자료

- [Go Modules Reference](https://go.dev/ref/mod) — `go mod` 공식 문서
- [Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors) — 에러 래핑 공식 블로그
- [Go Project Layout](https://github.com/golang-standards/project-layout) — 커뮤니티 표준 디렉터리 구조 (참고용)
- [Effective Go - Errors](https://go.dev/doc/effective_go#errors)
- [Dave Cheney - Don't just check errors, handle them gracefully](https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully) — 에러 처리 베스트 프랙티스
- 책: 『Go 100 Mistakes and How to Avoid Them』 — 에러 처리 / 모듈 관련 챕터 추천
