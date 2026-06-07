# Go 언어 프로그래밍 1일차 — C 개발자를 위한 Go 입문

> **대상**: C 언어 개발 경험이 있고 Linux 환경에서 프로그래밍이 가능한 수강생
> **목표**: Go 언어의 철학을 이해하고, 기본 문법·자료구조·인터페이스까지 다루어 C 코드를 Go로 옮길 수 있는 수준에 도달한다.
> **준비물**: Linux(또는 WSL) 환경, 텍스트 에디터(VS Code 권장), 터미널

---

## 📋 1일차 시간표

| 교시 | 주제 | 핵심 내용 |
|---|---|---|
| 1교시 | Go 언어 개요 및 개발환경 구축 | Go의 역사·철학, C와의 차이, 설치 |
| 2교시 | Go 기본 문법 I | 변수, 상수, 타입 시스템 |
| 3교시 | Go 기본 문법 II | 함수, 다중 반환값, 포인터 |
| 4교시 | Go 기본 문법 III | 구조체, 메서드 |
| 5교시 | Go 자료구조 | 배열, 슬라이스, 맵 |
| 6교시 | 인터페이스 기초 | Duck Typing |
| 7교시 | 실습 | C 코드를 Go로 변환 |
| 8교시 | 실습 | 간단한 CLI 프로그램 작성 |

---

# 1교시. Go 언어 개요 및 개발환경 구축

## 1.1 왜 Go인가? — Go의 탄생 배경

Go는 2007년 Google에서 **Robert Griesemer, Rob Pike, Ken Thompson** 세 사람이 설계를 시작했고, 2009년에 공개되었습니다. 주목할 점은 설계자들의 배경입니다.

- **Ken Thompson**: UNIX와 C 언어의 원조 설계자(B 언어 → C로 이어지는 계보)
- **Rob Pike**: UNIX 시스템 개발자, Plan 9 운영체제 핵심 멤버
- **Robert Griesemer**: V8 JavaScript 엔진, Java HotSpot 컴파일러 개발자

즉, **C와 시스템 프로그래밍의 DNA를 그대로 이어받은 언어**입니다. 그래서 C 개발자가 보면 어색하지 않은 부분이 많습니다.

### Google이 Go를 만든 이유

당시 Google 내부의 고민은 이랬습니다.

1. **C++ 빌드 시간이 너무 길다** — 대규모 코드베이스에서 빌드만 30분~1시간
2. **동시성 처리가 어렵다** — pthread, mutex의 복잡도가 너무 높음
3. **의존성 관리가 지옥** — `#include`, `Makefile`, 헤더/소스 분리의 피로감
4. **멀티코어 시대 대응 부족** — 1코어 시절 언어로 멀티코어를 다루는 한계

Go는 이 문제들을 **"단순하게"** 해결하는 데 집중했습니다.

## 1.2 Go의 설계 철학

> *"Less is exponentially more."* — Rob Pike

Go의 철학을 한 줄로 요약하면 **"의도적으로 단순하게"** 입니다. 다른 현대 언어들이 기능을 늘려갈 때, Go는 오히려 빼는 방향을 택했습니다.

### Go의 핵심 철학 5가지

1. **단순함(Simplicity)** — 키워드 25개. C는 32개, Java는 50+. 언어 명세서가 매우 얇음.
2. **명시성(Explicit)** — 암묵적 동작 최소화. 마법(magic) 없음.
3. **합성(Composition)** — 상속(Inheritance) 없음. 인터페이스와 임베딩으로 해결.
4. **동시성(Concurrency)** — 언어 차원에서 `goroutine`, `channel` 제공.
5. **빠른 빌드(Fast Build)** — 의존성 그래프를 단순화하여 컴파일 속도 극대화.

### "있어야 할 것이 없는" 언어

C 개발자 시각에서 보면 **놀랍게 빠진 기능들**이 있습니다.

| 일반적인 언어 기능 | Go에서는? |
|---|---|
| 클래스(Class) | 없음 — 구조체 + 메서드로 대체 |
| 상속(Inheritance) | 없음 — 임베딩(embedding)으로 대체 |
| 제네릭(과거) | 1.18 이전엔 없었음 (현재는 있음) |
| 예외(try/catch) | 없음 — `error` 반환값으로 대체 |
| 매크로(`#define`) | 없음 — 상수와 함수로 대체 |
| 헤더 파일 | 없음 — 패키지 단위 |
| while/do-while | 없음 — `for`만 있음 |
| 삼항 연산자 `?:` | 없음 — if-else로 작성 |
| 함수/메서드 오버로딩 | 없음 |
| 포인터 산술 | 없음 (CGO 제외) |

처음엔 "이게 없다고?" 싶지만, 실제로 코드를 써보면 **없어서 더 명확하다**는 걸 느끼게 됩니다.

## 1.3 Go vs C — 주요 차이점 개요

| 항목 | C | Go |
|---|---|---|
| 컴파일 결과 | 동적 링크 실행파일 (기본) | **정적 링크 단일 바이너리** |
| 메모리 관리 | 수동 (`malloc`/`free`) | **GC (Garbage Collector)** |
| 포인터 산술 | 가능 (`p++`) | **불가능** (안전성을 위해 금지) |
| 헤더 파일 | `.h` 분리 필요 | 없음 (패키지로 통합) |
| 빌드 도구 | `make`, `cmake` 등 외부 도구 | **`go build`** (언어 내장) |
| 의존성 관리 | 수동 / `pkg-config` 등 | **`go mod`** 내장 |
| 동시성 | `pthread` 라이브러리 | **`goroutine`, `channel` (언어 기본)** |
| 에러 처리 | 반환 코드 + `errno` | **명시적 `error` 반환값** |
| 문자열 | `char*` + null terminator | **UTF-8 기반 `string` 타입** |
| 배열 경계 검사 | 없음 (위험) | **있음** (런타임 panic) |
| Undefined Behavior | 매우 많음 | 거의 없음 (명세 명확) |

> **핵심 메시지**: Go는 C의 "직접 메모리/하드웨어 접근" 자유를 일부 포기하는 대신, **안전성과 생산성**을 얻은 언어입니다.

## 1.4 개발환경 구축 (Linux 기준)

이제 직접 손으로 Go를 설치해봅니다.

### Step 1. Go 설치

```bash
# 1) 최신 버전 다운로드 (1.24.x 기준)
cd /tmp
wget https://go.dev/dl/go1.24.0.linux-amd64.tar.gz

# 2) 기존 설치가 있다면 제거 후 설치
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.24.0.linux-amd64.tar.gz

# 3) PATH 등록 — ~/.bashrc에 추가
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
source ~/.bashrc
```

