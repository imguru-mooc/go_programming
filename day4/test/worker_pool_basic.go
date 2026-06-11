
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
	TaskID   int
	WorkerID int // 💡 어떤 워커가 처리했는지 추적하기 위해 필드 추가!
	Result   string
	Err      error
}

// process 함수에 workerID 매개변수 추가
func process(ctx context.Context, workerID int, t Task) Outcome {
	select {
	case <-time.After(200 * time.Millisecond):
		return Outcome{
			TaskID:   t.ID,
			WorkerID: workerID, // 워커 ID 기록
			Result:   fmt.Sprintf("done-%d", t.ID),
		}
	case <-ctx.Done():
		return Outcome{
			TaskID:   t.ID,
			WorkerID: workerID, // 취소 시점의 워커 ID 기록
			Err:      ctx.Err(),
		}
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
			// process 함수 호출 시 자신의 id를 넘겨줍니다.
			o := process(ctx, id, t)
			select {
			case results <- o:
			case <-ctx.Done():
				return
			}
		}
	}
}

func main() {
	// 실습 팁: 이 시간을 1초(time.Second)로 줄이면 실패(❌) 로그가 찍히기 시작합니다!
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

	// 결과 집계 및 실습 출력 파트
	var (
		completed int64
		failed    int64
	)
	
	fmt.Println("=== ⏳ 실시간 워커 처리 결과 수집 시작 ===")
	for o := range results {
		if o.Err != nil {
			atomic.AddInt64(&failed, 1)
			// 💡 실패 시 에러 사유와 담당 워커 출력
			fmt.Printf("[워커 %d] ❌ 작업-%02d 처리 실패: %v\n", o.WorkerID, o.TaskID, o.Err)
		} else {
			atomic.AddInt64(&completed, 1)
			// 💡 성공 시 결과물과 담당 워커 출력
			fmt.Printf("[워커 %d] ✅ 작업-%02d 처리 완료 -> %s\n", o.WorkerID, o.TaskID, o.Result)
		}
	}

	fmt.Println("=======================================")
	fmt.Printf("🏁 최종 결과 요약: 완료=%d 실패=%d\n", completed, failed)
}
