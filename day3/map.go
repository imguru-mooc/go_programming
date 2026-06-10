package main

import (
	"fmt"
	"sync"
)

func main() {
	var m sync.Map
	var wg sync.WaitGroup

	fmt.Println("1. 여러 고루틴에서 동시에 sync.Map에 데이터 쓰기 시작")
	fmt.Println("------------------------------------------------")

	// 5개의 고루틴을 띄워 동시에 맵에 데이터를 저장합니다.
	for i := 1; i <= 5; i++ {
		wg.Add(1) // 고루틴 개수 추가

		go func(id int) {
			defer wg.Done() // 고루틴 종료 시 대기 카운트 감소

			key := fmt.Sprintf("key%d", id)
			value := id * 100

			// 동시성 안전하게 데이터 저장
			m.Store(key, value)
			fmt.Printf("[고루틴 %d] 저장 작업 수행: %s -> %d\n", id, key, value)
		}(i)
	}

	// 모든 고루틴이 쓰기 작업을 마칠 때까지 기다립니다.
	wg.Wait()
	fmt.Println("------------------------------------------------")
	fmt.Println("2. 모든 동시성 쓰기 작업 완료\n")

	// 3. 데이터 단건 조회 (Load) 테스트
	fmt.Println("3. 특정 키 조회 (Load) 테스트")
	v, ok := m.Load("key3")
	if ok {
		// sync.Map에서 꺼낸 값은 interface{} 타입이므로 필요 시 형변환(Type Assertion)을 합니다.
		actualValue := v.(int)
		fmt.Printf("-> 조회 성공! key3의 값은: %d\n\n", actualValue)
	} else {
		fmt.Println("-> 값을 찾지 못했습니다.")
	}

	// 4. 전체 데이터 순회 (Range) 테스트
	fmt.Println("4. 전체 데이터 순회 (Range) 테스트")
	m.Range(func(k, v interface{}) bool {
		fmt.Printf("-> [보관 데이터] 키: %v, 값: %v\n", k, v)
		return true // true를 리턴해야 멈추지 않고 다음 데이터로 넘어갑니다.
	})
}