> **패키지 매니저로 설치하지 않는 이유**: `apt install golang`은 버전이 한참 뒤떨어진 경우가 많습니다. 공식 tarball 설치를 권장합니다.

### Step 2. 설치 확인

```bash
go version
# go version go1.24.0 linux/amd64

go env GOROOT GOPATH
# /usr/local/go
# /home/사용자명/go
```

- `GOROOT`: Go 자체가 설치된 위치 (`/usr/local/go`)
- `GOPATH`: 작성한 코드와 외부 패키지가 저장될 작업 공간 (`~/go`)

### Step 3. 첫 프로그램 작성 — Hello, World

```bash
# 작업 디렉터리 생성
mkdir -p ~/go-class/day1/hello
cd ~/go-class/day1/hello

# 모듈 초기화 — 이게 C의 Makefile + pkg-config 대체
go mod init hello
```

`hello.go` 파일을 만들고 다음을 입력:

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, Go!")
}
```

실행:

```bash
go run hello.go
# Hello, Go!

# 또는 빌드 후 실행
go build hello.go
./hello
```

### 🔍 C 코드와 비교해보기

같은 동작의 C 코드는 이렇습니다.

```c
#include <stdio.h>

int main(void) {
    printf("Hello, C!\n");
    return 0;
}
```

차이점을 짚어봅시다.

| 요소 | C | Go |
|---|---|---|
| 헤더 인클루드 | `#include <stdio.h>` | `import "fmt"` |
| 시작 함수 | `int main(void)` | `func main()` |
| 출력 함수 | `printf(...)` | `fmt.Println(...)` |
| 종료 코드 | `return 0;` | 명시 불필요 (필요시 `os.Exit`) |
| 파일 확장자 | `.c` / `.h` | `.go` 하나 |
| 빌드 | `gcc hello.c -o hello` | `go build hello.go` |
| 세미콜론 | 필요 | **자동 삽입** (작성 안 함) |

### Step 4. 핵심 명령어 정리

| 명령어 | 역할 |
|---|---|
| `go run <file.go>` | 빌드 + 실행 (임시 실행파일은 자동 삭제) |
| `go build <file.go>` | 실행 파일 생성 |
| `go fmt` | 코드 자동 정렬 (스타일 통일 강제) |
| `go vet` | 정적 분석 (C의 `lint` 대응) |
| `go mod init <name>` | 모듈 초기화 |
| `go mod tidy` | 의존성 자동 정리 |
| `go test` | 테스트 실행 |
| `go doc <pkg>` | 패키지 문서 보기 |

> **`gofmt` 철학**: Go는 코드 스타일을 **언어 차원에서 강제**합니다. 들여쓰기, 중괄호 위치, import 정렬을 개인 취향으로 두지 않습니다. 이는 협업 시 큰 장점입니다.

### ✅ 1교시 체크포인트

- [ ] `go version`이 정상 출력되는가?
- [ ] `hello.go`를 작성하고 실행했는가?
- [ ] C 컴파일 과정과 Go의 차이를 설명할 수 있는가?

---

# 2교시. Go 기본 문법 I — 변수, 상수, 타입 시스템

## 2.1 변수 선언 — 3가지 방식

C는 변수 선언이 한 가지 스타일이지만, Go는 상황에 따라 3가지를 제공합니다.

### 방식 ①: 전통적인 `var` 선언

```go
var age int = 30
var name string = "Alice"
```

### 방식 ②: 타입 생략 (타입 추론)

```go
var age = 30       // int로 추론
var pi = 3.14      // float64로 추론
var name = "Bob"   // string으로 추론
```

### 방식 ③: 짧은 선언 `:=` (가장 많이 씀)

```go
age := 30
name := "Charlie"
```

**`:=`는 함수 내부에서만 쓸 수 있습니다.** 전역 변수는 반드시 `var`로 선언합니다.

### 🔍 C와의 비교

```c
// C
int age = 30;
const double PI = 3.14;
char name[] = "Bob";
```

```go
// Go
age := 30
const PI = 3.14
name := "Bob"
```

**핵심 차이**:
- Go는 **타입을 변수명 뒤에** 적습니다 (`age int`).
  - 이유: 복잡한 타입 선언에서 왼쪽→오른쪽으로 자연스럽게 읽힘.
- 초기화하지 않은 변수도 **0값(zero value)으로 자동 초기화**됩니다. C의 쓰레기값 문제 없음.

### Zero Value 규칙

선언만 하고 초기화 안 해도 안전합니다.

```go
var n int      // 0
var f float64  // 0.0
var s string   // "" (빈 문자열, NULL 아님!)
var b bool     // false
var p *int     // nil
```

> **C와 가장 다른 부분**: C에서 `int x;`만 선언하면 스택 쓰레기값이 들어있어 `printf("%d", x)`가 위험합니다. Go는 절대 그런 일이 없습니다.

## 2.2 상수 — `const`

```go
const Pi = 3.14
const Greeting = "Hello"

// 그룹화
const (
    StatusOK       = 200
    StatusNotFound = 404
    StatusError    = 500
)
```

### iota — 자동 증가 상수 (C의 enum과 비슷)

```go
const (
    Sunday = iota  // 0
    Monday         // 1
    Tuesday        // 2
    Wednesday      // 3
    Thursday       // 4
    Friday         // 5
    Saturday       // 6
)
```

C의 enum과 비교:

```c
// C
enum Day {
    Sunday,    // 0
    Monday,    // 1
    Tuesday,   // 2
    // ...
};
```

`iota`는 더 유연합니다. 비트 플래그도 쉽게 만들 수 있습니다.

```go
const (
    FlagRead    = 1 << iota  // 1 (0b0001)
    FlagWrite                // 2 (0b0010)
    FlagExecute              // 4 (0b0100)
)
```

## 2.3 기본 타입(Built-in Types)

| 분류 | 타입 | 설명 |
|---|---|---|
| 정수 | `int8`, `int16`, `int32`, `int64` | 부호 있는 정수 |
| | `uint8`, `uint16`, `uint32`, `uint64` | 부호 없는 정수 |
| | `int`, `uint` | 플랫폼 의존 (보통 64비트) |
| | `byte` | `uint8`의 별칭 |
| | `rune` | `int32`의 별칭 (유니코드 코드포인트) |
| 실수 | `float32`, `float64` | IEEE 754 |
| 복소수 | `complex64`, `complex128` | C에 없는 타입 |
| 불린 | `bool` | `true` / `false` |
| 문자열 | `string` | **UTF-8 불변 시퀀스** |

