package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// LogEntry — 처리할 작업 단위
type LogEntry struct {
	ProducerID int
	Sequence   int
	Message    string
	Timestamp  time.Time
}

func (l LogEntry) String() string {
	return fmt.Sprintf("[P%d-#%d %s] %s",
		l.ProducerID, l.Sequence, l.Timestamp.Format("15:04:05.000"), l.Message)
}

func producer(id int, out chan<- LogEntry, done <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	defer fmt.Printf("[Producer %d] 종료\n", id)

	seq := 0
	for {
		select {
		case <-done:
			return
		default:
		}

		// 0~200ms 사이 랜덤 대기 (실제 로그 생성 흉내)
		time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)

		entry := LogEntry{
			ProducerID: id,
			Sequence:   seq,
			Message:    fmt.Sprintf("event-%d", seq),
			Timestamp:  time.Now(),
		}
		seq++

		// 송신할 때도 done 확인 - 채널이 가득 차면 빠져나갈 수 있도록
		select {
		case out <- entry:
		case <-done:
			return
		}
	}
}

func consumer(id int, in <-chan LogEntry, wg *sync.WaitGroup) {
	defer wg.Done()
	defer fmt.Printf("[Consumer %d] 종료\n", id)

	for entry := range in { // 채널 닫히면 자동 종료
		// 처리 시간 흉내 (50~150ms)
		time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)
		fmt.Printf("[Consumer %d] %s\n", id, entry)
	}
}

const (
	numProducers    = 3
	numConsumers    = 5
	channelCapacity = 10
)

func main() {
	rand.Seed(time.Now().UnixNano())

	logChan := make(chan LogEntry, channelCapacity)
	done := make(chan struct{})

	// 시그널 핸들러
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	var producerWG, consumerWG sync.WaitGroup

	// Producer 시작
	for i := 1; i <= numProducers; i++ {
		producerWG.Add(1)
		go producer(i, logChan, done, &producerWG)
	}

	// Consumer 시작
	for i := 1; i <= numConsumers; i++ {
		consumerWG.Add(1)
		go consumer(i, logChan, &consumerWG)
	}

	fmt.Printf("실행 중... (Ctrl+C로 종료)\n")

	// 시그널 대기
	<-sigs
	fmt.Println("\n종료 신호 수신 — 정리 중...")

	// 1) Producer들에게 종료 신호
	close(done)
	producerWG.Wait()
	fmt.Println("모든 Producer 종료됨")

	// 2) 채널 닫기 - 더 이상 송신 없음
	close(logChan)

	// 3) Consumer는 남은 데이터 처리 후 자동 종료
	consumerWG.Wait()
	fmt.Println("모든 Consumer 종료됨")

	fmt.Println("정상 종료")
}
