# Go 언어 프로그래밍 4일차 — 동시성 심화: Context, Worker Pool, Pipeline

> **대상**: 3일차까지 마친 C 개발자
> **목표**: `select` 심화, 타임아웃·취소 메커니즘, `context` 패키지를 마스터하고, **Worker Pool**과 **Pipeline** 같은 실전 동시성 패턴을 직접 구현할 수 있다.
> **준비물**: 3일차 환경, 멀티코어 시스템

---

## 📋 4일차 시간표

| 교시 | 주제 | 핵심 내용 |
|---|---|---|
| 1교시 | `select` 심화 | 다중 채널, 동적 선택, 패턴 |
| 2교시 | Timeout과 취소 | `time.After`, `Ticker`, 타임아웃 패턴 |
| 3교시 | Context 패키지 I | 취소 전파, `WithCancel` |
| 4교시 | Context 패키지 II | `WithDeadline`, `WithValue`, 모범 사례 |
| 5교시 | Worker Pool 패턴 | 정적/동적 풀, 부하 분산 |
| 6교시 | Pipeline 패턴 심화 | 다단계 변환, 백프레셔, 에러 전파 |
| 7교시 | 실습 | Worker Pool 구현 |
| 8교시 | 실습 | 파이프라인 기반 데이터 처리 |

---

# 1교시. `select` 심화 — 다중 채널 자유자재로

## 1.1 `select` 복습 + 깊이 들어가기

3일차에서 `select`의 기본은 다뤘습니다. 이번엔 **실전에서 어떻게 활용하는가**에 집중합니다.

```go
select {
case v := <-ch1:
    // ch1 수신
case ch2 <- x:
    // ch2 송신
case <-time.After(time.Second):
    // 타임아웃
default:
    // 위 어느 것도 준비 안 됨
}
```

### 동작 규칙 재정리

| 상황 | 동작 |
|---|---|
| 준비된 case 1개 | 그것 실행 |
| 준비된 case 여러 개 | **무작위** 선택 (편향 방지) |
| 준비된 case 없음 + default 있음 | default 실행 (non-blocking) |
| 준비된 case 없음 + default 없음 | 블록 |
| 모든 case가 nil 채널 + default 없음 | **영원히 블록** (데드락) |

### 🔑 핵심 — `select`는 C의 `poll`/`epoll`과 본질이 같다

```c
// C — poll로 다중 파일 디스크립터 감시
struct pollfd fds[3];
fds[0].fd = sock_a; fds[0].events = POLLIN;
fds[1].fd = sock_b; fds[1].events = POLLIN;
fds[2].fd = sock_c; fds[2].events = POLLIN;
poll(fds, 3, timeout);
```

```go
// Go — select로 다중 채널 감시
select {
case <-chA:
case <-chB:
case <-chC:
case <-time.After(timeout):
}
```

차이점은 Go의 `select`가 **언어 차원에서 통합**되어 있고, **임의의 채널 연산**(송신/수신/타임아웃)을 섞을 수 있다는 점입니다.

## 1.2 `default`를 활용한 Non-blocking 통신

블록하지 않고 채널 상태를 확인하는 데 씁니다.

### Non-blocking 송신

```go
select {
case ch <- value:
    fmt.Println("송신 성공")
default:
    fmt.Println("채널 가득 참 — 송신 포기")
}
```

### Non-blocking 수신

```go
select {
case v := <-ch:
    fmt.Println("받음:", v)
default:
    fmt.Println("받을 게 없음")
}
```

### 활용 예 — 드롭 가능한 알림

```go
// 알림 채널 - 가득 차면 그냥 버림 (overload 방지)
notify := make(chan struct{}, 1)

func notifyOnce() {
    select {
    case notify <- struct{}{}:
        // 알림 전송됨
    default:
        // 이미 알림이 큐에 있음 - 무시
    }
}
```

이 패턴은 **이벤트 결합(coalescing)**이라 불립니다. C에서는 직접 구현하기 번거롭지만 Go는 한 블록으로 끝납니다.

## 1.3 nil 채널 활용 — case를 동적으로 비활성화

`select`에서 **nil 채널은 영원히 블록**됩니다. 이걸 역으로 이용하면 case를 켜고 끌 수 있습니다.

```go
var inCh = make(chan int)
var outCh chan int  // nil 시작

// 어떤 조건에서 outCh를 활성화
if needForward {
    outCh = make(chan int)
}

select {
case v := <-inCh:
    fmt.Println("input:", v)
case outCh <- 42:  // outCh가 nil이면 이 case는 영원히 안 뽑힘
    fmt.Println("output sent")
}
```

### 실전 패턴 — 송수신 동시 처리

생산자가 데이터를 만들고 동시에 보내는 경우를 생각해봅시다.

```go
func producer(out chan<- int) {
    var pending []int
    var nextValue int

    for {
        var send chan<- int  // nil 시작
        var first int

        // pending이 있을 때만 send를 활성화
        if len(pending) > 0 {
            send = out
            first = pending[0]
        }

        select {
        case send <- first:
            // 첫 항목 송신 성공 → 큐에서 제거
            pending = pending[1:]
        case <-time.After(100 * time.Millisecond):
            // 일정 주기마다 데이터 생성
            pending = append(pending, nextValue)
            nextValue++
        }
    }
}
```

**핵심 트릭**: `pending`이 비었으면 `send`가 nil → 송신 case 자동 비활성. 데이터가 있을 때만 송신 시도.

C에서 이런 동작을 만들려면 별도 플래그나 조건 변수가 필요했지만, Go는 nil 채널 하나로 끝납니다.

## 1.4 `for-select` 패턴 — 가장 흔한 구조

워커 고루틴은 보통 이런 형태를 가집니다.

```go
func worker(jobs <-chan Job, done <-chan struct{}) {
    for {
        select {
        case job := <-jobs:
            process(job)
        case <-done:
            return
        }
    }
}
```

- **`for`로 계속 반복**, **`select`로 다음 할 일 결정**
- 종료 신호(`done`)를 항상 포함 — 누수 방지

### 안티 패턴 ① — `done` 누락

```go
// ❌ 종료할 방법이 없음
for {
    select {
    case job := <-jobs:
        process(job)
    }
}
```

### 안티 패턴 ② — `default`로 바쁜 루프

```go
// ❌ CPU 100% 점유
for {
    select {
    case job := <-jobs:
        process(job)
    default:
        // 매 사이클마다 여기로 — CPU 폭주
    }
}
```

`default`는 **정말 비동기적으로 확인만 하고 다른 일을 할 때**만 씁니다.

## 1.5 `reflect.Select` — 동적 select (참고)

`select`의 case 개수는 **컴파일 시점에 고정**됩니다. 런타임에 가변 개수의 채널을 다루려면 `reflect.Select`를 씁니다.

```go
import "reflect"

func dynamicSelect(channels []chan int) {
    cases := make([]reflect.SelectCase, len(channels))
    for i, ch := range channels {
        cases[i] = reflect.SelectCase{
            Dir:  reflect.SelectRecv,
            Chan: reflect.ValueOf(ch),
        }
    }
    chosen, value, ok := reflect.Select(cases)
    fmt.Printf("case %d, value %v, ok %v\n", chosen, value, ok)
}
```

**주의**: 매우 느립니다(일반 `select`의 10~100배). 정말 필요할 때만 사용하세요. 대안으로 fan-in 패턴(여러 채널 → 하나로 합치기)을 더 자주 씁니다.

## 1.6 🧪 실습 코드: `select_patterns.go`

```go
package main

import (
    "fmt"
    "time"
)

// 패턴 1: 이벤트 결합 (Coalescing)
func coalesceDemo() {
    fmt.Println("\n=== Pattern 1: Event Coalescing ===")
    notify := make(chan struct{}, 1)

    // 빠르게 10번 트리거 — 하나만 큐에 남음
    for i := 0; i < 10; i++ {
        select {
        case notify <- struct{}{}:
            fmt.Println("알림 큐잉됨")
        default:
            fmt.Println("이미 대기 중 — 드롭")
        }
    }

    // 소비
    <-notify
    fmt.Println("알림 소비")
}

// 패턴 2: nil 채널로 case 토글
func toggleDemo() {
    fmt.Println("\n=== Pattern 2: nil Channel Toggle ===")
    in := make(chan int)
    var out chan int  // nil

    go func() {
        for i := 1; i <= 5; i++ {
            in <- i
        }
        close(in)
    }()

    var pending int
    hasPending := false

    for {
        var send chan<- int
        if hasPending {
            send = out
            // 실제 소비자가 없으니 시뮬레이션
            // 여기서는 그냥 출력
            fmt.Printf("송신 준비: %d\n", pending)
            hasPending = false
            continue
        }

        select {
        case v, ok := <-in:
            if !ok {
                return
            }
            pending = v
            hasPending = true
        case send <- pending:
            // 실제론 여기서 소비자에게 전달됨
        }
    }
}

// 패턴 3: for-select with shutdown
func workerDemo() {
    fmt.Println("\n=== Pattern 3: for-select Worker ===")
    jobs := make(chan int)
    done := make(chan struct{})

    go func() {
        for {
            select {
            case job, ok := <-jobs:
                if !ok {
                    fmt.Println("워커: 채널 닫힘, 종료")
                    return
                }
                fmt.Printf("워커: 작업 %d 처리\n", job)
            case <-done:
                fmt.Println("워커: 강제 종료 신호")
                return
            }
        }
    }()

    for i := 1; i <= 3; i++ {
        jobs <- i
    }
    time.Sleep(100 * time.Millisecond)
    close(done)
    time.Sleep(100 * time.Millisecond)
}

func main() {
    coalesceDemo()
    workerDemo()
}
```