### 🔍 C와의 핵심 차이

1. **타입 크기가 명확함**: C의 `int`는 컴파일러/플랫폼마다 다르지만, Go의 `int32`는 **언제나 정확히 32비트**입니다.
2. **암시적 형변환 없음**: C는 `int + float`이 자동으로 됩니다. Go는 **반드시 명시적 변환** 필요.

```go
var i int = 10
var f float64 = 3.14

// result := i + f       // ❌ 컴파일 에러!
result := float64(i) + f  // ✅ 명시적 변환 필요
```

```c
// C는 자동 변환
int i = 10;
double f = 3.14;
double result = i + f;  // OK (암묵적 변환)
```

> **Go의 철학**: "암묵적 동작은 버그의 원천이다." 모든 변환을 눈에 보이게 만든다.

## 2.4 문자열 — C와 가장 다른 부분

C의 문자열은 `char*` + null terminator(`'\0'`)입니다. 길이는 `strlen()`으로 매번 세야 합니다.

Go의 문자열은 다릅니다.

```go
s := "안녕하세요, Go!"

fmt.Println(len(s))         // 19 (바이트 수)
fmt.Println(len([]rune(s))) // 9 (문자 수)
```

### Go 문자열의 특징

1. **불변(immutable)** — `s[0] = 'A'` 같은 수정 불가
2. **UTF-8 인코딩 내장** — 한글, 이모지 등 자연스럽게 처리
3. **길이 정보를 자체 보유** — `len(s)`는 O(1)
4. **null terminator 없음** — 바이트 시퀀스 자체

### 문자열 다루기

```go
s := "Hello, Go!"

// 부분 문자열 (slice 문법)
fmt.Println(s[0:5])  // "Hello"

// 결합
s2 := s + " World"

// 길이
fmt.Println(len(s))  // 10

// 반복 (rune 단위로 순회)
for i, r := range "한글" {
    fmt.Printf("index=%d rune=%c\n", i, r)
}
```

### 🧪 실습 코드: type_basics.go

```go
package main

import "fmt"

func main() {
    // 1. 다양한 선언 방식
    var a int = 10
    var b = 20
    c := 30

    // 2. Zero value
    var d int
    var e string

    fmt.Printf("a=%d b=%d c=%d d=%d e=%q\n", a, b, c, d, e)

    // 3. 명시적 변환
    var x int = 100
    var y float64 = float64(x) * 1.5
    fmt.Printf("y=%f\n", y)

    // 4. 상수와 iota
    const (
        Sunday = iota
        Monday
        Tuesday
    )
    fmt.Println(Sunday, Monday, Tuesday)

    // 5. 문자열
    s := "Hello, 한글!"
    fmt.Println("바이트 길이:", len(s))
    fmt.Println("문자 길이:", len([]rune(s)))
}
```

실행:
```bash
go run type_basics.go
```

### ✅ 2교시 체크포인트

- [ ] `var`, `:=`, `const`를 적절히 구분하여 쓸 수 있는가?
- [ ] zero value 개념을 설명할 수 있는가?
- [ ] C와 달리 Go는 암묵적 형변환이 없다는 점을 이해했는가?
- [ ] 문자열이 UTF-8 기반이라는 점을 이해했는가?

---

# 3교시. Go 기본 문법 II — 함수, 다중 반환값, 포인터

## 3.1 함수 — 가장 큰 변화

```go
func add(a int, b int) int {
    return a + b
}

// 같은 타입이면 묶기 가능
func add2(a, b int) int {
    return a + b
}
```

### 🔍 C와의 비교

```c
// C
int add(int a, int b) {
    return a + b;
}
```

```go
// Go
func add(a int, b int) int {
    return a + b
}
```

차이:
- **`func` 키워드**로 시작
- **반환 타입을 함수 뒤**에 둠
- **매개변수도 변수명 뒤에 타입**

## 3.2 다중 반환값 — Go의 시그니처 기능

C에서 함수가 여러 값을 반환하려면 포인터 매개변수를 써야 합니다.

```c
// C — 포인터로 두 번째 값 받기
int divide(int a, int b, int *remainder) {
    *remainder = a % b;
    return a / b;
}

int rem;
int quot = divide(10, 3, &rem);
```

Go는 그냥 여러 개 반환합니다.

```go
func divide(a, b int) (int, int) {
    return a / b, a % b
}

quot, rem := divide(10, 3)
fmt.Println(quot, rem)  // 3 1
```

### 명명된 반환값

```go
func divide(a, b int) (quotient int, remainder int) {
    quotient = a / b
    remainder = a % b
    return  // naked return — 자동으로 quotient, remainder 반환
}
```

### 에러 처리 패턴 — Go의 관용구

다중 반환값의 가장 큰 용도는 **에러 처리**입니다. C에서 함수가 실패하면 보통 -1을 반환하거나 `errno`를 설정하지만, Go는 명시적으로 에러를 반환합니다.

```go
import (
    "fmt"
    "os"
)

func main() {
    data, err := os.ReadFile("hello.txt")
    if err != nil {
        fmt.Println("파일 읽기 실패:", err)
        return
    }
    fmt.Println(string(data))
}
```

> **이 패턴이 Go 코드의 절반 이상을 차지합니다.** 반환값 두 번째에 에러가 있다는 컨벤션을 외워두세요.

## 3.3 포인터 — C와 비슷하지만 다르다

### Go 포인터 기본

```go
x := 10
p := &x        // p는 *int 타입 (x의 주소)
fmt.Println(*p) // 10 — 역참조

*p = 20         // x의 값이 20으로 변경
fmt.Println(x)  // 20
```

문법은 C와 똑같이 `&` (주소), `*` (역참조)입니다.

### 🔥 가장 중요한 차이: 포인터 산술 금지

```c
// C — 가능 (위험하지만)
int arr[5] = {1, 2, 3, 4, 5};
int *p = arr;
p++;         // OK — 다음 원소로 이동
printf("%d\n", *p);  // 2
```

```go
// Go — 불가능
arr := [5]int{1, 2, 3, 4, 5}
p := &arr[0]
// p++         // ❌ 컴파일 에러
// p = p + 1   // ❌ 컴파일 에러
```

**왜 막았을까?** 포인터 산술은 C 보안 취약점의 절대 다수(버퍼 오버플로우 등)의 원인입니다. Go는 안전성을 위해 의도적으로 제거했습니다. 배열을 순회하려면 슬라이스와 인덱스를 사용합니다.

