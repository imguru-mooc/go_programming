package main

import (
	"fmt"
	"sync"
	"time"
)

// 1. WaitGroup - 병렬 합계
func parallelSum(nums []int) int {
	n := len(nums)
	if n == 0 {
		return 0
	}
	mid := n / 2

	var wg sync.WaitGroup
	var left, right int

	wg.Add(2)
	go func() {
		defer wg.Done()
		for _, v := range nums[:mid] {
			left += v
		}
	}()
	go func() {
		defer wg.Done()
		for _, v := range nums[mid:] {
			right += v
		}
	}()
	wg.Wait()

	return left + right
}

// 2. Once - 게으른 초기화
type Config struct {
	Loaded time.Time
	Value  string
}

var (
	cfgOnce sync.Once
	cfg     *Config
)

func GetConfig() *Config {
	cfgOnce.Do(func() {
		fmt.Println("Config 로딩 중... (단 한 번만 출력)")
		time.Sleep(100 * time.Millisecond)
		cfg = &Config{
			Loaded: time.Now(),
			Value:  "production-config",
		}
	})
	return cfg
}

func main() {
	// WaitGroup 테스트
	nums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	fmt.Println("합계:", parallelSum(nums))

	// Once 테스트 - 동시 호출
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c := GetConfig()
			_ = c
		}()
	}
	wg.Wait()
	fmt.Println("Config 로딩 시각:", cfg.Loaded)
}
