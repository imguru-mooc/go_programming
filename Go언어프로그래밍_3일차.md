# Go 언어 프로그래밍 3일차 — 동시성: 고루틴과 채널

> **대상**: 1~2일차를 마친 C 개발자
> **목표**: Go의 시그니처 기능인 **동시성(concurrency)** 모델을 이해하고, 고루틴·채널·sync 패키지를 자유롭게 조합해 동시성 프로그램을 작성할 수 있다.
> **준비물**: 1~2일차 환경, 멀티코어 CPU(현대 CPU는 모두 해당)

---

## 📋 3일차 시간표

| 교시 | 주제 | 핵심 내용 |
|---|---|---|
| 1교시 | Go 런타임 스케줄러 | M:N 모델, OS 스레드 vs 고루틴 |
| 2교시 | 고루틴 심화 | 스택 관리, 컨텍스트 스위칭 비용 |
| 3교시 | Channel 기초 | Unbuffered Channel, 동기화 메커니즘 |
| 4교시 | Channel 심화 | Buffered Channel, 패턴 활용 |
| 5교시 | sync 패키지 I | `Mutex`, `RWMutex` |
| 6교시 | sync 패키지 II | `WaitGroup`, `Once`, `Cond` |
| 7교시 | 실습 | Producer-Consumer 패턴 |
| 8교시 | 실습 | Mutex vs Channel 비교 |

---

# 1교시. Go 런타임 스케줄러

## 1.1 왜 동시성인가?

C 개발자에게 동시성은 익숙한 주제입니다. `pthread`로 스레드를 만들고, `mutex`로 보호하고, `condition variable`로 신호를 주고받은 경험이 있을 겁니다.

그런데 **수만 개의 동시 작업**을 다룬다고 생각해봅시다. 예를 들어:
- 웹 서버가 동시에 10,000개 클라이언트와 통신
- 비동기 I/O 작업을 100,000개 동시에 시작
- 실시간 채팅 서버가 1,000,000개 연결 유지

C/pthread로는 어렵습니다. **OS 스레드 하나당 보통 2~8MB 스택**을 소비하므로, 10,000개면 20GB가 넘는 메모리가 필요합니다. 게다가 컨텍스트 스위칭 비용도 무시할 수 없습니다.

Go는 이 문제를 **언어 차원에서** 해결했습니다.

## 1.2 동시성(Concurrency) vs 병렬성(Parallelism)

자주 혼동되는 두 개념을 명확히 합시다.

| | 동시성 (Concurrency) | 병렬성 (Parallelism) |
|---|---|---|
| 의미 | 여러 작업을 **다루는** 능력 | 여러 작업을 **동시에 실행** |
| 필요 조건 | 단일 코어에서도 가능 | 멀티 코어 필요 |
| 비유 | 한 명의 요리사가 여러 요리를 번갈아 만듦 | 여러 명의 요리사가 동시에 요리 |
| Rob Pike의 정의 | "한 번에 여러 일을 다루기" | "한 번에 여러 일을 실행하기" |

Go는 **동시성을 기본**으로 설계되었고, 멀티코어에서 자연스럽게 **병렬 실행**도 됩니다.

## 1.3 OS 스레드와 고루틴 비교

C의 pthread는 OS 스레드 1:1 매핑입니다.

```c
// C — pthread
pthread_t tid;
pthread_create(&tid, NULL, worker, NULL);
pthread_join(tid, NULL);
```

각 pthread는 **OS 커널이 직접 관리**하는 스레드 하나에 대응합니다.

Go의 고루틴은 다릅니다.

```go
// Go — goroutine
go worker()  // 끝. 단 한 줄.
```

**`go` 키워드를 함수 호출 앞에 붙이기만 하면** 새 고루틴이 시작됩니다. 그런데 이 고루틴은 OS 스레드가 아닙니다.

### 비교 표

| 항목 | OS 스레드 (pthread) | 고루틴 |
|---|---|---|
| 생성 비용 | ~10μs+ (시스템 콜) | ~1μs (사용자 공간) |
| 초기 스택 | 2MB (Linux 기본) | **2KB** (1000배 작음) |
| 최대 개수 | 수천 개 (실제 수만 개 어려움) | **수십만~수백만 개** |
| 컨텍스트 스위칭 | 커널 모드 진입 (~1μs+) | 사용자 공간 (~수백 ns) |
| 스케줄러 | OS 커널 | Go 런타임 |
| 통신 수단 | 공유 메모리 + mutex | **채널** (권장) |
| 식별자 | `pthread_t` | 없음 (의도적으로 숨김) |

> **고루틴은 OS 스레드 위에 다중화된 가벼운 실행 단위**입니다. Java의 가상 스레드, Erlang의 프로세스와 유사한 개념입니다.

## 1.4 M:N 스케줄링 모델

Go 런타임은 **M개의 OS 스레드 위에 N개의 고루틴**을 매핑합니다. 보통 N >> M입니다.

### 세 가지 핵심 추상화: G, M, P

Go 스케줄러는 세 가지 구성요소로 동작합니다.

```text
┌─────────────────────────────────────────────────────┐
│                    Go Runtime                       │
│                                                     │
│   ┌──────────┐    ┌──────────┐    ┌──────────┐      │
│   │   P 0    │    │   P 1    │    │   P 2    │      │
│   │  [G][G]  │    │  [G][G]  │    │  [G][G]  │      │
│   │  [G][G]  │    │  [G]     │    │  [G][G]  │      │
│   │   ↑      │    │   ↑      │    │   ↑      │      │
│   └───┼──────┘    └───┼──────┘    └───┼──────┘      │
│       │ M0            │ M1            │ M2          │
└───────┼───────────────┼───────────────┼─────────────┘
        ↓               ↓               ↓
    [CPU Core 0]   [CPU Core 1]   [CPU Core 2]
```

| 약자 | 의미 | 역할 |
|---|---|---|
| **G** (Goroutine) | 고루틴 | 실제 실행 단위, 함수 + 스택 + 상태 |
| **M** (Machine) | OS 스레드 | 실제 CPU에서 실행되는 커널 스레드 |
| **P** (Processor) | 논리 프로세서 | G를 실행할 권한, 보통 코어 수만큼 존재 |

### 동작 흐름

1. P는 **자신의 로컬 큐**에 실행 대기 중인 G들을 가지고 있습니다.
2. M은 P 하나와 연결되어, P의 큐에서 G를 꺼내 실행합니다.
3. G가 시스템 콜에 들어가면 M이 블로킹되므로, 런타임이 **새 M을 만들거나 다른 M을 가져다 P를 붙입니다**.
4. P의 로컬 큐가 비면 다른 P에서 G를 훔쳐옵니다 (**work stealing**).

### GOMAXPROCS — P의 개수 조절

```go
import "runtime"

n := runtime.GOMAXPROCS(0)  // 현재 값 조회 (변경 없음)
fmt.Println("GOMAXPROCS:", n)
// 보통 코어 수와 같음

// 수동 변경
runtime.GOMAXPROCS(4)
```

기본값은 **CPU 코어 수**입니다. 평소엔 건드릴 필요 없습니다.

## 1.5 C pthread vs Go 고루틴 — 실제 비교

같은 동작을 두 언어로 작성해봅시다.

### C pthread로 1만 개 작업

```c
#include <pthread.h>
#include <stdio.h>

void *worker(void *arg) {
    long id = (long)arg;
    // 작업...
    return NULL;
}

int main(void) {
    pthread_t threads[10000];
    for (long i = 0; i < 10000; i++) {
        if (pthread_create(&threads[i], NULL, worker, (void*)i) != 0) {
            perror("pthread_create");  // 실패 가능성 큼!
            return 1;
        }
    }
    for (int i = 0; i < 10000; i++) {
        pthread_join(threads[i], NULL);
    }
    return 0;
}
```

이 코드를 그대로 돌리면 대부분의 시스템에서 **메모리 부족**이나 **`EAGAIN`** 에러로 실패합니다. 10,000개 × 2MB = **20GB**가 필요하기 때문입니다.

### Go 고루틴으로 1만 개 작업

```go
package main

import (
    "fmt"
    "sync"
)

func worker(id int, wg *sync.WaitGroup) {
    defer wg.Done()
    // 작업...
    _ = id
}

func main() {
    var wg sync.WaitGroup
    for i := 0; i < 10000; i++ {
        wg.Add(1)
        go worker(i, &wg)
    }
    wg.Wait()
    fmt.Println("완료")
}
```

이 코드는 **즉시 동작합니다.** 메모리 사용량은 약 20MB 수준(10,000 × 2KB).