### Go에 없는 것: 포인터의 포인터(`**T`)?

있긴 한데, **거의 안 씁니다**. C의 `argv`(`char **argv`) 같은 패턴은 Go에서 `[]string`(슬라이스)으로 대체됩니다.

### 함수에 포인터 넘기기

C와 동일하게 값을 수정하려면 포인터를 넘깁니다.

```go
func zero(p *int) {
    *p = 0
}

x := 100
zero(&x)
fmt.Println(x)  // 0
```

### `new()` — 메모리 할당

C의 `malloc` 대신 Go는 `new()`를 제공합니다 (단, **free는 없습니다 — GC가 처리**).

```go
p := new(int)   // *int, 0으로 초기화
*p = 42
fmt.Println(*p) // 42
// free(p) 같은 거 안 함 — GC가 자동 회수
```

### 🧪 실습 코드: pointer_practice.go

```go
package main

import "fmt"

// 두 값을 교환
func swap(a, b *int) {
    *a, *b = *b, *a
}

// 안전한 나눗셈 — 다중 반환값
func safeDiv(a, b int) (int, error) {
    if b == 0 {
        return 0, fmt.Errorf("0으로 나눌 수 없습니다")
    }
    return a / b, nil
}

func main() {
    // 1. 포인터로 교환
    x, y := 10, 20
    swap(&x, &y)
    fmt.Println("after swap:", x, y)  // 20 10

    // 2. 다중 반환값 + 에러 처리
    result, err := safeDiv(10, 0)
    if err != nil {
        fmt.Println("에러:", err)
    } else {
        fmt.Println("결과:", result)
    }

    result, err = safeDiv(10, 3)
    if err != nil {
        fmt.Println("에러:", err)
    } else {
        fmt.Println("결과:", result)
    }
}
```

### ✅ 3교시 체크포인트

- [ ] 다중 반환값을 활용한 함수를 작성할 수 있는가?
- [ ] 에러 처리 패턴 `if err != nil`을 자연스럽게 쓸 수 있는가?
- [ ] Go 포인터와 C 포인터의 차이를 설명할 수 있는가?

---

# 4교시. Go 기본 문법 III — 구조체, 메서드

## 4.1 구조체 — C의 struct와 거의 동일

```go
type Person struct {
    Name string
    Age  int
}
```

C와 비교:

```c
// C
struct Person {
    char name[50];
    int age;
};

struct Person p;
strcpy(p.name, "Alice");
p.age = 30;
```

```go
// Go
p := Person{Name: "Alice", Age: 30}
// 또는
p := Person{"Alice", 30}  // 순서대로

// 필드 접근은 . (C와 동일)
fmt.Println(p.Name)
```

### 포인터로 구조체 다루기

```go
p := &Person{Name: "Bob", Age: 25}
fmt.Println(p.Name)  // (*p).Name이 아니라 그냥 p.Name — Go가 자동 역참조
```

> C에서는 `p->name` vs `s.name`을 구분해야 하지만, **Go는 둘 다 `.`을 씁니다.** 컴파일러가 알아서 역참조합니다.

## 4.2 메서드 — 구조체에 함수 붙이기

C에는 없는 개념입니다. C++의 멤버 함수와 비슷하지만 더 단순합니다.

```go
type Rectangle struct {
    Width, Height float64
}

// 메서드 — Rectangle 타입에 Area라는 함수를 붙임
func (r Rectangle) Area() float64 {
    return r.Width * r.Height
}

func main() {
    rect := Rectangle{Width: 3, Height: 4}
    fmt.Println(rect.Area())  // 12
}
```

### `func (r Rectangle)` 의 의미 — Receiver

`(r Rectangle)`을 **리시버(receiver)**라고 합니다. C의 첫 번째 인자로 `self` 또는 `this`를 받는 패턴과 같습니다.

```c
// C — 비슷한 패턴
double rectangle_area(Rectangle *r) {
    return r->width * r->height;
}
```

```go
// Go — 더 자연스러움
func (r Rectangle) Area() float64 {
    return r.Width * r.Height
}
```

### 값 리시버 vs 포인터 리시버

```go
// 값 리시버 — 복사본을 받음
func (r Rectangle) Scale(factor float64) {
    r.Width *= factor  // 원본 안 변함!
}

// 포인터 리시버 — 원본을 수정 가능
func (r *Rectangle) ScalePtr(factor float64) {
    r.Width *= factor  // 원본 변함
}
```

**언제 어떤 걸 쓰나?**

| 상황 | 리시버 선택 |
|---|---|
| 구조체를 수정해야 한다 | 포인터 리시버 `*T` |
| 구조체가 크다(복사 비용) | 포인터 리시버 |
| 작고 변경 없는 데이터 | 값 리시버 `T` |
| **혼용하지 말 것** | 한 타입의 메서드는 일관성 있게 |

### 🧪 실습 코드: struct_methods.go

```go
package main

import "fmt"

type Account struct {
    Owner   string
    Balance float64
}

// 잔액 조회 — 값 리시버
func (a Account) GetBalance() float64 {
    return a.Balance
}

// 입금 — 포인터 리시버 (수정 필요)
func (a *Account) Deposit(amount float64) {
    a.Balance += amount
}

// 출금 — 포인터 리시버 + 에러 반환
func (a *Account) Withdraw(amount float64) error {
    if amount > a.Balance {
        return fmt.Errorf("잔액 부족: 요청 %.2f, 잔액 %.2f", amount, a.Balance)
    }
    a.Balance -= amount
    return nil
}

func main() {
    acc := Account{Owner: "Alice", Balance: 1000}

    acc.Deposit(500)
    fmt.Printf("입금 후: %.2f\n", acc.GetBalance())

    if err := acc.Withdraw(2000); err != nil {
        fmt.Println("출금 실패:", err)
    }

    if err := acc.Withdraw(300); err != nil {
        fmt.Println("출금 실패:", err)
    } else {
        fmt.Printf("출금 후: %.2f\n", acc.GetBalance())
    }
}
```

### ✅ 4교시 체크포인트

- [ ] 구조체를 정의하고 초기화할 수 있는가?
- [ ] 메서드의 값 리시버와 포인터 리시버를 구분해서 쓸 수 있는가?
- [ ] C struct와 Go struct의 사용 편의성 차이를 느꼈는가?

---

# 5교시. Go 자료구조 — 배열, 슬라이스, 맵

## 5.1 배열 — 고정 크기

