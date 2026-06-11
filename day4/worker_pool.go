package main

import (
	"fmt"
	"time"
	"sync"
	"context"
)

type Job struct {
    ID   int
    Data string
}

type Result struct {
    JobID  int
    Output string
}

func process(ctx context.Context, workerID int, job Job) Result {
	time.Sleep(100 * time.Millisecond) // 비즈니스 로직 시뮬레이션
	output := fmt.Sprintf("worker-%d processed job-%d", workerID, job.ID)
	return Result{JobID: job.ID, Output: output}
}

func worker(ctx context.Context, id int, jobs <-chan Job, results chan<- Result) {
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("worker %d: 취소됨\n", id)
			return
		case job, ok := <-jobs:
			if !ok {
				return // 채널 닫힘
			}

			select {
			case results <- process(ctx, id, job):
			case <-ctx.Done():
				fmt.Printf("worker %d: 결과 송신 중 취소됨\n", id)
				return
			}
		}
	}
}


func main() {
	const numWorkers = 5
	const numJobs = 20

	// 1. 타임아웃이 있는 컨텍스트를 생성합니다.
	// 💡 실습 팁: 원래 모든 작업을 끝내려면 약 0.4초가 걸립니다.
	// 아래 시간을 200*time.Millisecond로 줄이면 중간에 워커들이 취소되는 모습을 볼 수 있습니다!
	ctx, cancel := context.WithTimeout(context.Background(), 2000*time.Millisecond)
	defer cancel()

	jobs := make(chan Job)
	results := make(chan Result, numJobs)

	// 2. 워커 시작 (ctx를 매개변수로 전달)
	var wg sync.WaitGroup
	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			worker(ctx, id, jobs, results) // 매개변수 추가
		}(i)
	}

	// 3. 작업 송신 (송신 중에도 컨텍스트가 취소되면 중단하도록 보완)
	go func() {
		for j := 1; j <= numJobs; j++ {
			select {
			case jobs <- Job{ID: j, Data: fmt.Sprintf("data-%d", j)}:
			case <-ctx.Done():
				close(jobs)
				return
			}
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