### 100만 개도 가능

```go
for i := 0; i < 1_000_000; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        time.Sleep(time.Second)
    }()
}
wg.Wait()
```

현대적인 PC에서 **2GB 정도 메모리**로 동작합니다. C로는 상상하기 어려운 규모입니다.

## 1.6 Go 스케줄러가 일하는 시점

Go 스케줄러는 다음 시점에 고루틴을 전환합니다.

1. **함수 호출 시점**: 컴파일러가 삽입한 체크 코드가 트리거
2. **채널 송수신**: 즉시 다른 고루틴으로 전환
3. **시스템 콜**: 블로킹 발생 시 P를 다른 M에 넘김
4. **`time.Sleep`, `select`** 등 명시적 양보
5. **GC** 등 런타임 이벤트
6. **`runtime.Gosched()`** 명시적 양보 호출

> Go 1.14부터는 **선점형(preemptive) 스케줄링**도 도입되어, CPU만 쓰는 루프도 일정 시간(10ms)이 지나면 강제 전환됩니다. 그 전까지는 협력형(cooperative) 위주였습니다.

### ✅ 1교시 체크포인트

- [ ] 동시성과 병렬성의 차이를 설명할 수 있는가?
- [ ] 고루틴이 OS 스레드와 어떻게 다른지 말할 수 있는가?
- [ ] G, M, P의 역할을 그림으로 그릴 수 있는가?
- [ ] 왜 Go가 1만 개 고루틴을 쉽게 다루는지 설명할 수 있는가?

---

# 2교시. 고루틴 심화 — 스택 관리, 컨텍스트 스위칭

## 2.1 고루틴의 정체 — 코드부터

```go
go func() {
    fmt.Println("새 고루틴!")
}()
```

이게 전부입니다. `go` 키워드 다음에 **함수 호출**이 옵니다. **함수가 아니라 함수 호출**이라는 점을 주의하세요.

```go
go myFunc       // ❌ 컴파일 에러 — 호출이 아님
go myFunc()     // ✅ OK
go myFunc(1, 2) // ✅ OK
```

### 주의: main 함수 끝나면 모든 고루틴 종료

```go
func main() {
    go func() {
        fmt.Println("hello from goroutine")
    }()
    // main이 끝나면 위 고루틴도 강제 종료!
    // 아무 출력도 없을 가능성 큼
}
```

C에서 `main`이 `pthread_join` 없이 끝나면 백그라운드 스레드들이 죽는 것과 같습니다. **명시적으로 기다려야 합니다** (5~6교시에서 다룰 `WaitGroup` 또는 채널 사용).

임시로는 `time.Sleep`을 쓸 수 있지만 **운에 맡기는 코드**입니다.

```go
func main() {
    go func() {
        fmt.Println("hello")
    }()
    time.Sleep(100 * time.Millisecond)  // 좋은 패턴 아님
}
```

## 2.2 스택 관리 — Go의 마법

C pthread는 스레드 생성 시 **고정 크기 스택**(보통 2~8MB)을 미리 할당합니다. 재귀가 너무 깊으면 stack overflow로 죽습니다.

Go 고루틴은 **동적으로 자라고 줄어드는 스택**을 사용합니다.

```
초기:   ┌────┐
        │ 2KB│   ← 시작 크기
        └────┘

자람:   ┌────────┐
        │  8KB   │   ← 자동으로 확장
        └────────┘

더 자람:┌────────────────┐
        │     32KB       │
        └────────────────┘

최대:                   1GB(64-bit) 까지 가능
```

### 어떻게 가능한가?

Go 컴파일러는 **모든 함수 진입부에 스택 검사 코드**를 삽입합니다.

```
function prologue:
  if 스택 부족:
    더 큰 스택 할당
    기존 스택 내용 복사
    포인터들 재조정
    계속 실행
```

이 작업은 1μs 이내에 끝나며, 고루틴이 실제로 필요한 만큼만 메모리를 씁니다.

> **결과**: 대부분의 고루틴은 **2~8KB**로 충분하기 때문에, 백만 개 고루틴도 가벼운 메모리로 운영 가능합니다.

## 2.3 컨텍스트 스위칭 비용

| 항목 | OS 스레드 | 고루틴 |
|---|---|---|
| 모드 전환 | 사용자 → 커널 → 사용자 | 사용자 공간만 |
| 저장 레지스터 | 전체 (수십 개) | 일부 (Go 컨벤션상 더 적음) |
| TLB flush | 발생 가능 | 없음 |
| 캐시 | 손실 잦음 | 손실 적음 |
| 소요 시간 | ~1μs+ | **~수백 ns** |

**약 10~100배 차이**가 납니다. 작업이 짧을수록 이 차이가 더 두드러집니다.

## 2.4 고루틴의 흔한 함정 ① — 클로저 캡처

C 개발자에게 가장 헷갈리는 부분입니다.

### ❌ 잘못된 코드

```go
for i := 0; i < 5; i++ {
    go func() {
        fmt.Println(i)
    }()
}
time.Sleep(100 * time.Millisecond)
```

**예상**: 0 1 2 3 4 (순서는 다를 수 있음)
**실제 (Go 1.21 이전)**: 5 5 5 5 5

### 왜 그럴까?

클로저가 `i`를 **값이 아니라 참조로** 캡처하기 때문입니다. 고루틴이 실행될 때 이미 `i`는 루프 종료 후 5가 되어 있습니다.

### ✅ 해결책 ① — 인자로 넘기기 (가장 안전)

```go
for i := 0; i < 5; i++ {
    go func(n int) {
        fmt.Println(n)
    }(i)  // i를 값으로 복사하여 전달
}
```

### ✅ 해결책 ② — 루프 변수 재선언

```go
for i := 0; i < 5; i++ {
    i := i  // 같은 이름의 새 변수 (루프마다 새로 생성)
    go func() {
        fmt.Println(i)
    }()
}
```

### 🆕 Go 1.22+ 에서는 자동 해결

**Go 1.22부터는 루프 변수가 매 반복마다 새 변수로 취급**되어 위 문제가 사라졌습니다. 하지만 라이브러리 코드를 작성할 때는 이전 버전도 고려해야 합니다.

```bash
go version
# 1.22 이상이면 위 ❌ 코드도 정상 동작
```

## 2.5 고루틴의 흔한 함정 ② — 고루틴 누수

고루틴은 **자동으로 끝나지 않습니다.** 호출자가 죽어도 살아남아 메모리를 차지합니다.

### 누수 시나리오

```go
func leak() {
    ch := make(chan int)
    go func() {
        val := <-ch  // 누군가 보내주기 전까지 영원히 대기
        fmt.Println(val)
    }()
    // ch에 아무도 안 보냄 → 고루틴이 영원히 살아있음
}
```

```go
for i := 0; i < 10000; i++ {
    leak()
}
// 메모리 사용량 계속 증가!
```

### 진단

`runtime.NumGoroutine()`으로 현재 고루틴 개수를 볼 수 있습니다.

```go
fmt.Println("goroutines:", runtime.NumGoroutine())
```

운영 중인 서버에 이 값을 메트릭으로 노출해 두면 누수를 조기에 탐지할 수 있습니다.

### 해결 — 명시적 종료 신호

```go
func notLeak(done <-chan struct{}) {
    ch := make(chan int)
    go func() {
        select {
        case val := <-ch:
            fmt.Println(val)
        case <-done:
            return  // 종료 신호 받으면 빠져나옴
        }
    }()
}
```

`select`는 4교시에 자세히 다루지만, **여러 채널 중 먼저 준비된 쪽을 선택**한다고 기억해두세요.

## 2.6 `runtime.Gosched()` — 명시적 양보

```go
runtime.Gosched()
```

현재 고루틴을 잠시 멈추고 다른 고루틴에게 차례를 줍니다. C의 `sched_yield()`와 비슷한 역할이지만, **거의 쓸 일이 없습니다.** Go 스케줄러가 충분히 똑똑하기 때문입니다.

### 🧪 실습 코드: `goroutine_basic.go`

```go
package main

import (
    "fmt"
    "runtime"
    "sync"
    "time"
)

func main() {
    fmt.Println("CPU 수:", runtime.NumCPU())
    fmt.Println("시작 시 고루틴 수:", runtime.NumGoroutine())

    var wg sync.WaitGroup

    // 100만 개 고루틴 생성 — 진짜 가능한지 확인!
    start := time.Now()
    const N = 1_000_000

    for i := 0; i < N; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            _ = id * 2  // 간단한 작업
        }(i)
    }

    fmt.Println("최대 고루틴 수:", runtime.NumGoroutine())
    wg.Wait()
    fmt.Printf("%d개 고루틴 완료: %v\n", N, time.Since(start))
    fmt.Println("종료 시 고루틴 수:", runtime.NumGoroutine())
}
```