```go
var arr [5]int           // [0 0 0 0 0]
arr2 := [3]int{1, 2, 3}  // [1 2 3]
arr3 := [...]int{1, 2, 3, 4}  // 크기 자동 계산 = 4
```

C와 비교:

```c
// C
int arr[5];           // 쓰레기값 — 위험!
int arr2[3] = {1, 2, 3};
```

```go
// Go
var arr [5]int        // 자동 0 초기화
```

### 🔥 중요: Go 배열은 "값 타입"

C에서 배열을 함수에 넘기면 **포인터로 전달**(decay)됩니다. Go는 다릅니다.

```go
func modify(a [3]int) {
    a[0] = 999  // 복사본을 수정 — 원본은 그대로!
}

arr := [3]int{1, 2, 3}
modify(arr)
fmt.Println(arr)  // [1 2 3] — 안 변함!
```

```c
// C — 포인터로 전달되므로 원본이 변함
void modify(int a[]) {
    a[0] = 999;
}

int arr[3] = {1, 2, 3};
modify(arr);
// arr is now {999, 2, 3}
```

이런 이유로 **Go에서는 배열을 직접 쓰는 일이 드뭅니다.** 대신 **슬라이스**를 씁니다.

## 5.2 슬라이스 — 가장 중요한 자료구조

슬라이스는 **가변 길이 배열**입니다. 내부적으로는 (포인터, 길이, 용량)의 구조입니다.

```go
s := []int{1, 2, 3}     // 배열 선언과 비슷하지만 [] 안에 숫자 없음
s = append(s, 4)        // [1 2 3 4]
s = append(s, 5, 6, 7)  // [1 2 3 4 5 6 7]

fmt.Println(len(s))  // 7 — 길이
fmt.Println(cap(s))  // 8 (구현에 따라 다름) — 용량
```

### 슬라이스의 내부 구조

```
slice  →  ┌────────────┬─────┬─────┐
          │ pointer    │ len │ cap │
          └─────┬──────┴─────┴─────┘
                ↓
배열:    [1, 2, 3, 4, 5, _, _, _]
          ←─── len=5 ───→
          ←─────── cap=8 ─────────→
```

- **pointer**: 실제 데이터 배열의 시작
- **len**: 현재 원소 개수
- **cap**: 재할당 없이 담을 수 있는 최대 개수

### `make()` 로 슬라이스 만들기

```go
s := make([]int, 5)       // 길이 5, 용량 5
s := make([]int, 5, 10)   // 길이 5, 용량 10
```

### 슬라이싱 — 부분 잘라내기

```go
s := []int{10, 20, 30, 40, 50}
sub := s[1:4]   // [20 30 40]
sub2 := s[:3]   // [10 20 30]
sub3 := s[2:]   // [30 40 50]
```

**주의: 슬라이스는 원본 배열을 공유합니다!**

```go
s := []int{1, 2, 3, 4, 5}
sub := s[1:4]
sub[0] = 999
fmt.Println(s)  // [1 999 3 4 5] — 원본도 변함!
```

이건 처음엔 함정처럼 느껴지지만, **C의 포인터를 통한 효율성을 그대로 살리는 설계**입니다.

### append의 동작 — 용량 확장

```go
s := make([]int, 0, 3)
fmt.Println(len(s), cap(s))  // 0 3
s = append(s, 1, 2, 3)
fmt.Println(len(s), cap(s))  // 3 3
s = append(s, 4)  // 용량 초과 → 새 배열 할당
fmt.Println(len(s), cap(s))  // 4 6 (2배로 늘어남)
```

C에서 `realloc()`을 직접 해주던 일을 `append`가 알아서 처리합니다.

## 5.3 맵(map) — 해시 테이블

C에는 표준 해시맵이 없어서 직접 구현하거나 라이브러리를 써야 합니다. Go는 내장입니다.

```go
m := map[string]int{
    "apple":  100,
    "banana": 200,
}

m["cherry"] = 300

// 조회 — 두 값 반환 (값, 존재 여부)
if v, ok := m["apple"]; ok {
    fmt.Println("apple:", v)
}

// 삭제
delete(m, "banana")

// 순회 — 순서는 매번 다름!
for k, v := range m {
    fmt.Printf("%s = %d\n", k, v)
}
```

### `make()` 로 맵 만들기

```go
m := make(map[string]int)
m := make(map[string]int, 100)  // 초기 용량 힌트
```

> **중요**: `var m map[string]int`로만 선언하면 `nil` 맵이 되어 쓰기 시 패닉이 발생합니다. **반드시 `make` 또는 리터럴로 초기화**하세요.

### 🧪 실습 코드: collections.go

```go
package main

import "fmt"

func main() {
    // 1. 슬라이스 기본
    nums := []int{10, 20, 30}
    nums = append(nums, 40, 50)
    fmt.Println("nums:", nums)
    fmt.Println("len:", len(nums), "cap:", cap(nums))

    // 2. 슬라이싱
    sub := nums[1:4]
    fmt.Println("sub:", sub)

    // 3. 합계 함수
    fmt.Println("sum:", sum(nums))

    // 4. 맵으로 단어 빈도수
    text := []string{"go", "is", "fun", "go", "is", "fast", "go"}
    count := wordCount(text)
    for k, v := range count {
        fmt.Printf("%s: %d\n", k, v)
    }
}

func sum(s []int) int {
    total := 0
    for _, v := range s {
        total += v
    }
    return total
}

func wordCount(words []string) map[string]int {
    m := make(map[string]int)
    for _, w := range words {
        m[w]++
    }
    return m
}
```

### range 키워드

`range`는 슬라이스/맵을 순회하는 키워드입니다.

```go
// 슬라이스: 인덱스, 값
for i, v := range []int{10, 20, 30} {
    fmt.Println(i, v)
}

// 인덱스만 쓰고 싶을 때
for i := range nums {
    // ...
}

// 값만 쓰고 싶을 때 (인덱스 무시는 _)
for _, v := range nums {
    // ...
}

// 맵: 키, 값
for k, v := range m {
    // ...
}
```

### ✅ 5교시 체크포인트

- [ ] 배열과 슬라이스의 차이를 설명할 수 있는가?
- [ ] 슬라이스가 원본을 공유한다는 점을 이해했는가?
- [ ] `make`, `append`, `len`, `cap`을 사용할 수 있는가?
- [ ] 맵의 두 값 반환(`v, ok`)을 이해했는가?

---

# 6교시. 인터페이스 기초 — Duck Typing

## 6.1 인터페이스란?

