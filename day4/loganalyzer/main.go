// main.go
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"
)

func main() {
	paths := os.Args[1:]
	if len(paths) == 0 {
		paths = []string{"-"} // stdin
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Fprintln(os.Stderr, "\n중단 — 부분 결과 출력")
		cancel()
	}()

	start := time.Now()

	// 파이프라인 조립
	lines := readFiles(ctx, paths)
	parsed := parseStage(ctx, lines, 4) // 4개 파서
	errors := filterErrors(ctx, parsed)
	bucketed := bucketize(ctx, errors)
	stats := aggregate(ctx, bucketed)

	// 결과 출력
	elapsed := time.Since(start)
	fmt.Printf("\n=== ERROR 로그 시간대별 집계 ===\n")

	hours := make([]string, 0, len(stats))
	for h := range stats {
		hours = append(hours, h)
	}
	sort.Strings(hours)

	total := 0
	for _, h := range hours {
		fmt.Printf("%s : %d건\n", h, stats[h])
		total += stats[h]
	}
	fmt.Printf("\n총 ERROR: %d건 (소요: %v)\n", total, elapsed)
}