실행:
```bash
go run goroutine_basic.go
# CPU 수: 8
# 시작 시 고루틴 수: 1
# 최대 고루틴 수: 약 800000+
# 1000000개 고루틴 완료: 약 300ms
# 종료 시 고루틴 수: 1
```

C로 같은 일을 하려면 **백만 스레드는 시도조차 불가능**합니다.

### ✅ 2교시 체크포인트

- [ ] `go` 키워드로 고루틴을 만들 수 있는가?
- [ ] 고루틴 스택이 동적으로 자란다는 점을 이해했는가?
- [ ] 클로저 캡처 함정과 해결책을 알고 있는가?
- [ ] 고루틴 누수를 막을 방법을 설명할 수 있는가?

---

# 3교시. Channel 기초 — 고루틴 간 통신

## 3.1 Go의 동시성 철학

> *"Don't communicate by sharing memory; share memory by communicating."*
> — Rob Pike

C에서 스레드 간 통신은 보통 이렇게 합니다.

```c
// 공유 변수 + mutex로 보호
pthread_mutex_t lock;
int shared_value;

pthread_mutex_lock(&lock);
shared_value = 42;
pthread_mutex_unlock(&lock);
```

Go도 이런 방식이 가능하지만(5교시), **권장되는 방법은 채널**입니다.

```go
// 채널을 통해 값 전달 — 공유 변수 자체가 없음
ch := make(chan int)
go func() {
    ch <- 42  // 송신
}()
v := <-ch     // 수신
```

채널은 **CSP (Communicating Sequential Processes)** 모델에서 영감을 받았습니다. 1978년 Tony Hoare가 제안한 동시성 이론의 실용적 구현입니다.

## 3.2 채널 만들기와 기본 사용

```go
// 정수를 보내는 채널
ch := make(chan int)

// 송신
ch <- 42

// 수신
v := <-ch
v, ok := <-ch  // ok는 채널이 닫혔는지 여부
```

### 화살표 방향이 데이터 흐름을 표현

```text
ch <- v    : v를 ch에 보냄    (데이터가 채널로)
<-ch       : ch에서 받음       (데이터가 채널에서)
v := <-ch  : 받은 값을 v에 저장
```

## 3.3 Unbuffered Channel — 만남의 광장

`make(chan int)`처럼 용량 없이 만들면 **unbuffered channel**입니다.

### 핵심 동작: 송신과 수신이 동시에 만나야 한다

```go
ch := make(chan int)

go func() {
    fmt.Println("보내기 직전")
    ch <- 1  // 받는 사람 나타날 때까지 블록!
    fmt.Println("보내기 완료")
}()

time.Sleep(time.Second)
fmt.Println("받기 시작")
v := <-ch  // 송신자와 만남
fmt.Println("받음:", v)
```

출력:
```
보내기 직전
[1초 대기]
받기 시작
받음: 1
보내기 완료
```

송신/수신이 **랑데부(rendezvous)**한다고 표현합니다. 둘 다 만나야 진행됩니다.

### Unbuffered 채널 = 강력한 동기화 도구

이 특성을 이용하면 **두 고루틴의 진행을 동기화**할 수 있습니다. C의 세마포어와 비슷한 효과를 채널 하나로 얻습니다.

```go
done := make(chan struct{})  // 신호 전용 채널

go func() {
    fmt.Println("작업 중...")
    time.Sleep(time.Second)
    fmt.Println("작업 완료")
    done <- struct{}{}  // 신호 전송
}()

<-done  // 신호 받을 때까지 대기
fmt.Println("메인 계속")
```

> **관용구**: `chan struct{}`은 "신호" 용도. `struct{}`는 0바이트라 메모리 효율 최고.

## 3.4 채널 닫기

송신이 끝났음을 알리려면 채널을 닫습니다.

```go
close(ch)
```

### 규칙

| 행동 | 동작 |
|---|---|
| 닫힌 채널에 송신 | **panic** |
| 닫힌 채널에서 수신 | 즉시 zero value 반환 (블록되지 않음) |
| 닫힌 채널 닫기 | **panic** |
| nil 채널 송수신 | 영원히 블록 |

### `v, ok := <-ch` 패턴

```go
v, ok := <-ch
if !ok {
    // 채널이 닫혔고 값이 더 없음
}
```

### `range`로 순회

채널을 `range`로 돌리면 닫힐 때까지 자동으로 수신합니다.

```go
ch := make(chan int)

go func() {
    defer close(ch)
    for i := 1; i <= 5; i++ {
        ch <- i
    }
}()

for v := range ch {
    fmt.Println(v)
}
// 1 2 3 4 5
```

**닫기 책임은 송신자**가 집니다. 수신자가 닫으면 송신자가 panic할 수 있기 때문입니다.

## 3.5 채널 방향 — 일방통행 채널

함수 인자 타입에 방향을 명시할 수 있습니다.

```go
// 송신 전용
func producer(ch chan<- int) {
    ch <- 42
    // <-ch  // ❌ 컴파일 에러
}

// 수신 전용
func consumer(ch <-chan int) {
    v := <-ch
    fmt.Println(v)
    // ch <- 1  // ❌ 컴파일 에러
}

func main() {
    ch := make(chan int)
    go producer(ch)
    consumer(ch)
}
```

C에 비유하면 `const` 같은 역할입니다. **컴파일 시점에 의도를 강제**하여 버그를 줄입니다.

## 3.6 데드락 — 가장 흔한 함정

**모든 고루틴이 멈추면** Go 런타임이 데드락을 감지하고 패닉합니다.

```go
func main() {
    ch := make(chan int)
    ch <- 1  // 받는 사람 없음 → 영원히 블록
    // fatal error: all goroutines are asleep - deadlock!
}
```

### 흔한 데드락 패턴

```go
// ❌ 잘못 - 송신과 수신을 같은 고루틴에서
func main() {
    ch := make(chan int)
    ch <- 1     // 블록 — 받는 사람 필요
    <-ch        // 여기까지 못 옴
}

// ✅ 올바름 - 다른 고루틴에서 송신
func main() {
    ch := make(chan int)
    go func() { ch <- 1 }()  // 별도 고루틴
    <-ch                      // OK
}
```

Go 런타임의 데드락 감지는 **모든 고루틴이 동시에 멈췄을 때만** 동작합니다. 일부만 멈추면 감지되지 않으므로 주의해야 합니다.

## 3.7 🧪 실습 코드: `channel_basic.go`

```go
package main

import (
    "fmt"
    "time"
)

// 작업자 — 수신 전용 채널
func worker(id int, jobs <-chan int, results chan<- int) {
    for j := range jobs {
        fmt.Printf("worker %d: %d 처리 중...\n", id, j)
        time.Sleep(100 * time.Millisecond)
        results <- j * 2
    }
    fmt.Printf("worker %d: 종료\n", id)
}

func main() {
    jobs := make(chan int)
    results := make(chan int)

    // 워커 3개 시작
    for w := 1; w <= 3; w++ {
        go worker(w, jobs, results)
    }

    // 작업 5개 송신 후 채널 닫기
    go func() {
        for j := 1; j <= 5; j++ {
            jobs <- j
        }
        close(jobs)  // 더 이상 작업 없음 알림
    }()

    // 결과 5개 수신
    for r := 1; r <= 5; r++ {
        v := <-results
        fmt.Println("결과:", v)
    }
}
```

실행:
```bash
go run channel_basic.go
```

이 예제는 다음을 보여줍니다.
- **여러 워커가 같은 jobs 채널에서 작업 수신** (자연스러운 부하 분산)
- **`range`로 채널 순회**
- **`close`로 작업 종료 신호**

### ✅ 3교시 체크포인트

- [ ] Unbuffered 채널의 랑데부 의미를 설명할 수 있는가?
- [ ] 송수신 방향을 함수 시그니처에 표현할 수 있는가?
- [ ] `close`와 `range`의 협력 동작을 이해했는가?
- [ ] 데드락이 발생하는 패턴을 식별할 수 있는가?

---

# 4교시. Channel 심화 — Buffered, select, 패턴

## 4.1 Buffered Channel — 큐가 있는 채널

```go
ch := make(chan int, 3)  // 용량 3
```

unbuffered와의 차이:

