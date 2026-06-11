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
			r := CheckWithRetry(ctx, t.URL, DefaultRetry)

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
