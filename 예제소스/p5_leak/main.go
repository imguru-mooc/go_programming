package main

import (
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"time"
)

var leaked [][]byte

func leaker() {
	for {
		leaked = append(leaked, make([]byte, 100_000))
		time.Sleep(time.Second)
	}
}

func main() {
	go http.ListenAndServe(":6060", nil)
	go leaker()

	slog.Info("실행 중 - http://localhost:6060/debug/pprof/")
	select {}
}
