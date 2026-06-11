package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ============================================================
// 도전 과제 ① + ② + ③ 통합 데모
//   ① 메트릭     — Producer/Consumer별 처리량 집계 (metrics.go)
//   ② 백프레셔   — Producer 블록 시간 측정 (backpressure.go)
//   ③ 동적 풀    — Consumer 런타임 증감 (dynamic_pool.go)
//
// 시나리오: 워커 2명으로 시작 → 백프레셔 발생 → 워커 증원으로 해소
//           → 워커 감축 → 백프레셔 재발. 메트릭으로 전 과정 관측.
// ============================================================

// LogEntry — 7교시 강의자료와 동일한 작업 단위
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

const (
	numProducers = 3
	bufferSize   = 10
	runDuration  = 8 * time.Second
)

// producer — 도전 과제 ②의 백프레셔 측정 패턴 적용
// (non-blocking 송신 먼저 시도, 실패 시 시간 측정 후 blocking 송신)
func producer(id int, out chan<- LogEntry, done <-chan struct{},
	m *Metrics, bp *BackpressureMetrics, wg *sync.WaitGroup) {
	defer wg.Done()
	defer fmt.Printf("[Producer %d] 종료\n", id)

	seq := 0
	for {
		entry := LogEntry{
			ProducerID: id,
			Sequence:   seq,
			Message:    fmt.Sprintf("log event %d", seq),
			Timestamp:  time.Now(),
		}

		select {
		case out <- entry:
			// 즉시 성공 — 블록 없음
			m.IncProduced(id)
			seq++
		case <-done:
			return
		default:
			// 채널 가득 참 — 블록 시간 측정
			blockStart := time.Now()
			select {
			case out <- entry:
				bp.AddBlock(time.Since(blockStart))
				m.IncProduced(id)
				seq++
			case <-done:
				return
			}
		}

		// 생산 속도: 3명 × 100~200ms → 약 20개/s
		// (워커 1명 = 5개/s 처리이므로: 워커 2명이면 백프레셔 발생,
		//  4명이면 균형 — 스케일링 효과가 메트릭에 드러나도록 설계)
		time.Sleep(time.Duration(100+rand.Intn(100)) * time.Millisecond)
	}
}

func main() {
	jobs := make(chan LogEntry, bufferSize)
	done := make(chan struct{})

	metrics := NewMetrics()
	bp := &BackpressureMetrics{}
	pool := NewDynamicPool(jobs, metrics)

	// 시작: 워커 2명 (처리 능력 10/s < 생산 20/s → 백프레셔 유발)
	pool.AddWorker()
	pool.AddWorker()

	var prodWg sync.WaitGroup
	for i := 1; i <= numProducers; i++ {
		prodWg.Add(1)
		go producer(i, jobs, done, metrics, bp, &prodWg)
	}

	// === 런타임 스케일링 시나리오 ===

	// 3초 후: 백프레셔 누적 확인 → 워커 증원 (2 → 4)
	time.Sleep(3 * time.Second)
	fmt.Printf("\n--- [3s] 큐 길이 %d/%d, 워커 %d명 → 증원 ---\n",
		len(jobs), cap(jobs), pool.Size())
	pool.AddWorker()
	pool.AddWorker()

	// 3초 후: 워커 감축 (4 → 3) → 백프레셔 일부 재발
	time.Sleep(3 * time.Second)
	fmt.Printf("\n--- [6s] 큐 길이 %d/%d, 워커 %d명 → 감축 ---\n",
		len(jobs), cap(jobs), pool.Size())
	pool.RemoveWorker()

	time.Sleep(runDuration - 6*time.Second)

	// === Graceful shutdown (7교시 순서 그대로) ===
	// done close → producers 대기 → jobs close(워커 broadcast) → 워커 대기
	fmt.Println("\n--- 종료 시퀀스 시작 ---")
	close(done)
	prodWg.Wait()
	close(jobs)
	pool.Wait()

	// 도전 과제 ① + ② 보고서 출력
	metrics.Report()
	bp.Report()
}
