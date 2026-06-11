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
	"encoding/json"

	"healthcheck/internal/checker"
)

type JSONOutput struct {
	Results []JSONResult `json:"results"`
	Summary JSONSummary  `json:"summary"`
}

type JSONResult struct {
	URL        string `json:"url"`
	StatusCode int    `json:"status_code,omitempty"`
	LatencyMs  int64  `json:"latency_ms,omitempty"`
	Error      string `json:"error,omitempty"`
	Attempts   int    `json:"attempts"`
}

type JSONSummary struct {
	Total     int `json:"total"`
	Succeeded int `json:"succeeded"`
	Failed    int `json:"failed"`
}

func main() {
	var (
		workers = flag.Int("w", 5, "동시 워커 수")
		timeout = flag.Duration("t", 30*time.Second, "전체 타임아웃")
		jsonOut   = flag.Bool("json", false, "JSON 형식 출력")
	)
	flag.Parse()

	urls := flag.Args()
	if len(urls) == 0 {
		urls = []string{
			"https://go.dev",
			"https://github.com",
			"https://google.com",
			"https://example.com",
			"https://naver.com",
			"https://daum.com",
			"https://www.kakaocorp.com",
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

	var allResults []checker.Result
	for r := range pool.Results() {
		allResults = append(allResults, r)
	}

	if *jsonOut {
		var out JSONOutput
		for _, r := range allResults {
			jr := JSONResult{
				URL:      r.URL,
				Attempts: r.Attempts,
			}
			if r.Err != nil {
				jr.Error = r.Err.Error()
				out.Summary.Failed++
			} else {
				jr.StatusCode = r.StatusCode
				jr.LatencyMs = r.Latency.Milliseconds()
				out.Summary.Succeeded++
			}
			out.Results = append(out.Results, jr)
		}
		out.Summary.Total = len(allResults)
		json.NewEncoder(os.Stdout).Encode(out)
	} else {
		for _, r := range allResults {
			fmt.Println(r)
		}
	}

	// 성공한 요청들의 레이턴시만 수집
	latencies := make([]time.Duration, 0)
	for _, r := range allResults {
		if r.Err == nil {
			latencies = append(latencies, r.Latency)
		}
	}

	if len(latencies) > 0 {
		p := checker.Calculate(latencies)
		fmt.Printf("\n=== 응답 시간 통계 ===\n")
		fmt.Printf("  Min:  %v\n", p.Min)
		fmt.Printf("  P50:  %v\n", p.P50)
		fmt.Printf("  Mean: %v\n", p.Mean)
		fmt.Printf("  P95:  %v\n", p.P95)
		fmt.Printf("  P99:  %v\n", p.P99)
		fmt.Printf("  Max:  %v\n", p.Max)
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
