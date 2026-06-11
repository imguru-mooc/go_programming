// stage_aggregate.go
package main

import "context"

func aggregate(ctx context.Context, in <-chan HourEntry) map[string]int {
	counts := make(map[string]int)
	for he := range in {
		select {
		case <-ctx.Done():
			return counts
		default:
		}
		counts[he.Key.Hour]++
	}
	return counts
}