| | Unbuffered | Buffered |
|---|---|---|
| 송신 | 수신자 만날 때까지 블록 | **버퍼 여유 있으면 즉시 진행** |
| 수신 | 송신자 만날 때까지 블록 | **버퍼에 값 있으면 즉시 진행** |
| 용도 | 강한 동기화 | 비동기 통신, 부하 흡수 |

### 동작 예시

```go
ch := make(chan int, 2)

ch <- 1    // 즉시 진행 (버퍼: [1])
ch <- 2    // 즉시 진행 (버퍼: [1, 2])
// ch <- 3 // 여기서 블록 - 버퍼 가득

fmt.Println(<-ch)  // 1 (버퍼: [2])
ch <- 3            // 이제 OK (버퍼: [2, 3])
```

C로 비유하면 **producer-consumer 큐가 있는 파이프**입니다.

### `len`과 `cap`

```go
ch := make(chan int, 5)
ch <- 1
ch <- 2

fmt.Println(len(ch))  // 2 - 현재 들어있는 개수
fmt.Println(cap(ch))  // 5 - 최대 용량
```

## 4.2 언제 Buffered를 쓰나?

### ✅ 적절한 경우

1. **생산/소비 속도 차이 흡수** — 일시적 burst 처리
2. **확실히 알려진 작업 개수** — `make(chan int, N)`로 N개 송신
3. **고정 크기 워커 풀** — 동시에 N개 작업만 진행

### ❌ 부적절한 경우

1. **"빠르게 하기 위해" 무작정 버퍼링** — 문제 가리는 효과
2. **메모리 누수 우려** — 버퍼에 값이 쌓이기만 함
3. **순서 보장이 중요한데 buffered 사용** — 모든 채널은 FIFO이지만, 다중 송신자가 있으면 모듈 변경에 취약

**기본 원칙**: 일단 unbuffered로 시작하고, 성능 측정 후 필요하면 버퍼링 추가.

## 4.3 `select` — 채널 다중화

여러 채널을 동시에 다루는 핵심 구문입니다. C의 `select()` / `poll()`과 개념이 비슷하지만, 더 강력합니다.

```go
select {
case v := <-ch1:
    fmt.Println("ch1에서:", v)
case v := <-ch2:
    fmt.Println("ch2에서:", v)
case ch3 <- 42:
    fmt.Println("ch3에 송신")
}
```

### 동작 규칙

1. **준비된 case가 하나** → 그것을 실행
2. **여러 개 준비됨** → **랜덤하게** 하나 선택 (편향 방지)
3. **모두 준비 안 됨** → 하나라도 준비될 때까지 블록
4. **`default` 있음** → 즉시 default 실행 (non-blocking)

### Non-blocking 송수신

```go
select {
case v := <-ch:
    fmt.Println("받음:", v)
default:
    fmt.Println("받을 게 없음")
}
```

### Timeout 패턴 (4일차에서 깊이 다룸)

```go
select {
case v := <-ch:
    fmt.Println("받음:", v)
case <-time.After(time.Second):
    fmt.Println("타임아웃!")
}
```

## 4.4 자주 쓰는 채널 패턴들

### 패턴 ① — Done 채널로 종료 신호

```go
func worker(done <-chan struct{}) {
    for {
        select {
        case <-done:
            return
        default:
            // 작업 진행
        }
    }
}

func main() {
    done := make(chan struct{})
    go worker(done)
    time.Sleep(time.Second)
    close(done)  // 종료 신호 (모든 수신자에게 broadcast)
}
```

**`close(done)`이 핵심**입니다. 닫힌 채널은 모든 수신자에게 즉시 zero value를 반환하므로, **여러 고루틴에 한 번에 신호**를 보낼 수 있습니다.

### 패턴 ② — 생성자(Generator)

```go
func counter() <-chan int {
    ch := make(chan int)
    go func() {
        defer close(ch)
        for i := 1; i <= 5; i++ {
            ch <- i
        }
    }()
    return ch
}

func main() {
    for v := range counter() {
        fmt.Println(v)
    }
}
```

C++의 코루틴, Python의 generator와 같은 패턴입니다.

### 패턴 ③ — Fan-out (작업 분배)

여러 워커가 같은 채널에서 수신하면 자동으로 부하 분산됩니다.

```go
jobs := make(chan int)

// 워커 N개
for i := 0; i < 5; i++ {
    go func(id int) {
        for job := range jobs {
            process(id, job)
        }
    }(i)
}

// 작업 송신
for j := 1; j <= 100; j++ {
    jobs <- j
}
close(jobs)
```

### 패턴 ④ — Fan-in (결과 수집)

여러 채널을 하나로 합칩니다.

```go
func merge(chs ...<-chan int) <-chan int {
    out := make(chan int)
    var wg sync.WaitGroup

    for _, ch := range chs {
        wg.Add(1)
        go func(c <-chan int) {
            defer wg.Done()
            for v := range c {
                out <- v
            }
        }(ch)
    }

    go func() {
        wg.Wait()
        close(out)
    }()

    return out
}
```

### 패턴 ⑤ — Heartbeat (생존 신호)

```go
func worker(done <-chan struct{}) <-chan struct{} {
    heartbeat := make(chan struct{}, 1)
    go func() {
        defer close(heartbeat)
        tick := time.NewTicker(time.Second)
        defer tick.Stop()
        for {
            select {
            case <-done:
                return
            case <-tick.C:
                select {
                case heartbeat <- struct{}{}:
                default:
                }
            }
        }
    }()
    return heartbeat
}
```

## 4.5 🧪 실습 코드: `channel_patterns.go`

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

// 생성자
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

// 변환 단계
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

// Fan-in
func merge(chs ...<-chan int) <-chan int {
    var wg sync.WaitGroup
    out := make(chan int)

    output := func(c <-chan int) {
        defer wg.Done()
        for v := range c {
            out <- v
        }
    }

    wg.Add(len(chs))
    for _, c := range chs {
        go output(c)
    }

    go func() {
        wg.Wait()
        close(out)
    }()
    return out
}

func main() {
    // 파이프라인: 생성 → 제곱 → 수집
    in := generate(1, 2, 3, 4, 5)

    // Fan-out: 두 워커가 제곱 처리
    c1 := square(in)
    // 주의: 같은 in을 두 번 못 씀 - 별도 생성
    c2 := square(generate(6, 7, 8, 9, 10))

    // Fan-in: 결과 합치기
    for v := range merge(c1, c2) {
        fmt.Println(v)
    }

    // select + timeout 데모
    timeoutDemo()
}

func timeoutDemo() {
    ch := make(chan string)
    go func() {
        time.Sleep(2 * time.Second)
        ch <- "느린 응답"
    }()

    select {
    case v := <-ch:
        fmt.Println("받음:", v)
    case <-time.After(1 * time.Second):
        fmt.Println("타임아웃!")
    }
}
```

### ✅ 4교시 체크포인트

- [ ] unbuffered와 buffered 채널을 적절히 구분해 쓸 수 있는가?
- [ ] `select`로 여러 채널을 다중화할 수 있는가?
- [ ] `close(done)`을 broadcast 신호로 활용할 수 있는가?
- [ ] Fan-out / Fan-in 패턴을 구현할 수 있는가?

---

# 5교시. sync 패키지 I — Mutex와 RWMutex

## 5.1 채널만으로 부족할 때

지금까지 채널이 만능처럼 보였지만, 채널도 단점이 있습니다.

- **단순 카운터 증가** 같은 경우 채널은 오버킬
- **공유 자료구조 보호** 시 명시적 락이 더 자연스러움
- **성능이 매우 중요한 hot path**에서 채널보다 mutex가 빠를 수 있음

> **Go 격언**: "Channels orchestrate; mutexes serialize."
> 채널은 **고루틴 간 흐름 조율**, mutex는 **데이터 보호**에 적합합니다.

## 5.2 race condition 시연

먼저 문제를 봅시다.

```go
package main

import (
    "fmt"
    "sync"
)

var counter int

func main() {
    var wg sync.WaitGroup
    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            counter++  // 보호 없음 — race!
        }()
    }
    wg.Wait()
    fmt.Println(counter)  // 1000이 안 나옴!
}
```

여러 번 실행하면 결과가 매번 다릅니다(987, 994, 1000, ...).

### `-race` 플래그로 감지

```bash
go run -race main.go
```

```
==================
WARNING: DATA RACE
Write at 0x000000123456 by goroutine 7:
  main.main.func1()
      ...
Previous write at 0x000000123456 by goroutine 6:
  main.main.func1()
      ...
