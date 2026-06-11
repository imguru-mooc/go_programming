package main

import (
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// Metrics — 도전 과제 ① + ③ 통합판
//
// 도전 과제 ①에서는 Consumer가 5명 고정이라 [5]int64 배열을 썼지만,
// ③의 Dynamic pool에서는 워커 수가 런타임에 변하므로 배열로는 불가능.
//
// 설계: 워커 추가 시 RegisterConsumer로 카운터(*int64)를 발급받고,
// 워커는 그 포인터를 직접 들고 atomic.AddInt64만 수행합니다.
//   - 핫패스(작업 처리마다)에는 락이 전혀 없음 — atomic 연산만
//   - 맵 접근(등록/보고)에만 mutex 사용 — 드물게 발생
//
// C 비교: per-thread 통계를 thread-local 슬롯에 쌓고 최종 보고 때만
// 락을 잡고 합산하는 고전적 기법과 같은 발상입니다.
type Metrics struct {
	Produced  [numProducers]int64 // Producer는 3명 고정 → 배열 유지
	StartTime time.Time

	mu       sync.Mutex
	consumed map[int]*int64 // workerID → 카운터 (동적 등록)
}

func NewMetrics() *Metrics {
	return &Metrics{
		StartTime: time.Now(),
		consumed:  make(map[int]*int64),
	}
}

// IncProduced — Producer가 송신 성공 직후 호출 (id: 1-based)
func (m *Metrics) IncProduced(id int) {
	atomic.AddInt64(&m.Produced[id-1], 1)
}

// RegisterConsumer — 워커 추가 시 호출. 발급된 카운터 포인터를
// 워커가 직접 보관하므로, 이후 카운트 증가에 맵 접근이 불필요
// (맵 동시 읽기/쓰기 race를 원천 차단)
func (m *Metrics) RegisterConsumer(id int) *int64 {
	m.mu.Lock()
	defer m.mu.Unlock()

	counter := new(int64)
	m.consumed[id] = counter
	return counter
}

// Report — 종료 시점 처리량 보고
func (m *Metrics) Report() {
	elapsed := time.Since(m.StartTime).Seconds()

	fmt.Println("\n=== 메트릭 보고서 ===")
	fmt.Printf("총 실행 시간: %.2fs\n\n", elapsed)

	var totalProd, totalCons int64

	fmt.Println("Producers:")
	for i := range m.Produced {
		c := atomic.LoadInt64(&m.Produced[i])
		totalProd += c
		fmt.Printf("  P%d: %d개 (%.1f/s)\n", i+1, c, float64(c)/elapsed)
	}
	fmt.Printf("  total: %d (%.1f/s)\n\n", totalProd, float64(totalProd)/elapsed)

	// 동적 워커: 맵 순회는 무작위 순서이므로 ID 정렬 후 출력
	m.mu.Lock()
	ids := make([]int, 0, len(m.consumed))
	for id := range m.consumed {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	fmt.Println("Consumers (제거된 워커 포함, 전체 이력):")
	for _, id := range ids {
		c := atomic.LoadInt64(m.consumed[id])
		totalCons += c
		fmt.Printf("  C%d: %d개 (%.1f/s)\n", id, c, float64(c)/elapsed)
	}
	m.mu.Unlock()

	fmt.Printf("  total: %d (%.1f/s)\n", totalCons, float64(totalCons)/elapsed)

	if totalProd != totalCons {
		fmt.Printf("⚠️  미처리: %d개\n", totalProd-totalCons)
	}
}
