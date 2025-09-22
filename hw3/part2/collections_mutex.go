// collections_mutex.go
package main

import (
	"fmt"
	"sync"
	"time"
)

// SafeMap wraps a map with a mutex for safe concurrent access.
type SafeMap struct {
	mu sync.Mutex
	m  map[int]int
}

func (s *SafeMap) Set(k, v int) {
	s.mu.Lock()
	s.m[k] = v
	s.mu.Unlock()
}

func (s *SafeMap) Len() int {
	s.mu.Lock()
	n := len(s.m)
	s.mu.Unlock()
	return n
}

func main() {
	const G, N = 50, 1000 // 50 goroutines, 1000 writes each

	sm := &SafeMap{m: make(map[int]int)}
	var wg sync.WaitGroup
	wg.Add(G)

	start := time.Now()
	for g := 0; g < G; g++ {
		go func(g int) {
			defer wg.Done()
			for i := 0; i < N; i++ {
				sm.Set(g*N+i, i) // guarded write
			}
		}(g)
	}
	wg.Wait()
	elapsed := time.Since(start)

	fmt.Printf("len(m)=%d (expect %d), elapsed=%s\n", sm.Len(), G*N, elapsed)
}
