package main

import (
	"fmt"
	"sync/atomic"
	"time"
)

// BackpressureMetrics — 채널이 가득 차서 Producer가 블록된 시간/횟수 집계
//
// C 비교: bounded buffer에서 pthread_cond_wait에 들어가기 전후로
// clock_gettime(CLOCK_MONOTONIC)을 찍어 대기 시간을 재는 것과 동일한 발상.
// Go에서는 select + default로 "버퍼가 가득 찼는지"를 락 없이 감지할 수 있습니다.
type BackpressureMetrics struct {
	BlockedTime  int64 // 누적 블록 시간 (nanoseconds, atomic 접근)
	BlockedCount int64 // 블록 발생 횟수
}

// AddBlock — 블록이 발생했을 때 시간과 횟수를 누적
func (m *BackpressureMetrics) AddBlock(d time.Duration) {
	atomic.AddInt64(&m.BlockedTime, int64(d))
	atomic.AddInt64(&m.BlockedCount, 1)
}

// Report — 백프레셔 발생 현황 보고
//
// 해석 기준:
//   - 자주 발생 → Consumer 수 늘리기 또는 채널 버퍼 크기 늘리기
//   - 거의 없음 → 현재 설정이 적절
func (m *BackpressureMetrics) Report() {
	blockedNs := atomic.LoadInt64(&m.BlockedTime)
	count := atomic.LoadInt64(&m.BlockedCount)

	if count == 0 {
		fmt.Println("\n백프레셔: 발생 없음 ✅")
		return
	}

	avg := time.Duration(blockedNs / count)
	total := time.Duration(blockedNs)

	fmt.Printf("\n백프레셔 보고:\n")
	fmt.Printf("  발생 횟수: %d\n", count)
	fmt.Printf("  총 블록 시간: %v\n", total)
	fmt.Printf("  평균 블록 시간: %v\n", avg)
}
