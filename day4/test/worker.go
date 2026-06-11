package main

import (
	"fmt"
	"time"
	"sync"
)
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
