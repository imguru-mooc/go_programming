// stage_filter.go
package main

import (
	"context"
)

func filterErrors(ctx context.Context, in <-chan ParseResult) <-chan LogEntry {
	out := make(chan LogEntry, 100)
	go func() {
		defer close(out)
		for r := range in {
			if r.Err != nil {
				continue // 또는 별도 에러 채널
			}
			if r.Entry.Level != "ERROR" {
				continue
			}
			select {
			case out <- r.Entry:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// 시간을 시간(hour) 단위로 절삭
type HourKey struct {
	Hour string // "2024-01-15T10"
}

type HourEntry struct {
	Key   HourKey
	Entry LogEntry
}

func bucketize(ctx context.Context, in <-chan LogEntry) <-chan HourEntry {
	out := make(chan HourEntry, 100)
	go func() {
		defer close(out)
		for e := range in {
			key := HourKey{Hour: e.Timestamp.Format("2006-01-02T15")}
			select {
			case out <- HourEntry{Key: key, Entry: e}:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}