### ✅ 1교시 체크포인트

- [ ] `select`의 무작위 선택 동작을 이해했는가?
- [ ] non-blocking 송수신을 작성할 수 있는가?
- [ ] nil 채널로 case를 비활성화하는 트릭을 이해했는가?
- [ ] `for-select` 안티 패턴을 피할 수 있는가?

---

# 2교시. Timeout과 취소 — `time` 패키지 활용

## 2.1 왜 타임아웃이 중요한가?

분산 시스템에서 **무한정 기다리는 코드는 곧 버그**입니다.

- HTTP 요청이 응답 없이 멈춤 → 연결 누수
- 데이터베이스 쿼리가 안 끝남 → 연결 풀 고갈
- 파일 I/O가 멈춤 → 디스크 문제

C에서는 `select()`/`poll()`의 timeout 인자, `alarm()` 시그널, `SO_RCVTIMEO` 소켓 옵션 등 여러 방법이 흩어져 있었습니다. **Go는 `time` 패키지 + `select`로 일관되게** 해결합니다.

## 2.2 `time.After` — 일회성 타임아웃

```go
import "time"

select {
case v := <-ch:
    fmt.Println("받음:", v)
case <-time.After(2 * time.Second):
    fmt.Println("타임아웃!")
}
```

`time.After(d)`는 **`d` 시간 후에 값이 들어오는 채널**을 반환합니다.

### 동작 원리

```go
// 내부적으로 대략 이런 의미
func After(d Duration) <-chan Time {
    ch := make(chan Time, 1)
    AfterFunc(d, func() { ch <- time.Now() })
    return ch
}
```

새 타이머가 만들어지므로 **반복 사용 시 GC 부담**이 있습니다. 짧은 주기로 자주 호출하면 `time.Timer`나 `time.Ticker`를 직접 만드는 게 효율적입니다.

## 2.3 `time.Timer` — 재사용 가능한 타이머

```go
timer := time.NewTimer(2 * time.Second)
defer timer.Stop()  // 리소스 정리

select {
case v := <-ch:
    fmt.Println("받음:", v)
    if !timer.Stop() {
        <-timer.C  // 이미 발화했으면 채널 비우기
    }
case <-timer.C:
    fmt.Println("타임아웃!")
}
```

### `Reset`으로 재사용

```go
timer := time.NewTimer(time.Second)
defer timer.Stop()

for {
    select {
    case <-ch:
        if !timer.Stop() {
            <-timer.C
        }
        timer.Reset(time.Second)  // 다시 1초 카운트
    case <-timer.C:
        fmt.Println("idle 1초")
        timer.Reset(time.Second)
    }
}
```

### 🔥 주의 — `Stop`과 `Reset`의 정확한 사용

`timer.Stop()`은:
- 타이머가 아직 발화 안 했으면 `true` 반환, 타이머 취소
- 이미 발화했거나 정지된 경우 `false` 반환

발화 후 채널에 값이 남아있을 수 있으므로, **`Stop`이 `false`면 채널을 비워야** 합니다. 이건 Go 1.23부터 단순해졌지만 호환성을 위해 패턴을 알아두세요.

## 2.4 `time.Ticker` — 주기적 신호

```go
ticker := time.NewTicker(500 * time.Millisecond)
defer ticker.Stop()

for i := 0; i < 5; i++ {
    <-ticker.C
    fmt.Println("틱:", time.Now())
}
```

C에서 `alarm()` + 시그널 핸들러로 하던 일을 채널 수신으로 처리할 수 있습니다.

### 자주 쓰는 패턴 — 주기 작업 + 종료 신호

```go
func periodicWorker(done <-chan struct{}) {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            doWork()
        case <-done:
            return
        }
    }
}
```

### ❌ 함정 — `Stop` 잊으면 누수

```go
// ❌ 잘못 - ticker가 영원히 동작 + 고루틴 누수
go func() {
    ticker := time.NewTicker(time.Second)
    for {
        <-ticker.C
        // ...
    }
}()
```

**반드시 `defer ticker.Stop()`**.

## 2.5 타임아웃 패턴 모음

### 패턴 ① — 함수 호출에 타임아웃 걸기

```go
func callWithTimeout(timeout time.Duration) (string, error) {
    result := make(chan string, 1)
    go func() {
        result <- slowOperation()
    }()

    select {
    case r := <-result:
        return r, nil
    case <-time.After(timeout):
        return "", fmt.Errorf("타임아웃 (%v 초과)", timeout)
    }
}
```

**주의**: 타임아웃이 나도 백그라운드 고루틴은 계속 실행됩니다(누수). `slowOperation`이 취소 가능해야 진짜 해결됩니다 → 3교시 `context`에서 다룹니다.

### 패턴 ② — 첫 응답만 받기 (Race)

여러 소스에서 가장 빠른 응답을 채택.

```go
func fastest(urls []string) string {
    ch := make(chan string, len(urls))
    for _, url := range urls {
        go func(u string) {
            ch <- fetch(u)
        }(u)
    }
    return <-ch  // 가장 빠른 하나만
}
```

### 패턴 ③ — Rate Limiting (속도 제한)

```go
// 1초에 5번 이상 호출 안 되게
limit := time.Tick(200 * time.Millisecond)

for req := range requests {
    <-limit  // 200ms 간격 보장
    process(req)
}
```

`time.Tick`은 `time.Ticker`의 채널만 반환하는 간편 함수입니다. **취소 불가**이므로 단기 스크립트에서만 사용하세요. 장기 실행은 `NewTicker`.

## 2.6 C 시그널 기반 타이머와의 비교

| 항목 | C `alarm()` + SIGALRM | Go `time.After` |
|---|---|---|
| 시그널 핸들러 작성 필요 | ✅ | ❌ |
| 비동기 안전 함수만 사용 가능 | ✅ (제약 많음) | ❌ (제약 없음) |
| 여러 타이머 동시 사용 | 복잡함 (큐 직접 관리) | 자유로움 |
| 정확도 | OS 의존 | nanosecond 수준 |
| 데이터 흐름 | 전역 변수 + 시그널 | 채널 (지역적) |
| 취소 | 별도 메커니즘 | `Stop()` |

Go의 방식이 압도적으로 깔끔합니다.

## 2.7 🧪 실습 코드: `timeout_patterns.go`

```go
package main

import (
    "fmt"
    "math/rand"
    "time"
)

// 느린 작업 시뮬레이션
func slowQuery(id int) string {
    time.Sleep(time.Duration(rand.Intn(3000)) * time.Millisecond)
    return fmt.Sprintf("결과-%d", id)
}

// 타임아웃이 있는 호출
func queryWithTimeout(id int, timeout time.Duration) (string, error) {
    ch := make(chan string, 1)
    go func() {
        ch <- slowQuery(id)
    }()

    select {
    case r := <-ch:
        return r, nil
    case <-time.After(timeout):
        return "", fmt.Errorf("타임아웃 (id=%d, %v)", id, timeout)
    }
}

// 주기적 모니터
func monitor(done <-chan struct{}) {
    ticker := time.NewTicker(500 * time.Millisecond)
    defer ticker.Stop()

    count := 0
    for {
        select {
        case <-ticker.C:
            count++
            fmt.Printf("모니터 틱 #%d\n", count)
        case <-done:
            fmt.Println("모니터 종료")
            return
        }
    }
}

func main() {
    rand.Seed(time.Now().UnixNano())

    // 1. 타임아웃 있는 호출들
    for i := 1; i <= 5; i++ {
        result, err := queryWithTimeout(i, 1500*time.Millisecond)
        if err != nil {
            fmt.Printf("[%d] 실패: %v\n", i, err)
        } else {
            fmt.Printf("[%d] 성공: %s\n", i, result)
        }
    }

    // 2. 주기적 모니터
    fmt.Println("\n--- Ticker 데모 ---")
    done := make(chan struct{})
    go monitor(done)
    time.Sleep(2 * time.Second)
    close(done)
    time.Sleep(100 * time.Millisecond)
}
```

