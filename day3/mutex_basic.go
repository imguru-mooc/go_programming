package main

import (
	"fmt"
	"sync"
)

// 안전한 카운터
type Counter struct {
	mu    sync.Mutex
	value int
}

func (c *Counter) Inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value++
}

func (c *Counter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}

// 안전한 캐시
type Cache struct {
	rw   sync.RWMutex
	data map[string]string
}

func NewCache() *Cache {
	return &Cache{data: make(map[string]string)}
}

func (c *Cache) Get(key string) (string, bool) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	v, ok := c.data[key]
	return v, ok
}

func (c *Cache) Set(key, value string) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.data[key] = value
}

func main() {
	// 카운터 테스트
	var wg sync.WaitGroup
	c := &Counter{}

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Inc()
		}()
	}
	wg.Wait()
	fmt.Println("카운터:", c.Value()) // 정확히 1000

	// 캐시 테스트
	cache := NewCache()
	cache.Set("name", "Alice")
	cache.Set("age", "30")

	if v, ok := cache.Get("name"); ok {
		fmt.Println("name =", v)
	}
}
