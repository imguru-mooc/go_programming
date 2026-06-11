package main

import (
	"testing"
	"time"
)

// 테스트를 위해 패키지 레벨 변수로 설정
var notify = make(chan struct{}, 1)

func notifyOnce() {
	select {
	case notify <- struct{}{}:
		// 알림 전송됨
	default:
		// 이미 알림이 큐에 있음 - 무시
	}
}

func TestNotifyOnce_Behavior(t *testing.T) {
	// [준비] 테스트 시작 전 채널 비우기
	for len(notify) > 0 {
		<-notify
	}

	// 1. 첫 번째 알림 발송 -> 채널이 비어있으므로 정상적으로 들어가야 함 (크기 1)
	notifyOnce()
	if len(notify) != 1 {
		t.Errorf("첫 번째 알림 후 채널 크기는 1이어야 하지만, 실제 크기: %d", len(notify))
	}

	// 2. 두 번째 알림 발송 -> 채널이 꽉 찼으므로 default로 빠져서 무시되어야 함 (크기 유지 1)
	notifyOnce()
	if len(notify) != 1 {
		t.Errorf("중복된 알림이 버려지지 않고 채널에 쌓였습니다. 실제 크기: %d", len(notify))
	}

	// 3. 소비자가 알림을 하나 처리함 (채널 비우기)
	<-notify
	if len(notify) != 0 {
		t.Errorf("알림을 소비한 후 채널 크기는 0이어야 하지만, 실제 크기: %d", len(notify))
	}

	// 4. 세 번째 알림 발송 -> 채널이 다시 비었으므로 정상적으로 들어가야 함 (크기 1)
	notifyOnce()
	if len(notify) != 1 {
		t.Errorf("채널이 비었을 때의 재발송이 실패했습니다. 실제 크기: %d", len(notify))
	}

	// 청소
	<-notify
}

// 5. 실제 비동기 고루틴 환경에서 안전하게 작동하는지 테스트
func TestNotifyOnce_Concurrency(t *testing.T) {
	// 생산자 고루틴이 짧은 시간 동안 알림을 100번 미친듯이 보냄
	for i := 0; i < 100; i++ {
		go notifyOnce()
	}

	// 고루틴들이 실행될 시간을 아주 잠깐 제공
	time.Sleep(10 * time.Millisecond)

	// 아무리 동시에 100번을 보냈어도 버퍼 크기가 1이기 때문에 
	// 채널에는 절대 1개 초과의 신호가 쌓여있을 수 없음 (오버로드 방지 검증)
	if len(notify) > 1 {
		t.Errorf("동시성 환경에서 버퍼 오버로드가 발생했습니다. 크기: %d", len(notify))
	}
}