### ✅ 2교시 체크포인트

- [ ] `time.After`로 타임아웃을 구현할 수 있는가?
- [ ] `Timer.Stop()`을 올바르게 사용할 수 있는가?
- [ ] `Ticker`로 주기 작업을 만들 수 있는가?
- [ ] `time.After`의 누수 문제를 이해했는가?

---

# 3교시. Context 패키지 I — 취소 전파

## 3.1 왜 Context인가?

2교시에서 타임아웃 패턴을 봤지만, 한 가지 한계가 있었습니다.

```go
func callWithTimeout() (string, error) {
    ch := make(chan string, 1)
    go func() {
        ch <- slowOperation()  // 이 고루틴은 타임아웃 후에도 계속 실행됨
    }()
    select {
    case r := <-ch:
        return r, nil
    case <-time.After(timeout):
        return "", errors.New("timeout")
    }
}
```

**핵심 문제**: 호출자가 포기해도 `slowOperation`은 계속 자원을 씁니다. 만약 `slowOperation`이 다시 다른 함수들을 호출한다면, **취소 신호가 그 깊은 곳까지 전파되지 않습니다.**

### 진짜 필요한 것 — 호출 사슬을 따라 흐르는 취소 신호

```text
HTTP 핸들러
   ↓ (요청 타임아웃 5초)
서비스 함수
   ↓
DB 쿼리
   ↓
네트워크 호출

만약 클라이언트가 연결을 끊으면 → 모든 계층이 즉시 작업 중단해야 함
```

이 문제를 표준화한 것이 **`context.Context`** 입니다.

> Go 1.7부터 표준 라이브러리에 편입되었고, 이제는 거의 모든 표준 함수의 첫 인자가 `ctx`입니다.

## 3.2 `Context` 인터페이스

```go
type Context interface {
    Deadline() (deadline time.Time, ok bool)
    Done() <-chan struct{}
    Err() error
    Value(key any) any
}
```

| 메서드 | 역할 |
|---|---|
| `Done()` | 취소 신호 채널 — 닫히면 취소된 것 |
| `Err()` | 취소 사유 (`Canceled`, `DeadlineExceeded`) |
| `Deadline()` | 데드라인 시각 (있다면) |
| `Value()` | 요청 스코프 값 조회 |

**가장 자주 쓰는 메서드는 `Done()`** 입니다. 채널을 받아서 `select`로 감시하는 패턴이 핵심입니다.

## 3.3 두 가지 루트 Context

모든 `Context`는 루트에서 파생됩니다.

```go
ctx := context.Background()  // 일반적인 루트
ctx := context.TODO()        // 아직 어떤 ctx를 쓸지 모를 때 임시
```

| | `Background()` | `TODO()` |
|---|---|---|
| 용도 | 정상 진입점 (main, init, 테스트) | "나중에 정하겠다"는 표시 |
| 동작 차이 | 없음 | 없음 (의도 표시일 뿐) |

코드 리뷰에서 `context.TODO()`를 보면 "여기는 적절한 ctx를 받아오게 고쳐야 한다"는 의미로 읽힙니다.

## 3.4 `WithCancel` — 수동 취소

```go
ctx, cancel := context.WithCancel(parent)
defer cancel()  // 누수 방지 — 반드시!

go worker(ctx)

// 나중에...
cancel()  // worker가 ctx.Done()을 받음
```

### 동작 흐름

```text
parent ─────┐
            ↓
        WithCancel
            ↓
       child ctx
            │
            ↓
   worker가 ctx.Done() 감시
            │
            ↓
   cancel() 호출 시
            ↓
   ctx.Done() 채널이 닫힘
            ↓
   worker 빠져나옴
```

### Worker 측 코드

```go
func worker(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            fmt.Println("취소됨:", ctx.Err())
            return
        case <-time.After(time.Second):
            fmt.Println("작업 중...")
        }
    }
}
```

`ctx.Err()`은 `context.Canceled` (수동 취소) 또는 `context.DeadlineExceeded` (시간 초과)를 반환합니다.

## 3.5 전파 — Context 트리

`WithCancel`은 새 ctx를 만들고, 부모 ctx와 연결합니다. **부모가 취소되면 자식도 자동 취소**됩니다.

```go
root := context.Background()

ctxA, cancelA := context.WithCancel(root)
defer cancelA()

ctxB, cancelB := context.WithCancel(ctxA)  // ctxA의 자식
defer cancelB()

ctxC, cancelC := context.WithCancel(ctxB)  // ctxB의 자식
defer cancelC()

// cancelA() 호출 시:
// - ctxA.Done() 닫힘
// - ctxB.Done() 닫힘 (자동 전파)
// - ctxC.Done() 닫힘 (자동 전파)
```

이게 바로 "**호출 사슬을 따라 흐르는 취소 신호**"입니다.

## 3.6 함수 시그니처 컨벤션

**ctx는 함수의 첫 번째 인자**라는 강력한 컨벤션이 있습니다.

```go
// ✅ 표준 컨벤션
func DoSomething(ctx context.Context, arg1 string, arg2 int) error

// ❌ 비추천
func DoSomething(arg1 string, ctx context.Context) error
```

표준 라이브러리(`net/http`, `database/sql`, ...)가 모두 이 컨벤션을 따릅니다.

## 3.7 cancel 함수를 잊으면? — 누수

```go
ctx, cancel := context.WithCancel(parent)
// defer cancel() 빠뜨림!

// 작업이 끝나도 ctx는 살아있음
// → 부모가 죽을 때까지 자원 점유
```

**`go vet`이 이런 실수를 잡습니다**:

```bash
go vet ./...
# context not released ...
```

규칙: **`cancel`은 반드시 호출**, 가장 안전한 건 `defer cancel()`.

## 3.8 🧪 실습 코드: `context_cancel.go`

```go
package main

import (
    "context"
    "fmt"
    "time"
)

func worker(ctx context.Context, id int) {
    fmt.Printf("[worker %d] 시작\n", id)
    for {
        select {
        case <-ctx.Done():
            fmt.Printf("[worker %d] 중단: %v\n", id, ctx.Err())
            return
        case <-time.After(300 * time.Millisecond):
            fmt.Printf("[worker %d] 진행 중...\n", id)
        }
    }
}

// 자식 작업이 또 자식을 만드는 경우 — 전파 확인
func parentJob(ctx context.Context) {
    for i := 1; i <= 3; i++ {
        go worker(ctx, i)
    }

    <-ctx.Done()
    fmt.Println("[부모] 중단:", ctx.Err())
}

func main() {
    ctx, cancel := context.WithCancel(context.Background())

    go parentJob(ctx)

    // 1초 후 모든 작업 취소
    time.Sleep(1 * time.Second)
    fmt.Println("--- cancel() 호출 ---")
    cancel()

    // 정리 시간
    time.Sleep(500 * time.Millisecond)
    fmt.Println("--- 메인 종료 ---")
}
```

실행해서 `cancel()` 호출 시점에 **모든 워커가 즉시 멈추는** 모습을 확인하세요.

### ✅ 3교시 체크포인트

- [ ] `Context`의 4가지 메서드 역할을 설명할 수 있는가?
- [ ] `WithCancel`을 사용해 취소 가능한 작업을 만들 수 있는가?
- [ ] 부모 ctx 취소가 자식에게 전파됨을 이해했는가?
- [ ] `defer cancel()`을 반드시 호출해야 하는 이유를 알고 있는가?

---

# 4교시. Context 패키지 II — Deadline, Value, 모범 사례

## 4.1 `WithDeadline` / `WithTimeout` — 시간 기반 자동 취소

`WithTimeout`은 가장 자주 쓰입니다.

```go
ctx, cancel := context.WithTimeout(parent, 2*time.Second)
defer cancel()  // 일찍 끝나도 정리

// 2초 후 자동으로 ctx.Done() 닫힘
```

`WithDeadline`은 특정 시각을 지정합니다.

```go
deadline := time.Now().Add(5 * time.Second)
ctx, cancel := context.WithDeadline(parent, deadline)
defer cancel()
```

`WithTimeout(parent, d)` ≈ `WithDeadline(parent, time.Now().Add(d))`.

### 실전 예 — HTTP 요청 타임아웃

```go
import (
    "context"
    "io"
    "net/http"
    "time"
)

func fetch(url string, timeout time.Duration) ([]byte, error) {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err  // 타임아웃이면 ctx.Err() = DeadlineExceeded
    }
    defer resp.Body.Close()

    return io.ReadAll(resp.Body)
}
```

C에서 HTTP 클라이언트에 타임아웃 거는 데 `setsockopt`, `select`, 별도 시그널 처리 등이 필요했습니다. Go는 ctx 하나로 끝납니다.

## 4.2 더 짧은 데드라인이 우선