==================
```

**`-race`는 Go의 강력한 무기**입니다. 개발 중 항상 켜고 테스트하세요. C/C++의 ThreadSanitizer와 같은 역할이지만, 별도 설치 없이 즉시 쓸 수 있습니다.

## 5.3 `sync.Mutex`

가장 기본적인 락입니다.

```go
import "sync"

var (
    counter int
    mu      sync.Mutex
)

func increment() {
    mu.Lock()
    defer mu.Unlock()
    counter++
}
```

### C pthread_mutex와 비교

```c
// C
#include <pthread.h>

int counter;
pthread_mutex_t mu = PTHREAD_MUTEX_INITIALIZER;

void increment(void) {
    pthread_mutex_lock(&mu);
    counter++;
    pthread_mutex_unlock(&mu);
}
```

```go
// Go
var (
    counter int
    mu      sync.Mutex
)

func increment() {
    mu.Lock()
    defer mu.Unlock()
    counter++
}
```

| 항목 | C pthread_mutex | Go sync.Mutex |
|---|---|---|
| 초기화 | `PTHREAD_MUTEX_INITIALIZER` 또는 `pthread_mutex_init` | **불필요** (zero value 사용 가능) |
| 락 해제 보장 | 수동 (`pthread_mutex_unlock`) | **`defer`로 자동** |
| 에러 반환 | 있음 | 없음 (panic만) |
| 종료 시 해제 | `pthread_mutex_destroy` | 불필요 (GC가 처리) |

### `defer mu.Unlock()` 패턴

```go
func criticalSection() error {
    mu.Lock()
    defer mu.Unlock()  // 함수 종료 시점에 자동 해제

    if err := doSomething(); err != nil {
        return err  // 여기서 return해도 unlock 보장
    }
    if err := doOther(); err != nil {
        return err  // 여기서 panic이 나도 unlock 보장
    }
    return nil
}
```

C에서 여러 return 경로마다 `pthread_mutex_unlock`을 일일이 호출하거나 `goto cleanup` 패턴을 쓰던 번거로움이 사라집니다.

## 5.4 `sync.RWMutex` — 읽기/쓰기 분리

**읽기가 많고 쓰기가 적은** 경우, 일반 Mutex는 비효율적입니다. 읽기끼리는 충돌하지 않으므로 동시에 허용해도 안전합니다.

```go
var (
    cache = make(map[string]string)
    rw    sync.RWMutex
)

func read(key string) string {
    rw.RLock()         // 읽기 락 — 동시에 여러 개 가능
    defer rw.RUnlock()
    return cache[key]
}

func write(key, value string) {
    rw.Lock()          // 쓰기 락 — 단독
    defer rw.Unlock()
    cache[key] = value
}
```

### 락 종류 비교

| | 일반 Mutex | RWMutex (Read) | RWMutex (Write) |
|---|---|---|---|
| 다른 동일 락 보유자 허용? | ❌ | ✅ (Read끼리) | ❌ |
| Read 보유 중 Write 시도 | - | 블록 | - |
| Write 보유 중 Read 시도 | - | - | 블록 |
| 적합 상황 | 읽기/쓰기 비율 비슷 | 읽기 >> 쓰기 | (해당 없음) |

### 🔥 주의: RWMutex가 항상 더 빠른 건 아님

- **짧은 critical section**: 일반 Mutex가 더 빠를 수 있음 (RW는 오버헤드 큼)
- **쓰기 빈도가 높음**: Mutex와 비슷하거나 더 느림
- **벤치마크 필수**: 추측하지 말고 측정

## 5.5 Mutex 사용 시 주의사항

### ❌ 함정 ① — 값 복사

```go
// 잘못된 코드
type Counter struct {
    mu    sync.Mutex
    value int
}

func badPattern() {
    c := Counter{}
    other := c  // ❌ Mutex가 복사됨 — 새 락이 생김
}
```

Mutex를 포함한 구조체는 **포인터로 전달**해야 합니다.

```go
// 올바른 코드
func goodPattern() {
    c := &Counter{}
    work(c)  // 포인터 전달
}

func work(c *Counter) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.value++
}
```

`go vet`이 이런 실수를 잡아줍니다.

### ❌ 함정 ② — 잊은 Unlock

```go
mu.Lock()
if err != nil {
    return  // ❌ Unlock 안 함 — 데드락
}
mu.Unlock()
```

해결: **항상 `defer`**.

### ❌ 함정 ③ — 락 순서 불일치 (deadlock)

```go
// 고루틴 A
mu1.Lock()
mu2.Lock()
// ...
mu2.Unlock()
mu1.Unlock()

// 고루틴 B - 반대 순서!
mu2.Lock()
mu1.Lock()  // 데드락 가능
// ...
```

**락은 항상 같은 순서로 획득**하는 규칙을 정해두세요. C에서도 동일한 원칙입니다.

## 5.6 🧪 실습 코드: `mutex_basic.go`

```go
package main

import (
    "fmt"
    "sync"
)

// 안전한 카운터
type Counter struct {
    mu    sync.Mutex
    value int
}

func (c *Counter) Inc() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.value++
}

func (c *Counter) Value() int {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.value
}

// 안전한 캐시
type Cache struct {
    rw   sync.RWMutex
    data map[string]string
}

func NewCache() *Cache {
    return &Cache{data: make(map[string]string)}
}

func (c *Cache) Get(key string) (string, bool) {
    c.rw.RLock()
    defer c.rw.RUnlock()
    v, ok := c.data[key]
    return v, ok
}

func (c *Cache) Set(key, value string) {
    c.rw.Lock()
    defer c.rw.Unlock()
    c.data[key] = value
}

func main() {
    // 카운터 테스트
    var wg sync.WaitGroup
    c := &Counter{}

    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            c.Inc()
        }()
    }
    wg.Wait()
    fmt.Println("카운터:", c.Value()) // 정확히 1000

    // 캐시 테스트
    cache := NewCache()
    cache.Set("name", "Alice")
    cache.Set("age", "30")

    if v, ok := cache.Get("name"); ok {
        fmt.Println("name =", v)
    }
}
```

`-race`로 실행해 안전성 확인:
```bash
go run -race mutex_basic.go
```

### ✅ 5교시 체크포인트

- [ ] `sync.Mutex`를 `defer`와 함께 쓸 수 있는가?
- [ ] race condition을 `-race`로 탐지할 수 있는가?
- [ ] `RWMutex`가 유리한 상황을 판단할 수 있는가?
- [ ] Mutex 함정 3가지(복사, 미해제, 순서)를 피할 수 있는가?

---

# 6교시. sync 패키지 II — WaitGroup, Once, Cond

## 6.1 `sync.WaitGroup` — 여러 고루틴 기다리기

C에서 `pthread_join`을 여러 번 호출하던 패턴을 한 번에 처리합니다.

```go
var wg sync.WaitGroup

for i := 0; i < 5; i++ {
    wg.Add(1)               // 카운터 +1
    go func(id int) {
        defer wg.Done()     // 카운터 -1
        // 작업...
        fmt.Println("worker", id, "완료")
    }(i)
}

wg.Wait()  // 카운터가 0이 될 때까지 대기
fmt.Println("모두 끝남")
```

내부적으로 **세마포어 카운터**입니다.

| 메서드 | 효과 |
|---|---|
| `Add(n)` | 카운터에 n 더함 |
| `Done()` | 카운터 -1 (=`Add(-1)`) |
| `Wait()` | 카운터가 0이 될 때까지 블록 |

### 🔥 주의: Add는 고루틴 시작 **전**에

```go
// ❌ 잘못된 순서
go func() {
    wg.Add(1)  // 너무 늦음 — Wait가 먼저 통과할 수 있음
    defer wg.Done()
}()

// ✅ 올바른 순서
wg.Add(1)
go func() {
    defer wg.Done()
}()
```

### 패턴: 작업 분배 후 결과 수집

```go
func parallelMap(items []int, f func(int) int) []int {
    results := make([]int, len(items))
    var wg sync.WaitGroup

    for i, v := range items {
        wg.Add(1)
        go func(idx, val int) {
            defer wg.Done()
            results[idx] = f(val)  // 인덱스 다르므로 race 없음
        }(i, v)
    }

    wg.Wait()
    return results
}
```

각 고루틴이 **서로 다른 인덱스**에 쓰므로 락 없이도 안전합니다.

## 6.2 `sync.Once` — 정확히 한 번만 실행

싱글톤 초기화, 리소스 게으른 로딩 등에 사용합니다.

```go
var (
    once   sync.Once
    config *Config
)

