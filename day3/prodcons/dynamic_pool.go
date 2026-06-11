package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// DynamicPool — 도전 과제 ③ + 메트릭(①) 통합판
//
// 워커마다 개별 quit 채널을 두어 런타임 추가/제거를 지원하고,
// AddWorker 시 Metrics에서 발급받은 카운터를 워커에게 넘깁니다.
type DynamicPool struct {
	mu      sync.Mutex
	jobs    <-chan LogEntry
	metrics *Metrics
	workers map[int]chan struct{} // workerID → quit channel
	nextID  int
	wg      sync.WaitGroup
}

func NewDynamicPool(jobs <-chan LogEntry, m *Metrics) *DynamicPool {
	return &DynamicPool{
		jobs:    jobs,
		metrics: m,
		workers: make(map[int]chan struct{}),
		nextID:  1,
	}
}

// AddWorker — 새 워커 추가. 메트릭 카운터를 발급받아 워커에 전달
func (p *DynamicPool) AddWorker() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	id := p.nextID
	p.nextID++

	quit := make(chan struct{})
	p.workers[id] = quit

	counter := p.metrics.RegisterConsumer(id)

	p.wg.Add(1)
	go p.workerLoop(id, quit, counter)

	fmt.Printf("[Pool] 워커 %d 추가 (총 %d명)\n", id, len(p.workers))
	return id
}

// RemoveWorker — 가장 최근 워커(최대 ID) 제거.
// 제거되어도 해당 워커의 처리 이력은 Metrics에 남습니다.
func (p *DynamicPool) RemoveWorker() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.workers) == 0 {
		return
	}

	var lastID int
	for id := range p.workers {
		if id > lastID {
			lastID = id
		}
	}

	close(p.workers[lastID])
	delete(p.workers, lastID)
	fmt.Printf("[Pool] 워커 %d 제거 요청 (총 %d명)\n", lastID, len(p.workers))
}

// Size — 현재 워커 수 (스케일링 판단용)
func (p *DynamicPool) Size() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.workers)
}

// CloseAll — 남은 모든 워커에 종료 신호
func (p *DynamicPool) CloseAll() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for id, quit := range p.workers {
		close(quit)
		delete(p.workers, id)
	}
}

// Wait — 모든 워커 고루틴 종료 대기
func (p *DynamicPool) Wait() {
	p.wg.Wait()
}

// workerLoop — 개별 워커. counter는 자기 전용이므로 락 없이
// atomic 연산만으로 집계 (IncConsumed 헬퍼 대신 직접 보유 방식)
func (p *DynamicPool) workerLoop(id int, quit <-chan struct{}, counter *int64) {
	defer p.wg.Done()
	defer fmt.Printf("[Worker %d] 종료\n", id)

	for {
		select {
		case <-quit:
			return
		case entry, ok := <-p.jobs:
			if !ok {
				return // jobs close = 전체 워커 broadcast 종료
			}
			// 처리 시뮬레이션 (200ms)
			time.Sleep(200 * time.Millisecond)
			_ = entry

			atomic.AddInt64(counter, 1) // 자기 전용 카운터에 락 없이 누적
		}
	}
}