```go
parent, _ := context.WithTimeout(context.Background(), 5*time.Second)
child, _ := context.WithTimeout(parent, 10*time.Second)  // 부모보다 김

// child의 실제 데드라인은 5초 (부모 제한이 더 짧음)
```

부모가 먼저 만료되면 자식도 자동 취소됩니다. **자식의 데드라인은 부모를 넘어설 수 없습니다.**

## 4.3 `WithValue` — 요청 스코프 데이터 전달

호출 사슬을 따라 값을 전달할 수 있습니다.

```go
ctx := context.WithValue(parent, "requestID", "abc-123")

// 깊은 함수에서
reqID := ctx.Value("requestID").(string)
fmt.Println("Request ID:", reqID)
```

### 🔥 주의 — 매우 제한된 용도로만 사용

**`WithValue`는 안티 패턴이 되기 쉽습니다.** 다음 경우에만 사용:

✅ **적절한 사용**:
- 추적용 ID (`requestID`, `traceID`)
- 인증 정보 (`userID`, `token`)
- 로깅 메타데이터

❌ **부적절한 사용**:
- 함수 매개변수 대용 (명시적 인자가 항상 우선)
- 비즈니스 로직에 필요한 핵심 데이터
- 옵션 설정값

### 키 충돌 방지 — 사용자 정의 타입

```go
// 문자열 키는 다른 패키지와 충돌 위험!
ctx := context.WithValue(ctx, "user", u)  // ❌

// 사용자 정의 타입으로 키 만들기
type contextKey string

const userKey contextKey = "user"

ctx := context.WithValue(ctx, userKey, u)  // ✅
u := ctx.Value(userKey).(*User)
```

타입까지 일치해야 하므로 다른 패키지의 같은 이름 키와 충돌하지 않습니다.

### 패턴 — Getter/Setter 함수 노출

```go
package auth

type contextKey string

const userKey contextKey = "auth.user"

func WithUser(ctx context.Context, u *User) context.Context {
    return context.WithValue(ctx, userKey, u)
}

func UserFrom(ctx context.Context) (*User, bool) {
    u, ok := ctx.Value(userKey).(*User)
    return u, ok
}
```

호출 측은 키를 모르고도 안전하게 접근:

```go
ctx = auth.WithUser(ctx, currentUser)
// ...
if u, ok := auth.UserFrom(ctx); ok {
    fmt.Println(u.Name)
}
```

## 4.4 Context 모범 사례 정리

### ✅ DO

1. **ctx는 첫 번째 매개변수로**
2. **`cancel`은 반드시 호출** (defer 권장)
3. **고루틴에 ctx 전달** — 취소 가능하게
4. **타임아웃 설정** — 외부 호출은 무조건
5. **루트는 `Background()`** — main/test/init에서

### ❌ DON'T

1. **구조체 필드로 ctx 저장 금지** — ctx는 함수 인자로만
2. **nil ctx 전달 금지** — 의심스러우면 `context.TODO()`
3. **`WithValue`로 옵션 전달 금지** — 명시적 매개변수 사용
4. **ctx 없이 외부 호출 금지** — HTTP, DB 등은 항상 ctx 받음

## 4.5 ctx-aware 함수 작성하기

기존에 ctx를 받지 않던 함수를 ctx-aware로 만들어봅시다.

### Before (ctx 무관)

```go
func processItems(items []Item) error {
    for _, item := range items {
        if err := process(item); err != nil {
            return err
        }
    }
    return nil
}
```

### After (ctx 사용)

```go
func processItems(ctx context.Context, items []Item) error {
    for _, item := range items {
        // 매 반복마다 취소 확인
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }

        if err := process(ctx, item); err != nil {
            return err
        }
    }
    return nil
}
```

**관용구**: `select` + `default`로 **non-blocking 취소 확인**.

만약 `process` 자체가 오래 걸리는 작업이면, 그 안에서도 ctx를 감시해야 합니다.

## 4.6 C 시그널과 비교

| 항목 | C SIGTERM/SIGINT | Go Context |
|---|---|---|
| 전달 단위 | 프로세스 | 함수 호출 사슬 |
| 핸들러 위치 | 시그널 핸들러 (전역) | 각 함수 내부 |
| 다중 채널 가능 | 한 번에 하나 | 트리 구조로 다중 |
| 메타데이터 | 시그널 번호만 | 임의의 값 (`WithValue`) |
| 깊은 호출 통과 | 어려움 | 자연스러움 (인자 전달) |

C에서 `pthread`별로 시그널을 다루는 건 복잡한 일이었지만, Go의 Context는 **단순히 함수 인자**로 모든 게 해결됩니다.

## 4.7 🧪 실습 코드: `context_advanced.go`

```go
package main

import (
    "context"
    "errors"
    "fmt"
    "time"
)

type ctxKey string

const (
    requestIDKey ctxKey = "request-id"
    userIDKey    ctxKey = "user-id"
)

// 헬퍼 함수들
func WithRequestID(ctx context.Context, id string) context.Context {
    return context.WithValue(ctx, requestIDKey, id)
}

func RequestIDFrom(ctx context.Context) string {
    if v, ok := ctx.Value(requestIDKey).(string); ok {
        return v
    }
    return "unknown"
}

// 깊은 단계의 함수 - ctx-aware
func dbQuery(ctx context.Context) error {
    reqID := RequestIDFrom(ctx)
    fmt.Printf("[%s] DB 쿼리 시작\n", reqID)

    select {
    case <-time.After(2 * time.Second):  // 느린 쿼리 시뮬레이션
        fmt.Printf("[%s] DB 쿼리 완료\n", reqID)
        return nil
    case <-ctx.Done():
        fmt.Printf("[%s] DB 쿼리 중단: %v\n", reqID, ctx.Err())
        return ctx.Err()
    }
}

// 중간 계층
func service(ctx context.Context) error {
    reqID := RequestIDFrom(ctx)
    fmt.Printf("[%s] 서비스 호출\n", reqID)
    return dbQuery(ctx)
}

// 핸들러 - 1초 데드라인 설정
func handler(ctx context.Context) error {
    ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
    defer cancel()

    return service(ctx)
}

func main() {
    // 요청 1: 정상 (느린 작업이지만 데드라인 짧음)
    ctx1 := WithRequestID(context.Background(), "req-001")
    if err := handler(ctx1); err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            fmt.Println("→ 응답: 504 Gateway Timeout")
        }
    }

    fmt.Println()

    // 요청 2: 수동 취소
    ctx2, cancel := context.WithCancel(WithRequestID(context.Background(), "req-002"))
    go func() {
        time.Sleep(300 * time.Millisecond)
        cancel()
    }()

    if err := handler(ctx2); err != nil {
        fmt.Printf("종료: %v\n", err)
    }
}
```

기대 출력:
```
[req-001] 서비스 호출
[req-001] DB 쿼리 시작
[req-001] DB 쿼리 중단: context deadline exceeded
→ 응답: 504 Gateway Timeout

[req-002] 서비스 호출
[req-002] DB 쿼리 시작
[req-002] DB 쿼리 중단: context canceled
종료: context canceled
```

### ✅ 4교시 체크포인트

- [ ] `WithTimeout`과 `WithDeadline`을 구분해 쓸 수 있는가?
- [ ] `WithValue`의 안티 패턴을 피할 수 있는가?
- [ ] `select` + `<-ctx.Done()`을 깊은 함수에 적용할 수 있는가?
- [ ] ctx의 5가지 모범 사례를 떠올릴 수 있는가?

---

# 5교시. Worker Pool 패턴

## 5.1 왜 워커 풀인가?

3일차에서 봤듯이 Go는 백만 개 고루틴도 만들 수 있습니다. 그렇다고 **무제한으로 만드는 게 좋은 건 아닙니다.**

- 외부 자원(DB 연결, 파일 디스크립터)에는 한계가 있음
- 너무 많은 동시 호출은 외부 시스템을 죽일 수 있음
- 부하 분산이 자연스럽지 않음

**Worker Pool**은 **고정된 수의 워커**가 작업 채널에서 일감을 가져가 처리하는 패턴입니다.

```text
                       ┌────────────┐
                  ┌────│ Worker 1   │
                  │    └────────────┘
                  │
   jobs ──────────┼────┌────────────┐
   channel        ├────│ Worker 2   │  →  results
                  │    └────────────┘     channel
                  │
                  │    ┌────────────┐
                  └────│ Worker 3   │
                       └────────────┘
```

C에서 `pthread_pool`을 직접 구현해본 분이라면, **채널과 고루틴으로 얼마나 단순해지는지** 보면 놀랄 겁니다.

## 5.2 기본형 — 고정 크기 풀

