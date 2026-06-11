// stage_parse.go
package main

import (
	"context"
	"fmt"
	"regexp"
	"sync"
	"time"
)

// 로그 형식: 2024-01-15T10:30:45Z [LEVEL] message
var logPattern = regexp.MustCompile(`^(\S+)\s+\[(\w+)\]\s+(.*)$`)

type ParseResult struct {
	Entry LogEntry
	Err   error
}

func parseStage(ctx context.Context, in <-chan LogLine, numWorkers int) <-chan ParseResult {
	out := make(chan ParseResult, 100)
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ll := range in {
				result := parseOne(ll)
				select {
				case out <- result:
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func parseOne(ll LogLine) ParseResult {
	m := logPattern.FindStringSubmatch(ll.Line)
	if m == nil {
		return ParseResult{Err: fmt.Errorf("형식 오류: %s", ll.Line)}
	}

	ts, err := time.Parse(time.RFC3339, m[1])
	if err != nil {
		return ParseResult{Err: err}
	}

	return ParseResult{Entry: LogEntry{
		Timestamp: ts,
		Level:     m[2],
		Message:   m[3],
		Source:    ll.Source,
	}}
}