func GetConfig() *Config {
    once.Do(func() {
        // 여러 고루틴이 동시에 호출해도 단 한 번만 실행됨
        config = loadConfig()
    })
    return config
}
```

C에서 `pthread_once`와 같은 역할입니다.

```c
// C
static pthread_once_t once_control = PTHREAD_ONCE_INIT;
pthread_once(&once_control, init_function);
```

### 활용 예: 비싼 초기화

```go
type Service struct {
    once sync.Once
    db   *sql.DB
}

func (s *Service) DB() *sql.DB {
    s.once.Do(func() {
        s.db = openDB()
    })
    return s.db
}
```

여러 고루틴이 `Service.DB()`를 동시에 호출해도 `openDB()`는 **정확히 한 번** 실행됩니다.

## 6.3 `sync.Cond` — 조건 변수

특정 조건이 충족될 때까지 기다리는 메커니즘. C `pthread_cond_t`와 같은 개념입니다.

> **솔직한 조언**: Go에서는 `Cond` 대신 **채널**을 쓰는 게 보통 더 깔끔합니다. 하지만 알아둘 가치는 있습니다.

### 사용 예 — 버퍼가 빌 때까지 대기

```go
type Queue struct {
    mu    sync.Mutex
    cond  *sync.Cond
    items []int
}

func NewQueue() *Queue {
    q := &Queue{}
    q.cond = sync.NewCond(&q.mu)
    return q
}

func (q *Queue) Push(v int) {
    q.mu.Lock()
    defer q.mu.Unlock()
    q.items = append(q.items, v)
    q.cond.Signal()  // 대기자 한 명 깨움
}

func (q *Queue) Pop() int {
    q.mu.Lock()
    defer q.mu.Unlock()
    for len(q.items) == 0 {
        q.cond.Wait()  // 락 풀고 대기, 깨어나면 다시 락
    }
    v := q.items[0]
    q.items = q.items[1:]
    return v
}
```

### C와 비교

```c
// C
pthread_mutex_lock(&mu);
while (queue_empty()) {
    pthread_cond_wait(&cond, &mu);
}
// ...
pthread_mutex_unlock(&mu);
```

```go
// Go
q.mu.Lock()
for queueEmpty() {
    q.cond.Wait()
}
// ...
q.mu.Unlock()
```

거의 일대일 대응입니다.

**핵심 규칙**: `Wait`는 **반드시 `for` 루프 안에서**. spurious wakeup (이유 없이 깨어남) 가능성 때문입니다. 이것도 C와 동일합니다.

| 메서드 | 효과 |
|---|---|
| `Wait()` | 락 풀고 대기, 깨어나면 락 재획득 |
| `Signal()` | 대기 중인 고루틴 하나 깨움 |
| `Broadcast()` | 대기 중인 모든 고루틴 깨움 |

## 6.4 동일 효과를 채널로 (선호되는 방식)

위 Queue를 채널로 다시 쓰면:

```go
type ChanQueue struct {
    items chan int
}

func NewChanQueue(capacity int) *ChanQueue {
    return &ChanQueue{items: make(chan int, capacity)}
}

func (q *ChanQueue) Push(v int) {
    q.items <- v  // 가득 차면 블록
}

func (q *ChanQueue) Pop() int {
    return <-q.items  // 비어있으면 블록
}
```

훨씬 짧고 명확합니다. 대부분의 경우 채널이 답입니다.

## 6.5 sync.Map — 동시성 맵 (보너스)

읽기/쓰기 패턴이 매우 특수한 경우용 맵입니다.

```go
var m sync.Map

m.Store("key1", 100)
v, ok := m.Load("key1")
m.Range(func(k, v interface{}) bool {
    fmt.Println(k, v)
    return true  // false면 순회 종료
})
```

**언제 일반 map + Mutex보다 유리한가?**

- **거의 항상 다른 키에 쓰기/읽기** (각 키가 한 번 쓰이고 여러 번 읽힘)
- **수가 많고 안정적인 키 집합**

위 조건이 아니면 **`map + sync.RWMutex`가 보통 더 빠릅니다.**

## 6.6 🧪 실습 코드: `sync_advanced.go`

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

// 1. WaitGroup - 병렬 합계
func parallelSum(nums []int) int {
    n := len(nums)
    if n == 0 {
        return 0
    }
    mid := n / 2

    var wg sync.WaitGroup
    var left, right int

    wg.Add(2)
    go func() {
        defer wg.Done()
        for _, v := range nums[:mid] {
            left += v
        }
    }()
    go func() {
        defer wg.Done()
        for _, v := range nums[mid:] {
            right += v
        }
    }()
    wg.Wait()

    return left + right
}

// 2. Once - 게으른 초기화
type Config struct {
    Loaded   time.Time
    Value    string
}

var (
    cfgOnce sync.Once
    cfg     *Config
)

func GetConfig() *Config {
    cfgOnce.Do(func() {
        fmt.Println("Config 로딩 중... (단 한 번만 출력)")
        time.Sleep(100 * time.Millisecond)
        cfg = &Config{
            Loaded: time.Now(),
            Value:  "production-config",
        }
    })
    return cfg
}

func main() {
    // WaitGroup 테스트
    nums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
    fmt.Println("합계:", parallelSum(nums))

    // Once 테스트 - 동시 호출
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            c := GetConfig()
            _ = c
        }()
    }
    wg.Wait()
    fmt.Println("Config 로딩 시각:", cfg.Loaded)
}
```

실행:
```bash
go run sync_advanced.go
```

"Config 로딩 중..."이 **딱 한 번만** 출력됨을 확인하세요.

### ✅ 6교시 체크포인트

- [ ] `WaitGroup`으로 여러 고루틴 완료를 기다릴 수 있는가?
- [ ] `Once`로 일회성 초기화를 안전하게 할 수 있는가?
- [ ] `Cond`의 `Wait`를 왜 `for` 루프에 넣어야 하는지 설명할 수 있는가?
- [ ] 채널과 sync 패키지 중 선택 기준을 가지고 있는가?

---

# 7교시. 실습 — Producer-Consumer 패턴

이번 시간은 가장 고전적인 동시성 패턴인 **Producer-Consumer**를 채널로 구현합니다.

## 7.1 시나리오: 로그 처리 파이프라인

상상해봅시다.

- **Producer 3개**: 각각 로그를 생성 (예: 다른 서버에서 수집)
- **Consumer 5개**: 받은 로그를 처리 (예: 파일에 쓰기, DB에 저장)
- **Channel**: 로그를 버퍼링하여 부하 흡수
- **Graceful shutdown**: Ctrl+C 시 안전하게 종료

```text
┌─────────────┐                        ┌─────────────┐
│ Producer 1  │──┐                  ┌──│ Consumer 1  │
└─────────────┘  │                  │  └─────────────┘
┌─────────────┐  │   ┌──────────┐   │  ┌─────────────┐
│ Producer 2  │──┼──→│   Chan   │──→├──│ Consumer 2  │
└─────────────┘  │   │ (buffer) │   │  └─────────────┘
┌─────────────┐  │   └──────────┘   │  ┌─────────────┐
│ Producer 3  │──┘                  └──│ Consumer 3  │
└─────────────┘                        └─────────────┘
                                       ┌─────────────┐
                                       │ Consumer 4  │
                                       └─────────────┘
                                       ┌─────────────┐
                                       │ Consumer 5  │
                                       └─────────────┘
```

## 7.2 Step 1 — 프로젝트 셋업

```bash
mkdir -p ~/go-class/day3/prodcons
cd ~/go-class/day3/prodcons
go mod init prodcons
```

## 7.3 Step 2 — 로그 타입 정의

`main.go`:

```go
package main

import (
    "fmt"
    "math/rand"
    "os"
    "os/signal"
    "sync"
    "syscall"
    "time"
)

// LogEntry — 처리할 작업 단위
type LogEntry struct {
    ProducerID int
    Sequence   int
    Message    string
    Timestamp  time.Time
}

func (l LogEntry) String() string {
    return fmt.Sprintf("[P%d-#%d %s] %s",
        l.ProducerID, l.Sequence, l.Timestamp.Format("15:04:05.000"), l.Message)
}
```

## 7.4 Step 3 — Producer 구현

```go
func producer(id int, out chan<- LogEntry, done <-chan struct{}, wg *sync.WaitGroup) {
    defer wg.Done()
    defer fmt.Printf("[Producer %d] 종료\n", id)

    seq := 0
    for {
        select {
        case <-done:
            return
        default:
        }

        // 0~200ms 사이 랜덤 대기 (실제 로그 생성 흉내)
        time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)

        entry := LogEntry{
            ProducerID: id,
            Sequence:   seq,
            Message:    fmt.Sprintf("event-%d", seq),
            Timestamp:  time.Now(),
        }
        seq++

        // 송신할 때도 done 확인 - 채널이 가득 차면 빠져나갈 수 있도록
        select {
        case out <- entry:
        case <-done:
            return
        }
    }
}
```

