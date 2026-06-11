// types.go
package main

import "time"

type LogEntry struct {
	Timestamp time.Time
	Level     string
	Message   string
	Source    string
}

type LogLine struct {
	Source string // 어느 파일에서 왔는지
	Line   string
}