> "오리처럼 걷고 오리처럼 운다면, 그것은 오리다." — Duck Typing

C++에서 인터페이스는 "이 클래스가 어떤 인터페이스를 **구현한다**고 선언"해야 합니다(`class Dog : public Animal`). Go는 다릅니다.

**Go에서는 메서드 시그니처가 일치하기만 하면 자동으로 인터페이스를 만족합니다.**

```go
// 인터페이스 정의 — "Speak 메서드를 가진 모든 것"
type Speaker interface {
    Speak() string
}

// Dog 구조체
type Dog struct{}

func (d Dog) Speak() string {
    return "멍멍!"
}

// Cat 구조체
type Cat struct{}

func (c Cat) Speak() string {
    return "야옹!"
}

func main() {
    // 어디에도 "Dog는 Speaker이다"라고 명시하지 않았지만,
    // Speak() 메서드를 가지므로 자동으로 Speaker 인터페이스 만족
    var s Speaker

    s = Dog{}
    fmt.Println(s.Speak())  // 멍멍!

    s = Cat{}
    fmt.Println(s.Speak())  // 야옹!
}
```

### 🔍 C++ 가상함수와 비교

```cpp
// C++ — 명시적 상속 필요
class Speaker {
public:
    virtual std::string speak() = 0;
};

class Dog : public Speaker {  // 명시!
public:
    std::string speak() override { return "멍멍!"; }
};
```

```go
// Go — 암묵적, 선언 불필요
type Speaker interface {
    Speak() string
}

type Dog struct{}
func (d Dog) Speak() string { return "멍멍!" }
// "Speaker를 구현한다"고 적지 않음 — 메서드 시그니처가 맞으면 자동 충족
```

> **장점**: 라이브러리에서 정의한 타입에 대해서도 **나중에 인터페이스를 추가**할 수 있습니다. C++에서는 불가능합니다.

## 6.2 가장 많이 쓰이는 인터페이스: `fmt.Stringer`

표준 라이브러리에 정의되어 있는 인터페이스입니다.

```go
type Stringer interface {
    String() string
}
```

내 타입에 `String() string` 메서드를 붙이면 `fmt.Println`이 자동으로 그 함수를 호출합니다.

```go
type Point struct {
    X, Y int
}

func (p Point) String() string {
    return fmt.Sprintf("(%d, %d)", p.X, p.Y)
}

func main() {
    p := Point{3, 4}
    fmt.Println(p)  // (3, 4) — String() 자동 호출
}
```

C에서 구조체를 출력하려면 항상 `printf("(%d, %d)", p.x, p.y)`를 손으로 적어야 했죠. Go는 한 번만 메서드를 정의하면 됩니다.

## 6.3 빈 인터페이스 `interface{}` (또는 `any`)

```go
var x interface{}  // 또는 var x any (Go 1.18+)
x = 10
x = "hello"
x = Point{1, 2}
```

**모든 타입을 받을 수 있는 만능 타입**입니다. C의 `void *`와 비슷한 역할.

### 타입 단언(Type Assertion)

```go
var i any = "hello"

s := i.(string)        // OK
s, ok := i.(string)    // 안전한 방식

// 타입 스위치
switch v := i.(type) {
case int:
    fmt.Println("int:", v)
case string:
    fmt.Println("string:", v)
default:
    fmt.Println("unknown")
}
```

### 🧪 실습 코드: interface_basic.go

```go
package main

import "fmt"

// Shape 인터페이스
type Shape interface {
    Area() float64
    Perimeter() float64
}

type Rectangle struct {
    Width, Height float64
}

func (r Rectangle) Area() float64 {
    return r.Width * r.Height
}

func (r Rectangle) Perimeter() float64 {
    return 2 * (r.Width + r.Height)
}

type Circle struct {
    Radius float64
}

func (c Circle) Area() float64 {
    return 3.14159 * c.Radius * c.Radius
}

func (c Circle) Perimeter() float64 {
    return 2 * 3.14159 * c.Radius
}

// 인터페이스만 보고 동작 — 다형성
func describe(s Shape) {
    fmt.Printf("Area=%.2f  Perimeter=%.2f\n", s.Area(), s.Perimeter())
}

func main() {
    shapes := []Shape{
        Rectangle{Width: 3, Height: 4},
        Circle{Radius: 5},
    }

    for _, s := range shapes {
        describe(s)
    }
}
```

### ✅ 6교시 체크포인트

- [ ] Duck Typing의 의미를 설명할 수 있는가?
- [ ] 명시적 선언 없이 인터페이스를 구현할 수 있는가?
- [ ] `interface{}` 또는 `any`를 언제 쓰는지 이해했는가?
- [ ] C++ 가상함수와 Go 인터페이스의 차이를 말할 수 있는가?

---

# 7교시. 실습 — C 코드를 Go로 변환

이번 시간은 **실제 C 코드**를 보고 Go로 옮겨봅니다. 단순한 번역이 아니라, **Go다운 스타일**로 다시 쓰는 게 목표입니다.

## 실습 ①: 학생 평균 점수 계산

### 원본 C 코드

```c
#include <stdio.h>
#include <string.h>

typedef struct {
    char name[50];
    int  scores[3];   // 국어, 영어, 수학
} Student;

double average(int scores[], int n) {
    int sum = 0;
    for (int i = 0; i < n; i++) sum += scores[i];
    return (double)sum / n;
}

int main(void) {
    Student students[3] = {
        {"Alice", {90, 85, 92}},
        {"Bob",   {70, 65, 80}},
        {"Carol", {95, 100, 88}},
    };

    for (int i = 0; i < 3; i++) {
        double avg = average(students[i].scores, 3);
        printf("%s: %.2f\n", students[i].name, avg);
    }
    return 0;
}
```

### 🔄 Go 변환 — 단계별

**Step 1. 패키지와 import 설정**

```go
package main

import "fmt"
```

**Step 2. 구조체 정의 (C struct → Go struct)**

```go
type Student struct {
    Name   string
    Scores []int  // 고정 크기 대신 슬라이스 사용 — Go다움
}
```

**Step 3. 평균 함수 (배열 + 길이 인자 → 슬라이스)**

```go
func average(scores []int) float64 {
    sum := 0
    for _, s := range scores {
        sum += s
    }
    return float64(sum) / float64(len(scores))
}
```

C에서는 `int scores[], int n`처럼 길이를 따로 받지만, Go는 슬라이스에 길이가 들어있어 더 간결합니다.

**Step 4. main 함수**