**핵심 포인트**:
- `select` 안에 `<-done`을 두어 **언제든 빠져나올 수 있게** 함
- 송신도 `select`로 감싸 채널이 가득 차도 종료 가능
- `defer wg.Done()`으로 종료 보장

## 7.5 Step 4 — Consumer 구현

```go
func consumer(id int, in <-chan LogEntry, wg *sync.WaitGroup) {
    defer wg.Done()
    defer fmt.Printf("[Consumer %d] 종료\n", id)

    for entry := range in {  // 채널 닫히면 자동 종료
        // 처리 시간 흉내 (50~150ms)
        time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)
        fmt.Printf("[Consumer %d] %s\n", id, entry)
    }
}
```

**Consumer는 더 간단합니다**:
- `range in`이 채널이 닫힐 때까지 자동 수신
- 별도 종료 신호 불필요 (채널 닫힘이 종료 신호)

## 7.6 Step 5 — Graceful Shutdown

Ctrl+C (SIGINT) 신호를 받아 안전하게 종료시키는 부분.

```go
const (
    numProducers   = 3
    numConsumers   = 5
    channelCapacity = 10
)

func main() {
    rand.Seed(time.Now().UnixNano())

    logChan := make(chan LogEntry, channelCapacity)
    done := make(chan struct{})

    // 시그널 핸들러
    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

    var producerWG, consumerWG sync.WaitGroup

    // Producer 시작
    for i := 1; i <= numProducers; i++ {
        producerWG.Add(1)
        go producer(i, logChan, done, &producerWG)
    }

    // Consumer 시작
    for i := 1; i <= numConsumers; i++ {
        consumerWG.Add(1)
        go consumer(i, logChan, &consumerWG)
    }

    fmt.Printf("실행 중... (Ctrl+C로 종료)\n")

    // 시그널 대기
    <-sigs
    fmt.Println("\n종료 신호 수신 — 정리 중...")

    // 1) Producer들에게 종료 신호
    close(done)
    producerWG.Wait()
    fmt.Println("모든 Producer 종료됨")

    // 2) 채널 닫기 - 더 이상 송신 없음
    close(logChan)

    // 3) Consumer는 남은 데이터 처리 후 자동 종료
    consumerWG.Wait()
    fmt.Println("모든 Consumer 종료됨")

    fmt.Println("정상 종료")
}
```

## 7.7 Step 6 — 종료 순서의 중요성

위 순서가 매우 중요합니다.

```
1. close(done)        → Producer들에게 "이제 그만 보내라"
2. producerWG.Wait()  → 모든 Producer가 송신 완전히 멈출 때까지 대기
3. close(logChan)     → 채널 닫기 (이제 안전 — 송신자 없음)
4. consumerWG.Wait()  → Consumer가 남은 데이터 처리 후 종료
```

**잘못된 순서**:

```go
// ❌ 잘못 — Producer가 아직 송신 중인데 채널 닫음
close(logChan)
close(done)
// → "send on closed channel" panic!
```

**규칙**: **송신자가 모두 멈춘 후에만 채널을 닫는다.**

## 7.8 Step 7 — 실행과 관찰

```bash
go build -o prodcons
./prodcons
```

출력 예시:
```
실행 중... (Ctrl+C로 종료)
[Consumer 1] [P1-#0 16:23:45.123] event-0
[Consumer 3] [P2-#0 16:23:45.156] event-0
[Consumer 2] [P3-#0 16:23:45.234] event-0
[Consumer 4] [P1-#1 16:23:45.345] event-1
...
^C
종료 신호 수신 — 정리 중...
[Producer 1] 종료
[Producer 2] 종료
[Producer 3] 종료
모든 Producer 종료됨
[Consumer 1] [P3-#5 16:23:46.456] event-5
[Consumer 2] [P2-#7 16:23:46.567] event-7
...
[Consumer 1] 종료
[Consumer 2] 종료
[Consumer 3] 종료
[Consumer 4] 종료
[Consumer 5] 종료
모든 Consumer 종료됨
정상 종료
```

**관찰 포인트**:
- 종료 시점에 채널에 남아있던 데이터를 Consumer들이 끝까지 처리
- 한 데이터가 두 번 처리되지 않음 (채널이 보장)
- 데이터 누락 없음

## 7.9 race 검사

```bash
go run -race main.go
# 잠시 후 Ctrl+C
# race 경고 없으면 안전
```

## 7.10 🎯 도전 과제 (선택)

1. **메트릭 추가** — Producer/Consumer별 처리 개수를 atomic 카운터로 집계해 종료 시 출력
2. **백프레셔(Backpressure) 측정** — 채널이 가득 차서 Producer가 블록된 시간을 측정
3. **Dynamic worker pool** — Consumer 개수를 런타임에 늘리고 줄일 수 있게 변경
4. **Priority queue** — LogEntry에 우선순위 필드를 두고, 높은 우선순위가 먼저 처리되도록

### ✅ 7교시 체크포인트

- [ ] Producer-Consumer 구조를 채널로 구현할 수 있는가?
- [ ] Graceful shutdown 순서(done → producers → channel → consumers)를 이해했는가?
- [ ] `select` + `done` 패턴을 활용할 수 있는가?
- [ ] `close(channel)`이 broadcast 신호로도 동작함을 알고 있는가?

---

# 8교시. 실습 — Mutex vs Channel 비교

같은 문제를 두 가지 방식으로 풀어보고, **언제 무엇을 써야 하는지** 감을 잡습니다.

## 8.1 문제 정의: 안전한 카운터

여러 고루틴이 카운터 값을 동시에 증감시키는 상황. 다음을 비교합니다.

1. **Race 버전**: 보호 없음 (틀린 결과)
2. **Mutex 버전**: `sync.Mutex` 사용
3. **Channel 버전**: 채널로 직렬화
4. **Atomic 버전**: `sync/atomic` 사용 (보너스)

## 8.2 Step 1 — 공통 인터페이스

```bash
mkdir -p ~/go-class/day3/counter
cd ~/go-class/day3/counter
go mod init counter
```

`counter.go`:

```go
package main

type Counter interface {
    Inc()
    Value() int
}
```

## 8.3 Step 2 — Race 버전 (의도적 버그)

```go
type RaceCounter struct {
    value int
}

func (c *RaceCounter) Inc()       { c.value++ }
func (c *RaceCounter) Value() int { return c.value }
```

## 8.4 Step 3 — Mutex 버전

```go
import "sync"

type MutexCounter struct {
    mu    sync.Mutex
    value int
}

func (c *MutexCounter) Inc() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.value++
}

func (c *MutexCounter) Value() int {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.value
}
```

## 8.5 Step 4 — Channel 버전

채널로 카운터를 보호하는 발상은 다음과 같습니다. **모든 연산을 한 고루틴에서만 처리**하면 race가 원천 차단됩니다.

```go
type ChannelCounter struct {
    incCh chan struct{}    // 증가 요청
    valCh chan chan int    // 값 조회 요청 (응답 채널 동봉)
    quit  chan struct{}
}

func NewChannelCounter() *ChannelCounter {
    c := &ChannelCounter{
        incCh: make(chan struct{}, 100),
        valCh: make(chan chan int),
        quit:  make(chan struct{}),
    }
    go c.loop()
    return c
}

func (c *ChannelCounter) loop() {
    value := 0
    for {
        select {
        case <-c.incCh:
            value++
        case respCh := <-c.valCh:
            respCh <- value
        case <-c.quit:
            return
        }
    }
}

func (c *ChannelCounter) Inc() {
    c.incCh <- struct{}{}
}

func (c *ChannelCounter) Value() int {
    resp := make(chan int)
    c.valCh <- resp
    return <-resp
}

func (c *ChannelCounter) Close() {
    close(c.quit)
}
```

**핵심 아이디어**:
- `loop` 고루틴이 카운터의 **유일한 소유자**
- 모든 연산은 채널로 직렬화되어 race 발생 불가
- 락 없이도 안전

## 8.6 Step 5 — Atomic 버전 (보너스)

가장 빠른 방법. CPU 명령어 수준의 원자 연산을 사용합니다.

```go
import "sync/atomic"

type AtomicCounter struct {
    value int64
}

func (c *AtomicCounter) Inc() {
    atomic.AddInt64(&c.value, 1)
}

func (c *AtomicCounter) Value() int {
    return int(atomic.LoadInt64(&c.value))
}
```