```go
type Job struct {
    ID   int
    Data string
}

type Result struct {
    JobID  int
    Output string
}

func worker(id int, jobs <-chan Job, results chan<- Result) {
    for job := range jobs {
        // 처리
        output := fmt.Sprintf("worker-%d processed job-%d", id, job.ID)
        time.Sleep(100 * time.Millisecond)
        results <- Result{JobID: job.ID, Output: output}
    }
}

func main() {
    const numWorkers = 5
    const numJobs = 20

    jobs := make(chan Job)
    results := make(chan Result, numJobs)

    // 워커 시작
    var wg sync.WaitGroup
    for i := 1; i <= numWorkers; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            worker(id, jobs, results)
        }(i)
    }

    // 작업 송신
    go func() {
        for j := 1; j <= numJobs; j++ {
            jobs <- Job{ID: j, Data: fmt.Sprintf("data-%d", j)}
        }
        close(jobs)
    }()

    // 모든 워커 종료 시 results도 닫기
    go func() {
        wg.Wait()
        close(results)
    }()

    // 결과 수집
    for r := range results {
        fmt.Println(r.Output)
    }
}
```

### 핵심 포인트

1. **`jobs` 채널 하나를 여러 워커가 공유** → 자동 부하 분산
2. **`close(jobs)`로 작업 종료 신호** → 워커들이 `range`에서 자연스럽게 빠져나옴
3. **`wg.Wait()` 후 `close(results)`** → 결과 채널도 깔끔하게 마감

## 5.3 Context로 전체 풀 취소

위 코드에 ctx 지원을 추가합니다.

```go
func worker(ctx context.Context, id int, jobs <-chan Job, results chan<- Result) {
    for {
        select {
        case <-ctx.Done():
            fmt.Printf("worker %d: 취소됨\n", id)
            return
        case job, ok := <-jobs:
            if !ok {
                return  // 채널 닫힘
            }
            // 결과 송신도 ctx 감시
            select {
            case results <- process(ctx, job):
            case <-ctx.Done():
                return
            }
        }
    }
}
```

호출 측:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// 워커들에게 ctx 전달
for i := 1; i <= numWorkers; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        worker(ctx, id, jobs, results)
    }(i)
}
```

5초 안에 끝나지 않으면 모든 워커가 동시에 중단됩니다.

## 5.4 결과 순서 보존 문제

위 패턴의 문제: **결과 순서가 작업 순서와 다를 수 있습니다.** 빠른 워커가 먼저 응답하기 때문입니다.

### 해결책 ① — 결과에 인덱스 포함, 호출 측이 정렬

```go
type Result struct {
    Index  int
    Output string
}

// 모든 결과를 모은 후
results := make([]Result, numJobs)
for r := range resultCh {
    results[r.Index] = r
}
```

### 해결책 ② — 각 작업마다 응답 채널을 동봉

```go
type Job struct {
    ID    int
    Data  string
    Reply chan Result  // 본인용 응답 채널
}

func main() {
    for j := 1; j <= numJobs; j++ {
        reply := make(chan Result, 1)
        jobs <- Job{ID: j, Reply: reply}
        results = append(results, <-reply)  // 순서 보존
    }
}
```

다만 이 방식은 호출 측이 **동기적으로 기다리므로** 동시성 효과가 줄어듭니다.

## 5.5 Bounded Concurrency — 세마포어 패턴

워커 풀의 변형. 풀을 명시적으로 만들지 않고, **동시 실행 개수만 제한**합니다.

```go
sem := make(chan struct{}, 10)  // 동시 10개 제한

var wg sync.WaitGroup
for _, item := range items {
    wg.Add(1)
    sem <- struct{}{}  // 슬롯 획득 (가득 차면 블록)
    go func(it Item) {
        defer wg.Done()
        defer func() { <-sem }()  // 슬롯 반환
        process(it)
    }(item)
}
wg.Wait()
```

**`chan struct{}`가 세마포어 역할**을 합니다. 매우 Go다운 패턴입니다.

| 패턴 | 적합한 상황 |
|---|---|
| Worker Pool (고정) | 작업이 끊임없이 들어옴 (서버) |
| Bounded Concurrency | 일회성 batch 작업 (스크립트) |

## 5.6 진행 상황 모니터링

장시간 실행되는 풀은 모니터링이 필요합니다.

```go
type PoolStats struct {
    Submitted int64
    Completed int64
    Failed    int64
}

var stats PoolStats

func worker(ctx context.Context, jobs <-chan Job) {
    for job := range jobs {
        atomic.AddInt64(&stats.Submitted, 1)
        if err := process(ctx, job); err != nil {
            atomic.AddInt64(&stats.Failed, 1)
        } else {
            atomic.AddInt64(&stats.Completed, 1)
        }
    }
}

// 모니터 고루틴
go func() {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()
    for {
        select {
        case <-ticker.C:
            fmt.Printf("진행: 완료=%d 실패=%d\n",
                atomic.LoadInt64(&stats.Completed),
                atomic.LoadInt64(&stats.Failed))
        case <-ctx.Done():
            return
        }
    }
}()
```

3일차의 atomic, ticker, ctx가 모두 등장합니다.

## 5.7 🧪 실습 코드: `worker_pool_basic.go`

```go
package main

import (
    "context"
    "fmt"
    "sync"
    "sync/atomic"
    "time"
)

type Task struct {
    ID int
}

type Outcome struct {
    TaskID int
    Result string
    Err    error
}

func process(ctx context.Context, t Task) Outcome {
    // 처리 시간 시뮬레이션
    select {
    case <-time.After(200 * time.Millisecond):
        return Outcome{TaskID: t.ID, Result: fmt.Sprintf("done-%d", t.ID)}
    case <-ctx.Done():
        return Outcome{TaskID: t.ID, Err: ctx.Err()}
    }
}

func worker(ctx context.Context, id int, tasks <-chan Task, results chan<- Outcome) {
    for {
        select {
        case <-ctx.Done():
            return
        case t, ok := <-tasks:
            if !ok {
                return
            }
            o := process(ctx, t)
            select {
            case results <- o:
            case <-ctx.Done():
                return
            }
        }
    }
}

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    const (
        numWorkers = 5
        numTasks   = 30
    )

    tasks := make(chan Task)
    results := make(chan Outcome, numTasks)

    var workersWG sync.WaitGroup
    for i := 1; i <= numWorkers; i++ {
        workersWG.Add(1)
        go func(id int) {
            defer workersWG.Done()
            worker(ctx, id, tasks, results)
        }(i)
    }

    // 작업 송신
    go func() {
        defer close(tasks)
        for i := 1; i <= numTasks; i++ {
            select {
            case tasks <- Task{ID: i}:
            case <-ctx.Done():
                return
            }
        }
    }()

    // 워커 모두 종료 시 결과 채널 닫기
    go func() {
        workersWG.Wait()
        close(results)
    }()

    // 결과 집계
    var (
        completed int64
        failed    int64
    )
    for o := range results {
        if o.Err != nil {
            atomic.AddInt64(&failed, 1)
        } else {
            atomic.AddInt64(&completed, 1)
        }
    }

    fmt.Printf("최종: 완료=%d 실패=%d\n", completed, failed)
}
```

### ✅ 5교시 체크포인트

- [ ] 고정 크기 워커 풀을 채널로 구현할 수 있는가?
- [ ] Context로 풀 전체를 취소할 수 있는가?
- [ ] Bounded Concurrency 패턴(세마포어 채널)을 사용할 수 있는가?
- [ ] 결과 순서 보존 방법 2가지를 알고 있는가?

---

# 6교시. Pipeline 패턴 심화

## 6.1 Pipeline이란?

**파이프라인**은 데이터가 여러 단계를 거치며 변환되는 패턴입니다.

```text
generator → stage1 → stage2 → stage3 → consumer
```

각 단계는 **고루틴 + 채널**로 구현됩니다. UNIX의 파이프(`cat file | grep foo | sort | uniq`)와 같은 사고방식입니다.

```bash
# UNIX
cat data.txt | grep ERROR | awk '{print $3}' | sort | uniq -c
```

Go의 파이프라인도 똑같은 구조를 가집니다. 차이는 **각 단계가 메모리 내에서 고루틴으로 실행**된다는 점입니다.

## 6.2 기본형 — 3단계 파이프라인

```go
// Stage 1: 정수 생성
func generate(nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for _, n := range nums {
            out <- n
        }
    }()
    return out
}

// Stage 2: 제곱
func square(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for n := range in {
            out <- n * n
        }
    }()
    return out
}

// Stage 3: 합계 누적 출력
func runningSum(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        sum := 0
        for n := range in {
            sum += n
            out <- sum
        }
    }()
    return out
}