```go
func main() {
    students := []Student{
        {"Alice", []int{90, 85, 92}},
        {"Bob",   []int{70, 65, 80}},
        {"Carol", []int{95, 100, 88}},
    }

    for _, s := range students {
        avg := average(s.Scores)
        fmt.Printf("%s: %.2f\n", s.Name, avg)
    }
}
```

### 최종 Go 코드 (`student.go`)

```go
package main

import "fmt"

type Student struct {
    Name   string
    Scores []int
}

func average(scores []int) float64 {
    if len(scores) == 0 {
        return 0
    }
    sum := 0
    for _, s := range scores {
        sum += s
    }
    return float64(sum) / float64(len(scores))
}

func main() {
    students := []Student{
        {"Alice", []int{90, 85, 92}},
        {"Bob",   []int{70, 65, 80}},
        {"Carol", []int{95, 100, 88}},
    }

    for _, s := range students {
        avg := average(s.Scores)
        fmt.Printf("%s: %.2f\n", s.Name, avg)
    }
}
```

실행:
```bash
go run student.go
```

기대 출력:
```
Alice: 89.00
Bob: 71.67
Carol: 94.33
```

### 비교 포인트

| 항목 | C | Go |
|---|---|---|
| 문자열 | `char name[50]` (고정 50바이트) | `string` (가변, 효율적) |
| 배열 길이 | 함수에 따로 전달 | 슬라이스에 내장 |
| 메모리 안전 | 오버플로우 가능 | 자동 검사 |
| 0 나누기 방어 | 작성 안 함 (위험) | `if len == 0` 명시 |

## 실습 ②: 단순 연결 리스트 → Go 슬라이스

### 원본 C 코드

```c
#include <stdio.h>
#include <stdlib.h>

typedef struct Node {
    int          value;
    struct Node *next;
} Node;

Node *push(Node *head, int v) {
    Node *n = malloc(sizeof(Node));
    n->value = v;
    n->next  = head;
    return n;
}

void print_list(Node *head) {
    while (head) {
        printf("%d -> ", head->value);
        head = head->next;
    }
    printf("NULL\n");
}

void free_list(Node *head) {
    while (head) {
        Node *next = head->next;
        free(head);
        head = next;
    }
}

int main(void) {
    Node *head = NULL;
    head = push(head, 1);
    head = push(head, 2);
    head = push(head, 3);
    print_list(head);
    free_list(head);
    return 0;
}
```

### 🔄 Go 변환 — 슬라이스가 답이다

Go에서는 연결 리스트를 손으로 만들 필요가 거의 없습니다. 슬라이스가 대부분의 용도를 커버합니다.

```go
package main

import "fmt"

func main() {
    var list []int
    list = append(list, 1)
    list = append(list, 2)
    list = append(list, 3)
    // 또는: list := []int{1, 2, 3}

    // 출력
    for i, v := range list {
        if i > 0 {
            fmt.Print(" -> ")
        }
        fmt.Print(v)
    }
    fmt.Println()

    // free 같은 거 안 함 — GC가 알아서
}
```

**얻은 것**:
- `malloc`/`free` 불필요
- `Node` 구조체 불필요
- 메모리 누수 걱정 없음
- 코드 1/3로 단축

**잃은 것**:
- 직접 메모리 레이아웃 제어 (대부분 필요 없음)

### ✅ 7교시 체크포인트

- [ ] C struct를 Go struct로 옮길 수 있는가?
- [ ] 고정 배열 + 길이 인자를 슬라이스로 대체할 수 있는가?
- [ ] malloc/free 없이 동적 자료구조를 다룰 수 있는가?

---

# 8교시. 실습 — 간단한 CLI 프로그램 작성

마지막 실습으로 **단어 빈도수 계산기**를 만듭니다. 표준 입력 또는 파일에서 텍스트를 읽어 단어별 등장 횟수를 출력하는 프로그램입니다.

## 8.1 요구사항

1. 명령행 인자로 파일 경로를 받는다.
2. 인자가 없으면 표준 입력(stdin)을 읽는다.
3. 단어를 공백 기준으로 분리한다.
4. 단어별 등장 횟수를 내림차순으로 정렬해 출력한다.
5. 상위 10개만 출력한다.
6. 에러는 명확한 메시지와 함께 종료 코드 1로 나간다.

## 8.2 단계별 구현

### Step 1. 프로젝트 생성

```bash
mkdir -p ~/go-class/day1/wordcount
cd ~/go-class/day1/wordcount
go mod init wordcount
touch main.go
```

### Step 2. 골격 작성

```go
package main

import (
    "fmt"
    "os"
)

func main() {
    fmt.Println("args:", os.Args)
}
```

실행해 인자가 어떻게 들어오는지 확인합니다.

```bash
go run main.go hello.txt foo bar
# args: [/tmp/.../main hello.txt foo bar]
```

> `os.Args[0]`은 프로그램 이름, `os.Args[1:]`이 실제 인자입니다. C의 `argv`와 같습니다.

### Step 3. 입력 소스 결정

```go
package main

import (
    "fmt"
    "io"
    "os"
)

func main() {
    var reader io.Reader

    if len(os.Args) > 1 {
        // 파일 모드
        f, err := os.Open(os.Args[1])
        if err != nil {
            fmt.Fprintf(os.Stderr, "파일 열기 실패: %v\n", err)
            os.Exit(1)
        }
        defer f.Close()
        reader = f
    } else {
        // stdin 모드
        reader = os.Stdin
    }

    // 일단 그대로 출력
    io.Copy(os.Stdout, reader)
}
```

`defer f.Close()`는 **함수가 끝날 때 자동으로 실행**되는 구문입니다. C에서 `goto cleanup` 패턴을 쓰던 일을 한 줄로 해결합니다.

실행:
```bash
echo "hello world" > test.txt
go run main.go test.txt
# hello world

echo "from stdin" | go run main.go
# from stdin
```

### Step 4. 단어 단위로 읽기

```go
import (
    "bufio"
    "fmt"
    "io"
    "os"
)

func countWords(r io.Reader) map[string]int {
    counts := make(map[string]int)
    scanner := bufio.NewScanner(r)
    scanner.Split(bufio.ScanWords)  // 공백 단위 분리
    for scanner.Scan() {
        word := scanner.Text()
        counts[word]++
    }
    return counts
}
```

- `bufio.Scanner`: 줄/단어/문자 단위로 효율적으로 읽는 도구
- `bufio.ScanWords`: 공백을 구분자로 단어 추출
- `counts[word]++`: 존재하지 않으면 0에서 시작 → 1 (Go 맵의 편리한 점)