C로 치면 GCC의 `__atomic_fetch_add` 같은 빌트인입니다.

## 8.7 Step 6 — 벤치마크 작성

`counter_test.go`:

```go
package main

import (
    "sync"
    "testing"
)

const N = 100

func runBench(b *testing.B, newCounter func() Counter) {
    for i := 0; i < b.N; i++ {
        c := newCounter()
        var wg sync.WaitGroup
        for j := 0; j < N; j++ {
            wg.Add(1)
            go func() {
                defer wg.Done()
                c.Inc()
            }()
        }
        wg.Wait()
        if closer, ok := c.(interface{ Close() }); ok {
            closer.Close()
        }
    }
}

func BenchmarkMutex(b *testing.B) {
    runBench(b, func() Counter { return &MutexCounter{} })
}

func BenchmarkChannel(b *testing.B) {
    runBench(b, func() Counter { return NewChannelCounter() })
}

func BenchmarkAtomic(b *testing.B) {
    runBench(b, func() Counter { return &AtomicCounter{} })
}
```

벤치마크 실행:

```bash
go test -bench=. -benchmem
```

전형적인 결과 (시스템마다 다름):

```
BenchmarkMutex-8     30000   42000 ns/op    288 B/op   5 allocs/op
BenchmarkChannel-8    5000  250000 ns/op   5120 B/op  10 allocs/op
BenchmarkAtomic-8    50000   28000 ns/op    160 B/op   4 allocs/op
```

## 8.8 Step 7 — 결과 해석

**일반적 경향** (N=100 단순 증가의 경우):

| 방식 | 속도 | 메모리 | 코드 복잡도 |
|---|---|---|---|
| Atomic | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ (특수 API) |
| Mutex | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ (단순) |
| Channel | ⭐⭐ | ⭐⭐ | ⭐⭐ (이해 어려움) |

### "그럼 Mutex만 쓰면 되겠네?" — 그렇지 않다!

이 결과는 **단순 카운터 증가**의 경우입니다. 채널이 빛나는 상황은 다릅니다.

## 8.9 채널이 빛나는 경우

### ① 복잡한 상태 머신

상태 전이가 여러 개 있고, 외부 입력에 반응해야 한다면 채널이 자연스럽습니다.

```go
// 예: 연결 상태 머신
for {
    select {
    case msg := <-incomingMsg:
        handleMessage(msg)
    case <-ctx.Done():
        cleanup()
        return
    case <-time.After(idleTimeout):
        sendHeartbeat()
    case cmd := <-control:
        handleControl(cmd)
    }
}
```

이걸 Mutex로 짜려면 매우 복잡해집니다.

### ② 파이프라인

데이터가 단계별로 변환되며 흐르는 구조.

```go
input → stage1 → stage2 → stage3 → output
```

채널이 각 단계를 자연스럽게 연결합니다.

### ③ 명확한 소유권 이동

"이 데이터는 이제 너 거야" 같은 의미 전달.

```go
job := makeJob()
jobsChan <- job  // 소유권을 워커에게 이동
// 여기서부터는 job을 건드리지 않음
```

## 8.10 선택 가이드라인

| 상황 | 추천 |
|---|---|
| 단순한 카운터, 플래그 | **atomic** |
| 짧은 critical section + 공유 자료구조 | **Mutex** |
| 읽기 >> 쓰기 | **RWMutex** |
| 일회성 초기화 | **Once** |
| 작업 분배 | **Channel** |
| 파이프라인, fan-in/out | **Channel** |
| 종료 신호 broadcast | **close(channel)** |
| 여러 고루틴 완료 대기 | **WaitGroup** |
| 복잡한 상태 머신 | **Channel + select** |
| 우선순위/타임아웃 처리 | **Channel + select** |

> **실무 격언**: 처음엔 Mutex로 시작하고, 코드가 복잡해지면 채널로 리팩토링하라. 채널을 무리해서 쓰면 오히려 복잡해진다.

## 8.11 race 검증

각 버전을 `-race`로 돌려 검증해보세요.

```bash
go test -race -run=^$ -bench=BenchmarkMutex
go test -race -run=^$ -bench=BenchmarkChannel
go test -race -run=^$ -bench=BenchmarkAtomic
```

세 가지 모두 race 없음이 출력됩니다.

이제 race 버전도 추가해 비교해보세요:

```go
func BenchmarkRace(b *testing.B) {
    runBench(b, func() Counter { return &RaceCounter{} })
}
```

```bash
go test -race -run=^$ -bench=BenchmarkRace
# ==================
# WARNING: DATA RACE
# ...
```

### ✅ 8교시 체크포인트

- [ ] 같은 문제를 Mutex/Channel/Atomic으로 풀어볼 수 있는가?
- [ ] 각 방식의 성능 차이를 벤치마크로 측정할 수 있는가?
- [ ] 언제 어떤 동시성 도구를 선택할지 가이드가 생겼는가?
- [ ] `-race` 플래그의 가치를 체감했는가?

---

# 🎓 3일차 마무리

## 오늘 배운 것

1. **런타임 스케줄러**: M:N 모델, G/M/P 추상화, work stealing
2. **고루틴 심화**: 동적 스택, 클로저 캡처 함정, 누수 방지
3. **Channel 기초**: Unbuffered의 랑데부, 송수신 방향, `close` + `range`
4. **Channel 심화**: Buffered, `select`, 다양한 패턴 (fan-out/in, generator)
5. **sync 패키지 I**: `Mutex`, `RWMutex`, race 탐지
6. **sync 패키지 II**: `WaitGroup`, `Once`, `Cond`, `sync.Map`
7. **실전 ①**: Producer-Consumer with graceful shutdown
8. **실전 ②**: Mutex vs Channel vs Atomic 비교

## 한 줄 요약

> **"Goroutines are about doing many things at once; channels are about how they talk to each other."** — 고루틴은 가벼운 실행 단위, 채널은 그들을 묶는 결합조직. 둘이 함께일 때 Go의 동시성이 빛납니다.

## 복습 과제

다음 시간 전에 다음을 직접 해보세요.

1. **고루틴 누수 탐지기** — `runtime.NumGoroutine()`을 1초마다 출력하는 모니터를 작성하고, 의도적으로 누수를 만들어 증가하는 모습 관찰
2. **Fan-out / Fan-in 파이프라인** — 정수 슬라이스를 받아 → 제곱하고(N개 워커 fan-out) → 결과 합치기(fan-in) → 총합 계산
3. **Read-heavy 캐시** — `RWMutex`를 쓴 캐시와 `sync.Map`을 쓴 캐시를 벤치마크로 비교
4. **Bounded worker pool** — 최대 N개 작업만 동시에 실행하는 풀 (semaphore 패턴: `make(chan struct{}, N)`)

## 다음 시간 예고 — 4일차

3일차에서 배운 동시성 도구들을 **실전 패턴**으로 확장합니다.

- **`select` 심화**: 다중 채널 처리, non-blocking 통신
- **Timeout과 취소**: `time.After`, `time.Ticker`, 타임아웃 패턴
- **Context 패키지**: 취소 전파, 데드라인, 요청 스코프 데이터
- **Worker Pool 패턴**: 동적 워커 관리, 부하 분산
- **Pipeline 패턴**: 단계별 데이터 처리, Fan-out/Fan-in 심화
- **실습**: Worker Pool 구현, 파이프라인 기반 데이터 처리

---

## 📚 참고 자료

- [Go Concurrency Patterns (Rob Pike, 2012)](https://go.dev/talks/2012/concurrency.slide) — 채널 패턴의 정전
- [Advanced Go Concurrency Patterns (Sameer Ajmani, 2013)](https://go.dev/talks/2013/advconc.slide) — `select` 심화
- [The Go Memory Model](https://go.dev/ref/mem) — 동시성 안전성의 공식 명세
- [Effective Go - Concurrency](https://go.dev/doc/effective_go#concurrency)
- [Visualizing Concurrency in Go (Ivan Daniluk)](https://divan.dev/posts/go_concurrency_visualize/) — 고루틴 동작을 시각화한 명작
- 책: 『Concurrency in Go』 (Katherine Cox-Buday, O'Reilly) — 가장 깊이 있는 채널/패턴 책

> **개인적 추천**: 위 책의 4~5장 (Concurrency Patterns)을 정독하면 Go 동시성의 90%를 마스터합니다. 영어가 부담되면 일독 후 한국어 번역본도 참고하세요.