func main() {
    for v := range runningSum(square(generate(1, 2, 3, 4, 5))) {
        fmt.Println(v)
    }
    // 1, 5, 14, 30, 55
}
```

### 파이프라인의 미덕

1. **단순함** — 각 단계가 작고 독립적
2. **재사용** — `square`는 어떤 정수 스트림에도 적용 가능
3. **자연스러운 동시성** — 각 단계가 동시에 진행
4. **메모리 효율** — 한 번에 하나씩 흘러가므로 큰 데이터도 OK

## 6.3 Fan-out / Fan-in 심화

병목 단계를 여러 워커로 분산.

```text
generator → square (N개 워커) → merge → consumer
              ↓ ↓ ↓
              ↓ ↓ ↓
              fan-out
              ↓
              fan-in
```

```go
// Fan-out: 같은 입력 채널을 N개 워커가 소비
// Fan-in: N개 출력을 하나로 합침

func merge(channels ...<-chan int) <-chan int {
    out := make(chan int)
    var wg sync.WaitGroup

    output := func(c <-chan int) {
        defer wg.Done()
        for v := range c {
            out <- v
        }
    }

    wg.Add(len(channels))
    for _, c := range channels {
        go output(c)
    }

    go func() {
        wg.Wait()
        close(out)
    }()

    return out
}

func main() {
    in := generate(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)

    // Fan-out: 3개 워커가 동시에 square
    c1 := square(in)
    c2 := square(in)  // 같은 in을 공유 — 자동 부하 분산
    c3 := square(in)

    // Fan-in
    for v := range merge(c1, c2, c3) {
        fmt.Println(v)
    }
}
```

**여기서 중요**: `c1`, `c2`, `c3` 모두 같은 `in`을 받습니다. 채널은 **여러 수신자가 안전하게 공유 가능**하며, 각 값은 **딱 하나의 수신자만** 받습니다.

## 6.4 백프레셔 (Backpressure)

생산자가 소비자보다 빠를 때, **자연스럽게 속도가 맞춰지는 현상**.

```go
in := generate(1, 2, 3, ..., 1000000)  // 빠르게 생성
sq := square(in)                        // 채널이 가득 차면 generate가 블록!
```

각 단계의 채널이 unbuffered거나 작은 버퍼라면, 다음 단계가 느리면 이전 단계도 자동으로 멈춥니다. **명시적인 속도 제어 없이 부하 균형**이 이뤄집니다.

### 버퍼 크기의 영향

```go
out := make(chan int)        // 강한 결합 — 1대1 페이스
out := make(chan int, 100)   // 약간의 버퍼 — burst 흡수
out := make(chan int, 10000) // 큰 버퍼 — 거의 비동기
```

기본은 unbuffered나 작은 버퍼. **큰 버퍼는 메모리 폭증 위험**.

## 6.5 에러 전파

파이프라인 단계에서 에러가 발생하면 어떻게 처리할까요?

### 방법 ① — 결과 타입에 에러 포함

```go
type Result struct {
    Value int
    Err   error
}

func parse(in <-chan string) <-chan Result {
    out := make(chan Result)
    go func() {
        defer close(out)
        for s := range in {
            n, err := strconv.Atoi(s)
            out <- Result{Value: n, Err: err}
        }
    }()
    return out
}
```

다음 단계가 에러를 보고 분기 처리합니다.

### 방법 ② — 별도 에러 채널

```go
func parse(in <-chan string) (<-chan int, <-chan error) {
    out := make(chan int)
    errCh := make(chan error, 1)
    go func() {
        defer close(out)
        defer close(errCh)
        for s := range in {
            n, err := strconv.Atoi(s)
            if err != nil {
                errCh <- err
                return  // 또는 계속
            }
            out <- n
        }
    }()
    return out, errCh
}
```

복잡하므로 단순한 경우엔 방법 ①이 권장됩니다.

### 방법 ③ — Context로 빠른 종료

에러가 나면 **ctx를 cancel**하여 모든 단계 동시 중단.

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

stage1 := generate(ctx, items...)
stage2 := transform(ctx, stage1)
stage3 := process(ctx, stage2)

for r := range stage3 {
    if r.Err != nil {
        cancel()  // 모든 단계 동시 종료
        break
    }
}
```

## 6.6 파이프라인 + Context 통합 예제

```go
func gen(ctx context.Context, nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for _, n := range nums {
            select {
            case out <- n:
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}

func sq(ctx context.Context, in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for n := range in {
            select {
            case out <- n * n:
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}
```

**관용구**: 모든 송신에 `select { case out <- v: case <-ctx.Done(): return }`.

ctx 취소 시 **누수 없이 즉시 종료**됩니다.

## 6.7 🧪 실습 코드: `pipeline_demo.go`

```go
package main

import (
    "context"
    "fmt"
    "sync"
    "time"
)

// Stage 1: 정수 생성
func gen(ctx context.Context, n int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for i := 1; i <= n; i++ {
            select {
            case out <- i:
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}

// Stage 2: 제곱 (느린 작업)
func square(ctx context.Context, in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for n := range in {
            time.Sleep(50 * time.Millisecond)
            select {
            case out <- n * n:
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}

// Fan-in
func merge(ctx context.Context, chs ...<-chan int) <-chan int {
    out := make(chan int)
    var wg sync.WaitGroup
    wg.Add(len(chs))

    for _, c := range chs {
        go func(c <-chan int) {
            defer wg.Done()
            for v := range c {
                select {
                case out <- v:
                case <-ctx.Done():
                    return
                }
            }
        }(c)
    }

    go func() {
        wg.Wait()
        close(out)
    }()

    return out
}

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    // 입력
    input := gen(ctx, 20)

    // Fan-out: 4개 워커로 square 처리
    workers := make([]<-chan int, 4)
    for i := range workers {
        workers[i] = square(ctx, input)
    }

    // Fan-in
    out := merge(ctx, workers...)

    // 결과
    total := 0
    count := 0
    for v := range out {
        total += v
        count++
    }
    fmt.Printf("처리한 개수: %d, 합계: %d\n", count, total)

    if ctx.Err() != nil {
        fmt.Println("종료 사유:", ctx.Err())
    }
}
```

### ✅ 6교시 체크포인트

- [ ] 3단계 이상의 파이프라인을 작성할 수 있는가?
- [ ] Fan-out/Fan-in의 자동 부하 분산을 설명할 수 있는가?
- [ ] 백프레셔의 의미를 이해했는가?
- [ ] 파이프라인 단계에 ctx를 통합할 수 있는가?

---

# 7교시. 실습 — Worker Pool 구현

이번 실습은 **다운로드 + 처리** 작업을 처리하는 본격적인 Worker Pool을 만듭니다.

## 7.1 시나리오: 다중 URL 헬스 체크 도구

- 여러 URL을 입력받음
- N개 워커가 동시에 HTTP HEAD/GET 호출
- 응답 시간, 상태 코드, 에러를 결과로 수집
- Context로 전체 타임아웃 적용
- Ctrl+C로 graceful shutdown
- 진행 상황 실시간 출력

## 7.2 Step 1 — 프로젝트 셋업

```bash
mkdir -p ~/go-class/day4/healthcheck
cd ~/go-class/day4/healthcheck
go mod init healthcheck
mkdir -p cmd/healthcheck
mkdir -p internal/checker
```

```
healthcheck/
├── go.mod
├── Makefile
├── cmd/healthcheck/main.go
└── internal/checker/
    ├── checker.go
    └── pool.go
```

## 7.3 Step 2 — 타입 정의

`internal/checker/checker.go`:

```go
package checker

import (
    "context"
    "fmt"
    "net/http"
    "time"
)

// Target은 검사 대상
type Target struct {
    URL string
}

// Result는 검사 결과
type Result struct {
    URL        string
    StatusCode int
    Latency    time.Duration
    Err        error
}

func (r Result) String() string {
    if r.Err != nil {
        return fmt.Sprintf("❌ %s — 에러: %v", r.URL, r.Err)
    }
    return fmt.Sprintf("✅ %s — %d (%v)", r.URL, r.StatusCode, r.Latency)
}

// Check은 단일 URL을 검사한다
func Check(ctx context.Context, url string) Result {
    start := time.Now()

    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return Result{URL: url, Err: err}
    }

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return Result{URL: url, Err: err, Latency: time.Since(start)}
    }
    defer resp.Body.Close()

    return Result{
        URL:        url,
        StatusCode: resp.StatusCode,
        Latency:    time.Since(start),
    }
}
```

## 7.4 Step 3 — Worker Pool 구현

`internal/checker/pool.go`:

```go
package checker

import (
    "context"
    "sync"
    "sync/atomic"
)

// Pool은 워커 풀
type Pool struct {
    NumWorkers int
    targets    chan Target
    results    chan Result

    completed int64
    failed    int64
}

// NewPool은 새 풀을 만든다
func NewPool(numWorkers int) *Pool {
    return &Pool{
        NumWorkers: numWorkers,
        targets:    make(chan Target),
        results:    make(chan Result, numWorkers*2),
    }
}

// Results는 결과 채널을 반환
func (p *Pool) Results() <-chan Result {
    return p.results
}

// Stats는 현재까지 통계
func (p *Pool) Stats() (completed, failed int64) {
    return atomic.LoadInt64(&p.completed),
           atomic.LoadInt64(&p.failed)
}

// Submit은 검사 대상을 큐에 넣는다
func (p *Pool) Submit(ctx context.Context, t Target) error {
    select {
    case p.targets <- t:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}

// Run은 워커들을 시작한다. 호출자는 Submit 후 Close를 호출해야 한다.
func (p *Pool) Run(ctx context.Context) {
    var wg sync.WaitGroup

    for i := 0; i < p.NumWorkers; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            p.workerLoop(ctx, id)
        }(i)
    }

    // 모든 워커 종료 시 결과 채널 닫기
    go func() {
        wg.Wait()
        close(p.results)
    }()
}

// Close는 더 이상 작업이 없음을 알린다.
func (p *Pool) Close() {
    close(p.targets)
}

func (p *Pool) workerLoop(ctx context.Context, id int) {
    for {
        select {
        case <-ctx.Done():
            return
        case t, ok := <-p.targets:
            if !ok {
                return
            }
            r := Check(ctx, t.URL)

            if r.Err != nil {
                atomic.AddInt64(&p.failed, 1)
            } else {
                atomic.AddInt64(&p.completed, 1)
            }

            select {
            case p.results <- r:
            case <-ctx.Done():
                return
            }
        }
    }
}
```

## 7.5 Step 4 — main.go (CLI)

`cmd/healthcheck/main.go`:

```go
package main

import (
    "context"
    "flag"
    "fmt"
    "os"
    "os/signal"
    "sync"
    "syscall"
    "time"

    "healthcheck/internal/checker"
)

func main() {
    var (
        workers = flag.Int("w", 5, "동시 워커 수")
        timeout = flag.Duration("t", 30*time.Second, "전체 타임아웃")
    )
    flag.Parse()

    urls := flag.Args()
    if len(urls) == 0 {
        urls = []string{
            "https://go.dev",
            "https://github.com",
            "https://google.com",
            "https://example.com",
            "https://invalid-url-foo-bar-baz.test",  // 실패 사례
        }
    }

    // 시그널 핸들링 포함 ctx
    ctx, cancel := context.WithTimeout(context.Background(), *timeout)
    defer cancel()

    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
    go func() {
        <-sigs
        fmt.Println("\n중단 신호 — 정리 중...")
        cancel()
    }()

    pool := checker.NewPool(*workers)
    pool.Run(ctx)

    // 제출
    var submitWG sync.WaitGroup
    submitWG.Add(1)
    go func() {
        defer submitWG.Done()
        defer pool.Close()
        for _, u := range urls {
            if err := pool.Submit(ctx, checker.Target{URL: u}); err != nil {
                fmt.Fprintln(os.Stderr, "제출 실패:", err)
                return
            }
        }
    }()

    // 진행 상황 모니터
    monitorDone := make(chan struct{})
    go monitor(ctx, pool, monitorDone)

    // 결과 수집
    for r := range pool.Results() {
        fmt.Println(r)
    }

    submitWG.Wait()
    close(monitorDone)

    completed, failed := pool.Stats()
    fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
    fmt.Printf("완료: %d  실패: %d\n", completed, failed)
}

func monitor(ctx context.Context, pool *checker.Pool, done <-chan struct{}) {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-done:
            return
        case <-ticker.C:
            c, f := pool.Stats()
            fmt.Fprintf(os.Stderr, "[모니터] 완료=%d 실패=%d\n", c, f)
        }
    }
}
```

## 7.6 Step 5 — Makefile

```makefile
BINARY := healthcheck

.PHONY: all build run test clean

all: build

build:
	go build -o bin/$(BINARY) ./cmd/$(BINARY)

run: build
	./bin/$(BINARY) -w 5 -t 10s

test:
	go test -race ./...

clean:
	rm -rf bin/
```

## 7.7 Step 6 — 빌드 & 실행

```bash
make build
./bin/healthcheck -w 3 -t 5s https://go.dev https://github.com https://example.com
```

기대 출력:
```
✅ https://example.com — 200 (245ms)
✅ https://go.dev — 200 (380ms)
✅ https://github.com — 200 (412ms)
[모니터] 완료=3 실패=0
━━━━━━━━━━━━━━━━━━━━━━━━━━
완료: 3  실패: 0
```

### Ctrl+C 테스트

```bash
./bin/healthcheck -w 2 -t 60s https://go.dev https://github.com ... (많이)
# 도중에 Ctrl+C
```

출력:
```
✅ https://go.dev — 200 (...)
^C
중단 신호 — 정리 중...
━━━━━━━━━━━━━━━━━━━━━━━━━━
완료: 1  실패: 0
```

## 7.8 🎯 도전 과제 (선택)

1. **재시도 로직** — 실패 시 최대 3회 재시도 (지수 백오프)
2. **결과 JSON 출력** — `-json` 플래그로 결과를 JSON 형식으로
3. **URL 파일 입력** — `-f urls.txt`로 URL을 파일에서 읽기
4. **응답 시간 통계** — p50/p95/p99 출력
5. **Rate Limit** — `-rps 10` 옵션으로 초당 요청 수 제한

### ✅ 7교시 체크포인트

- [ ] Worker Pool을 패키지로 깔끔하게 분리할 수 있는가?
- [ ] ctx로 전체 풀 취소를 구현할 수 있는가?
- [ ] atomic 카운터로 진행 상황을 모니터링할 수 있는가?
- [ ] graceful shutdown을 시그널 + ctx로 구현할 수 있는가?

---

# 8교시. 실습 — 파이프라인 기반 데이터 처리

마지막 실습은 **로그 파일 분석 파이프라인**입니다.

## 8.1 시나리오: 로그 파일 분석기

- 입력: 다수의 로그 파일 (또는 stdin)
- 단계:
  1. **읽기**: 파일에서 줄 읽기
  2. **파싱**: 각 줄을 LogEntry로 변환
  3. **필터링**: ERROR 레벨만
  4. **변환**: 시간대별로 카운트
  5. **집계**: 최종 통계 출력

```text
files → [read] → lines → [parse] → entries → [filter] → errors → [count] → stats
```

각 단계는 별도 고루틴 + 채널.

## 8.2 Step 1 — 프로젝트 셋업

```bash
mkdir -p ~/go-class/day4/loganalyzer
cd ~/go-class/day4/loganalyzer
go mod init loganalyzer
```

## 8.3 Step 2 — 타입 정의

```go
// types.go
package main

import "time"

type LogEntry struct {
    Timestamp time.Time
    Level     string
    Message   string
    Source    string
}

type LogLine struct {
    Source string  // 어느 파일에서 왔는지
    Line   string
}
```

## 8.4 Step 3 — Stage 1: 파일 읽기

```go
// stage_read.go
package main

import (
    "bufio"
    "context"
    "fmt"
    "io"
    "os"
)

func readFiles(ctx context.Context, paths []string) <-chan LogLine {
    out := make(chan LogLine, 100)

    go func() {
        defer close(out)

        for _, path := range paths {
            select {
            case <-ctx.Done():
                return
            default:
            }

            var reader io.ReadCloser
            if path == "-" {
                reader = io.NopCloser(os.Stdin)
            } else {
                f, err := os.Open(path)
                if err != nil {
                    fmt.Fprintf(os.Stderr, "파일 열기 실패: %s: %v\n", path, err)
                    continue
                }
                reader = f
            }

            scanner := bufio.NewScanner(reader)
            for scanner.Scan() {
                line := LogLine{Source: path, Line: scanner.Text()}
                select {
                case out <- line:
                case <-ctx.Done():
                    reader.Close()
                    return
                }
            }
            reader.Close()
        }
    }()

    return out
}
```

## 8.5 Step 4 — Stage 2: 파싱 (Fan-out)

여러 워커가 동시에 파싱.

```go
// stage_parse.go
package main

import (
    "context"
    "regexp"
    "sync"
    "time"
)

// 로그 형식: 2024-01-15T10:30:45Z [LEVEL] message
var logPattern = regexp.MustCompile(`^(\S+)\s+\[(\w+)\]\s+(.*)$`)

type ParseResult struct {
    Entry LogEntry
    Err   error
}

func parseStage(ctx context.Context, in <-chan LogLine, numWorkers int) <-chan ParseResult {
    out := make(chan ParseResult, 100)
    var wg sync.WaitGroup

    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for ll := range in {
                result := parseOne(ll)
                select {
                case out <- result:
                case <-ctx.Done():
                    return
                }
            }
        }()
    }

    go func() {
        wg.Wait()
        close(out)
    }()

    return out
}

func parseOne(ll LogLine) ParseResult {
    m := logPattern.FindStringSubmatch(ll.Line)
    if m == nil {
        return ParseResult{Err: fmt.Errorf("형식 오류: %s", ll.Line)}
    }

    ts, err := time.Parse(time.RFC3339, m[1])
    if err != nil {
        return ParseResult{Err: err}
    }

    return ParseResult{Entry: LogEntry{
        Timestamp: ts,
        Level:     m[2],
        Message:   m[3],
        Source:    ll.Source,
    }}
}
```

