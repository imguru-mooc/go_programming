package main

import (
	"context"
	"fmt"
	"time"
)

func worker(ctx context.Context) {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop() // 함수가 끝날 때 타이머 리소스 확실히 해제

    for {
        select {
        case <-ctx.Done():
            fmt.Println("취소됨:", ctx.Err())
            return
        case <-ticker.C: // 이미 만들어진 틱 채널을 재사용
            fmt.Println("작업 중...")
        }
    }
}

func main() {
	fmt.Println("🚀 메인 프로그램 시작")

	// 1. 취소 기능이 탑재된 컨텍스트(ctx)와 취소 트리거 함수(cancel)를 생성합니다.
	ctx, cancel := context.WithCancel(context.Background())

	// 2. 일꾼을 백그라운드(고루틴)에서 실행합니다.
	go worker(ctx)

	// 3. 메인 프로그램은 3.5초 동안 아무것도 안 하고 기다립니다.
	// 그동안 일꾼 고루틴은 1초마다 "작업 중..."을 출력할 것입니다.
	time.Sleep(3500 * time.Millisecond)

	fmt.Println("\n📢 메인: 이제 작업 끝났으니 일꾼을 퇴근시키겠습니다.")
	
	// 4. 취소 신호를 보냅니다! (worker 내부의 ctx.Done()이 작동하게 됨)
	cancel()

	// 일꾼이 짐을 싸고(?) 완전히 종료될 때까지 잠시 대기
	time.Sleep(100 * time.Millisecond)
	fmt.Println("🏁 메인 프로그램 종료")
}