### Step 5. 정렬

```go
import (
    "sort"
    // ...
)

type wordFreq struct {
    Word  string
    Count int
}

func topN(counts map[string]int, n int) []wordFreq {
    list := make([]wordFreq, 0, len(counts))
    for w, c := range counts {
        list = append(list, wordFreq{w, c})
    }
    sort.Slice(list, func(i, j int) bool {
        return list[i].Count > list[j].Count  // 내림차순
    })
    if len(list) > n {
        list = list[:n]
    }
    return list
}
```

`sort.Slice`는 **익명 함수(클로저)**를 비교자로 받습니다. C의 `qsort` + 함수 포인터와 비슷하지만 훨씬 간결합니다.

### Step 6. 전체 합치기 — 최종 `main.go`

```go
package main

import (
    "bufio"
    "fmt"
    "io"
    "os"
    "sort"
)

type wordFreq struct {
    Word  string
    Count int
}

func countWords(r io.Reader) map[string]int {
    counts := make(map[string]int)
    scanner := bufio.NewScanner(r)
    scanner.Split(bufio.ScanWords)
    for scanner.Scan() {
        word := scanner.Text()
        counts[word]++
    }
    return counts
}

func topN(counts map[string]int, n int) []wordFreq {
    list := make([]wordFreq, 0, len(counts))
    for w, c := range counts {
        list = append(list, wordFreq{w, c})
    }
    sort.Slice(list, func(i, j int) bool {
        return list[i].Count > list[j].Count
    })
    if len(list) > n {
        list = list[:n]
    }
    return list
}

func openInput() (io.ReadCloser, error) {
    if len(os.Args) > 1 {
        return os.Open(os.Args[1])
    }
    // stdin은 Close 없음 — nop wrapper
    return io.NopCloser(os.Stdin), nil
}

func main() {
    in, err := openInput()
    if err != nil {
        fmt.Fprintf(os.Stderr, "입력 열기 실패: %v\n", err)
        os.Exit(1)
    }
    defer in.Close()

    counts := countWords(in)
    top := topN(counts, 10)

    fmt.Println("=== Top 10 Words ===")
    for i, wf := range top {
        fmt.Printf("%2d. %-20s %d\n", i+1, wf.Word, wf.Count)
    }
}
```

### Step 7. 테스트

테스트용 텍스트:
```bash
cat > sample.txt << 'EOF'
go is a programming language go was created at google
go is fast go is simple go is fun go go go
EOF

go run main.go sample.txt
```

기대 출력:
```
=== Top 10 Words ===
 1. go                   7
 2. is                   4
 3. a                    1
 4. programming          1
 5. language             1
 ...
```

stdin 모드:
```bash
echo "hello hello world go go go go" | go run main.go
```

### Step 8. 빌드 후 단일 바이너리 배포

```bash
go build -o wordcount
./wordcount sample.txt

# 정적 링크된 단일 바이너리!
file wordcount
# wordcount: ELF 64-bit LSB executable, statically linked, ...

ls -lh wordcount
# 약 2MB
```

> **Go의 강점**: C++/Rust처럼 컴파일된 단일 바이너리가 나옵니다. 의존성 라이브러리 같이 배포할 필요 없이 그냥 복사만 하면 됩니다.

### 🎯 도전 과제 (선택)

배운 것을 활용해 다음 기능을 추가해보세요.

1. **대소문자 통일**: "Go"와 "go"를 같은 단어로 취급 (힌트: `strings.ToLower`)
2. **구두점 제거**: "hello," 와 "hello"를 같이 취급
3. **상위 N개 인자**: `-n 5` 같은 옵션 (힌트: `flag` 패키지)
4. **최소 길이 필터**: 1글자 단어는 제외

### ✅ 8교시 체크포인트

- [ ] `os.Args`로 명령행 인자를 받을 수 있는가?
- [ ] 파일과 stdin을 같은 코드로 처리할 수 있는가? (io.Reader 추상화)
- [ ] `defer`의 동작을 이해했는가?
- [ ] `sort.Slice`로 정렬을 할 수 있는가?
- [ ] `go build`로 단일 바이너리를 만들 수 있는가?

---

# 🎓 1일차 마무리

## 오늘 배운 것

1. **Go의 철학**: 단순함, 명시성, 합성 중심 설계
2. **개발환경**: Go 설치, `go mod`, `go run`, `go build`
3. **기본 문법**: 변수, 상수, 다중 반환값, 포인터
4. **구조체와 메서드**: 값 리시버 vs 포인터 리시버
5. **자료구조**: 배열, 슬라이스, 맵 (실무에선 슬라이스가 압도적)
6. **인터페이스**: Duck Typing의 자유로움
7. **실전**: C 코드 변환, 작동하는 CLI 프로그램 작성

## 한 줄 요약

> **"Go는 C의 단순함 + 안전성 + 동시성"** — 오늘은 앞의 두 가지를 배웠고, 내일부터 본격적으로 동시성 세계를 만납니다.

## 복습 과제

다음 시간 전에 다음 코드를 직접 작성해보고 오세요.

1. **FizzBuzz**를 Go로 작성 (1~30까지, 3의 배수면 Fizz, 5의 배수면 Buzz, 둘 다면 FizzBuzz)
2. **회문(Palindrome) 검사** 함수 만들기 — `func isPalindrome(s string) bool`
3. **간단한 스택 자료구조**를 슬라이스로 구현 (Push, Pop, Peek 메서드)

## 다음 시간 예고 — 2일차

- **에러 처리 패턴**: `error` 인터페이스, `errors` 패키지, `panic`/`recover`
- **모듈 시스템**: `go mod` 깊이 있게, 의존성 관리 전략
- **패키지 설계**: 가시성 규칙, 폴더 구조
- **빌드 시스템**: `go build`, `go install`, `Makefile` 통합
- **실습**: 멀티 패키지 프로젝트 구성

---

## 📚 참고 자료

- [공식 Go Tour (한글)](https://go-tour-ko.appspot.com/)
- [Effective Go (영문)](https://go.dev/doc/effective_go) — 필독 권장
- [Go by Example](https://gobyexample.com/) — 짧은 예제 모음
- [Go 표준 라이브러리 문서](https://pkg.go.dev/std)
- 책: 『프로그래밍 언어 Go』 (Donovan & Kernighan) — C 만든 사람이 쓴 책

> Kernighan은 C 책(K&R)을 쓴 분이기도 합니다. Go 책에서 C 출신 시각으로 비교 설명이 많습니다. 강력 추천.