## 8.6 Step 5 — Stage 3: 필터 + Stage 4: 시간대별 키 변환

```go
// stage_filter.go
package main

import (
    "context"
    "fmt"
)

func filterErrors(ctx context.Context, in <-chan ParseResult) <-chan LogEntry {
    out := make(chan LogEntry, 100)
    go func() {
        defer close(out)
        for r := range in {
            if r.Err != nil {
                continue  // 또는 별도 에러 채널
            }
            if r.Entry.Level != "ERROR" {
                continue
            }
            select {
            case out <- r.Entry:
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}

// 시간을 시간(hour) 단위로 절삭
type HourKey struct {
    Hour string  // "2024-01-15T10"
}

type HourEntry struct {
    Key   HourKey
    Entry LogEntry
}

func bucketize(ctx context.Context, in <-chan LogEntry) <-chan HourEntry {
    out := make(chan HourEntry, 100)
    go func() {
        defer close(out)
        for e := range in {
            key := HourKey{Hour: e.Timestamp.Format("2006-01-02T15")}
            select {
            case out <- HourEntry{Key: key, Entry: e}:
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}
```

## 8.7 Step 6 — Stage 5: 집계 (싱크)

```go
// stage_aggregate.go
package main

import "context"

func aggregate(ctx context.Context, in <-chan HourEntry) map[string]int {
    counts := make(map[string]int)
    for he := range in {
        select {
        case <-ctx.Done():
            return counts
        default:
        }
        counts[he.Key.Hour]++
    }
    return counts
}
```

## 8.8 Step 7 — main.go (조립)

```go
// main.go
package main

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "sort"
    "syscall"
    "time"
)

func main() {
    paths := os.Args[1:]
    if len(paths) == 0 {
        paths = []string{"-"}  // stdin
    }

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
    go func() {
        <-sigs
        fmt.Fprintln(os.Stderr, "\n중단 — 부분 결과 출력")
        cancel()
    }()

    start := time.Now()

    // 파이프라인 조립
    lines := readFiles(ctx, paths)
    parsed := parseStage(ctx, lines, 4)  // 4개 파서
    errors := filterErrors(ctx, parsed)
    bucketed := bucketize(ctx, errors)
    stats := aggregate(ctx, bucketed)

    // 결과 출력
    elapsed := time.Since(start)
    fmt.Printf("\n=== ERROR 로그 시간대별 집계 ===\n")

    hours := make([]string, 0, len(stats))
    for h := range stats {
        hours = append(hours, h)
    }
    sort.Strings(hours)

    total := 0
    for _, h := range hours {
        fmt.Printf("%s : %d건\n", h, stats[h])
        total += stats[h]
    }
    fmt.Printf("\n총 ERROR: %d건 (소요: %v)\n", total, elapsed)
}
```

## 8.9 Step 8 — 테스트 데이터 만들기

```bash
cat > sample.log << 'EOF'
2024-01-15T10:00:01Z [INFO] 서버 시작
2024-01-15T10:00:05Z [ERROR] DB 연결 실패
2024-01-15T10:15:22Z [WARN] 느린 응답
2024-01-15T10:30:11Z [ERROR] 인증 실패
2024-01-15T10:45:33Z [INFO] 정상 요청
2024-01-15T11:00:12Z [ERROR] 메모리 부족
2024-01-15T11:30:45Z [ERROR] 디스크 풀
2024-01-15T11:45:00Z [INFO] 백업 완료
2024-01-15T12:15:22Z [ERROR] 네트워크 단절
EOF

go run . sample.log
```

기대 출력:
```
=== ERROR 로그 시간대별 집계 ===
2024-01-15T10 : 2건
2024-01-15T11 : 2건
2024-01-15T12 : 1건

총 ERROR: 5건 (소요: 12.3ms)
```

## 8.10 부하 테스트

큰 로그 파일을 만들어 성능을 봅시다.

```bash
# 100만 줄 로그 생성
go run -mod=mod gen_logs.go > big.log
# (또는 직접 셸 스크립트로 생성)

time go run . big.log
```

각 단계가 동시에 실행되므로, **단일 스레드 처리보다 빠릅니다**. 특히 파싱 단계가 4개 워커로 fan-out되어 효과가 큽니다.

## 8.11 🎯 도전 과제 (선택)

1. **메트릭 단계 추가** — 각 단계의 처리 속도(ops/sec)를 별도 채널로 송신해 메트릭 고루틴이 출력
2. **여러 레벨 분리** — `ERROR`, `WARN`, `INFO`별로 각자 통계
3. **에러 채널** — 파싱 실패한 줄을 별도 채널로 분리 후 stderr 출력
4. **JSON 출력** — 결과를 JSON으로 직렬화

### ✅ 8교시 체크포인트

- [ ] 5단계 파이프라인을 구성할 수 있는가?
- [ ] 한 단계를 Fan-out으로 병렬화할 수 있는가?
- [ ] ctx로 모든 단계를 동시에 중단할 수 있는가?
- [ ] 채널 닫기 책임을 각 단계가 명확히 가지고 있는가?

---

# 🎓 4일차 마무리

## 오늘 배운 것

1. **`select` 심화**: nil 채널 트릭, non-blocking 통신, `for-select` 패턴
2. **Timeout/취소**: `time.After`, `Timer`, `Ticker`의 정확한 사용
3. **Context I**: `Background`/`TODO`, `WithCancel`, 취소 전파
4. **Context II**: `WithTimeout`, `WithValue` 베스트/안티 패턴
5. **Worker Pool**: 정적 풀, Bounded Concurrency (세마포어 채널)
6. **Pipeline**: 다단계 변환, Fan-out/Fan-in, 백프레셔, 에러 전파
7. **실전 ①**: URL 헬스 체커 (Worker Pool + ctx + graceful shutdown)
8. **실전 ②**: 로그 파일 분석 파이프라인 (5단계 데이터 흐름)

## 한 줄 요약

> **"Concurrency in Go is like LEGO — small, well-designed pieces that combine into anything."** — 고루틴 + 채널 + select + ctx, 단 4가지 도구로 모든 동시성 문제를 표현할 수 있습니다.

## 복습 과제

다음 시간 전에 다음을 직접 해보세요.

1. **Web Scraper 미니 프로젝트** — URL 목록을 받아 페이지 제목을 추출하는 도구. Worker Pool + ctx + 재시도 로직 포함
2. **Rate Limiter** — `time.Ticker`를 이용해 초당 N개로 호출 제한하는 헬퍼 함수 작성
3. **Context Tree Visualizer** — `WithCancel`/`WithTimeout`을 중첩해 만들고, 어떤 시점에 어디가 취소되는지 로그로 추적
4. **Pipeline 일반화** — 임의 타입의 데이터 파이프라인을 만들기 위한 제네릭 기반 헬퍼 (`Stage[In, Out any](fn func(In) Out) func(<-chan In) <-chan Out`)

## 다음 시간 예고 — 5일차 (최종일)

마지막 날은 **실전 프로젝트 통합**입니다.

- **표준 라이브러리 활용**: `net/http`, `encoding/json`, `database/sql`
- **테스트와 벤치마크**: 표 기반 테스트, mock, `testing` 패키지 심화
- **로깅과 메트릭**: `log/slog`(Go 1.21+), 구조화 로깅
- **종합 프로젝트**: REST API 서버 또는 동시성 작업 처리 시스템
- **Go 운영 노하우**: 프로파일링(`pprof`), 디버깅, 메모리 분석

---

## 📚 참고 자료

- [Context Package Documentation](https://pkg.go.dev/context) — 공식 문서
- [Go Concurrency Patterns: Context (2014)](https://go.dev/blog/context) — context 도입 배경 글
- [Go Concurrency Patterns: Pipelines (2014)](https://go.dev/blog/pipelines) — 파이프라인 정전
- [Practical Go: Real world advice for writing maintainable Go programs](https://dave.cheney.net/practical-go) — Dave Cheney의 실용 가이드
- [Go Patterns - Worker Pool](https://gobyexample.com/worker-pools)
- 책: 『Concurrency in Go』 Chapter 4 (Katherine Cox-Buday) — 파이프라인과 컨텍스트
- 영상: [Rob Pike - Go Concurrency Patterns (Google I/O 2012)](https://www.youtube.com/watch?v=f6kdp27TYZs) — 클래식

> **마지막 팁**: 동시성 코드는 **단순함이 곧 정확함**입니다. 영리한 트릭보다 명확한 구조를 추구하세요. Rob Pike 본인이 이 점을 가장 강조합니다.
