# Go 벤치마크 실습자료

## 목표

Go의 `testing` 패키지를 사용해서 함수 성능을 측정하는 방법을 단계적으로 실습합니다.

최종적으로 다음 명령을 사용할 수 있습니다.

```bash
go test -bench=.
go test -bench=. -benchmem
go test -run=^$ -bench=.
```

---

# 1단계. 프로젝트 만들기

```bash
mkdir go-benchmark-practice
cd go-benchmark-practice
go mod init go-benchmark-practice
```

구조:

```text
go-benchmark-practice/
└── go.mod
```

---

# 2단계. 계산 함수 만들기

`calculator.go` 파일을 만듭니다.

```go
package calculator

import "fmt"

func Add(a int, b int) int {
	return a + b
}

func Mul(a int, b int) int {
	return a * b
}

func Calculate(op string, a int, b int) (int, error) {
	switch op {
	case "add":
		return a + b, nil
	case "mul":
		return a * b, nil
	default:
		return 0, fmt.Errorf("지원하지 않는 연산입니다: %s", op)
	}
}
```

---

# 3단계. 일반 테스트 작성하기

벤치마크 전에 함수가 정상 동작하는지 확인합니다.

`calculator_test.go`

```go
package calculator

import "testing"

func TestAdd(t *testing.T) {
	result := Add(10, 20)

	if result != 30 {
		t.Errorf("원하는 값: %d, 실제 값: %d", 30, result)
	}
}

func TestMul(t *testing.T) {
	result := Mul(10, 20)

	if result != 200 {
		t.Errorf("원하는 값: %d, 실제 값: %d", 200, result)
	}
}

func TestCalculate(t *testing.T) {
	result, err := Calculate("add", 10, 20)

	if err != nil {
		t.Errorf("에러가 발생하면 안 됩니다: %v", err)
	}

	if result != 30 {
		t.Errorf("원하는 값: %d, 실제 값: %d", 30, result)
	}
}
```

실행:

```bash
go test
```

자세히 보기:

```bash
go test -v
```

---

# 4단계. 가장 간단한 벤치마크 작성하기

`calculator_benchmark_test.go`

```go
package calculator

import "testing"

func BenchmarkAdd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Add(10, 20)
	}
}
```

벤치마크 함수 이름 규칙:

```go
func Benchmark이름(b *testing.B)
```

핵심 반복문:

```go
for i := 0; i < b.N; i++ {
	_ = Add(10, 20)
}
```

`b.N`은 Go가 자동으로 정하는 반복 횟수입니다.

실행:

```bash
go test -bench=.
```

출력 예:

```text
BenchmarkAdd-8        1000000000        0.3000 ns/op
PASS
```

---

# 5단계. 벤치마크 결과 해석하기

예:

```text
BenchmarkAdd-8        1000000000        0.3000 ns/op
```

| 항목 | 의미 |
|---|---|
| `BenchmarkAdd-8` | 벤치마크 이름 |
| `1000000000` | 반복 횟수 |
| `0.3000 ns/op` | 1회 실행당 평균 시간 |
| `ns/op` | nanoseconds per operation |

즉, `Add()` 함수 1회 실행 평균 시간이 `0.3000 ns`라는 뜻입니다.

---

# 6단계. 곱셈 벤치마크 추가하기

`calculator_benchmark_test.go`

```go
package calculator

import "testing"

func BenchmarkAdd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Add(10, 20)
	}
}

func BenchmarkMul(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Mul(10, 20)
	}
}
```

실행:

```bash
go test -bench=.
```

---

# 7단계. Calculate 함수 벤치마크하기

```go
func BenchmarkCalculateAdd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = Calculate("add", 10, 20)
	}
}
```

`Calculate()`는 내부에서 문자열 `"add"`를 비교하고 `switch`를 실행합니다.  
따라서 단순 `Add()`보다 느릴 수 있습니다.

비교 대상:

```go
Add(10, 20)
```

```go
Calculate("add", 10, 20)
```

---

# 8단계. 서브 벤치마크 작성하기

여러 연산을 하나의 벤치마크 함수에서 나누어 측정할 수 있습니다.

```go
func BenchmarkCalculate(b *testing.B) {
	benchmarks := []struct {
		name string
		op   string
		a    int
		b    int
	}{
		{"add", "add", 10, 20},
		{"mul", "mul", 10, 20},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = Calculate(bm.op, bm.a, bm.b)
			}
		})
	}
}
```

출력 예:

```text
BenchmarkCalculate/add-8        100000000        10.5 ns/op
BenchmarkCalculate/mul-8        100000000        10.6 ns/op
```

---

# 9단계. b.ResetTimer 사용하기

준비 작업이 있는 경우 준비 시간은 측정에서 제외할 수 있습니다.

```go
func BenchmarkCalculateWithResetTimer(b *testing.B) {
	op := "add"
	a := 10
	c := 20

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = Calculate(op, a, c)
	}
}
```

`b.ResetTimer()` 의미:

```text
이 줄 이전의 준비 시간은 벤치마크 측정에서 제외한다.
```

---

# 10단계. 메모리 할당까지 확인하기

```bash
go test -bench=. -benchmem
```

출력 예:

```text
BenchmarkCalculate/add-8     100000000    10.5 ns/op    0 B/op    0 allocs/op
```

| 항목 | 의미 |
|---|---|
| `ns/op` | 1회 실행당 걸린 시간 |
| `B/op` | 1회 실행당 사용한 메모리 |
| `allocs/op` | 1회 실행당 메모리 할당 횟수 |

좋은 성능의 함수는 보통 다음처럼 나옵니다.

```text
0 B/op    0 allocs/op
```

---

# 11단계. 특정 벤치마크만 실행하기

```bash
go test -bench=BenchmarkAdd
```

이름에 `Calculate`가 들어간 벤치마크만 실행:

```bash
go test -bench=Calculate
```

일반 테스트는 생략하고 벤치마크만 실행:

```bash
go test -run=^$ -bench=.
```

---

# 12단계. 최종 전체 코드

## calculator.go

```go
package calculator

import "fmt"

func Add(a int, b int) int {
	return a + b
}

func Mul(a int, b int) int {
	return a * b
}

func Calculate(op string, a int, b int) (int, error) {
	switch op {
	case "add":
		return a + b, nil
	case "mul":
		return a * b, nil
	default:
		return 0, fmt.Errorf("지원하지 않는 연산입니다: %s", op)
	}
}
```

## calculator_test.go

```go
package calculator

import "testing"

func TestAdd(t *testing.T) {
	result := Add(10, 20)

	if result != 30 {
		t.Errorf("원하는 값: %d, 실제 값: %d", 30, result)
	}
}

func TestMul(t *testing.T) {
	result := Mul(10, 20)

	if result != 200 {
		t.Errorf("원하는 값: %d, 실제 값: %d", 200, result)
	}
}

func TestCalculate(t *testing.T) {
	result, err := Calculate("add", 10, 20)

	if err != nil {
		t.Errorf("에러가 발생하면 안 됩니다: %v", err)
	}

	if result != 30 {
		t.Errorf("원하는 값: %d, 실제 값: %d", 30, result)
	}
}
```

## calculator_benchmark_test.go

```go
package calculator

import "testing"

func BenchmarkAdd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Add(10, 20)
	}
}

func BenchmarkMul(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Mul(10, 20)
	}
}

func BenchmarkCalculateAdd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = Calculate("add", 10, 20)
	}
}

func BenchmarkCalculate(b *testing.B) {
	benchmarks := []struct {
		name string
		op   string
		a    int
		b    int
	}{
		{"add", "add", 10, 20},
		{"mul", "mul", 10, 20},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = Calculate(bm.op, bm.a, bm.b)
			}
		})
	}
}

func BenchmarkCalculateWithResetTimer(b *testing.B) {
	op := "add"
	a := 10
	c := 20

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = Calculate(op, a, c)
	}
}
```

---

# 13단계. 실행 명령 정리

| 명령 | 의미 |
|---|---|
| `go test` | 일반 테스트 실행 |
| `go test -v` | 일반 테스트 자세히 보기 |
| `go test -bench=.` | 모든 벤치마크 실행 |
| `go test -bench=. -v` | 벤치마크 자세히 보기 |
| `go test -bench=. -benchmem` | 메모리 할당 정보 포함 |
| `go test -bench=BenchmarkAdd` | 특정 벤치마크만 실행 |
| `go test -run=^$ -bench=.` | 일반 테스트 생략, 벤치마크만 실행 |

---

# 14단계. 실습 체크리스트

- [ ] `go mod init go-benchmark-practice` 실행
- [ ] `calculator.go` 작성
- [ ] `calculator_test.go` 작성
- [ ] `go test -v` 실행
- [ ] `calculator_benchmark_test.go` 작성
- [ ] `go test -bench=.` 실행
- [ ] `go test -bench=. -benchmem` 실행
- [ ] `BenchmarkAdd`와 `BenchmarkCalculateAdd` 결과 비교
- [ ] `b.ResetTimer()` 사용
- [ ] `go test -run=^$ -bench=.` 실행

---

# 15단계. 핵심 요약

Go 벤치마크 함수 기본 형태:

```go
func BenchmarkName(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// 성능을 측정할 코드
	}
}
```

핵심 개념:

| 개념 | 설명 |
|---|---|
| `BenchmarkXxx` | 벤치마크 함수 이름 규칙 |
| `*testing.B` | 벤치마크용 객체 |
| `b.N` | Go가 자동으로 정하는 반복 횟수 |
| `b.Run()` | 서브 벤치마크 실행 |
| `b.ResetTimer()` | 준비 시간 제외 |
| `-bench=.` | 모든 벤치마크 실행 |
| `-benchmem` | 메모리 할당 정보 출력 |

가장 많이 쓰는 명령:

```bash
go test -bench=. -benchmem
```
