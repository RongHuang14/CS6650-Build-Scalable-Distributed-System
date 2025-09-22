// collections_rwmutex.go
package main

import (
	"fmt"
	"sync"
	"time"
)

// SafeRWMap wraps a map with an RWMutex to allow concurrent reads and exclusive writes.
type SafeRWMap struct {
	mu sync.RWMutex // Replaced Mutex with RWMutex to allow multiple readers and single writer
	m  map[int]int
}

// Set writes to the map with an exclusive lock.
func (s *SafeRWMap) Set(k, v int) { 
	s.mu.Lock() // Lock for writing (exclusive access)
	s.m[k] = v
	s.mu.Unlock()
}

// Len returns the length of the map using a shared lock for reading.
func (s *SafeRWMap) Len() int {
	s.mu.RLock() // Lock for reading (multiple readers allowed)
	n := len(s.m)
	s.mu.RUnlock()
	return n
}

func main() {
	const G, N = 50, 1000 // 50 goroutines, 1000 writes each
	sm := &SafeRWMap{m: make(map[int]int)}

	var wg sync.WaitGroup
	wg.Add(G)

	start := time.Now()
	for g := 0; g < G; g++ {
		go func(g int) {
			defer wg.Done()
			for i := 0; i < N; i++ {
				sm.Set(g*N+i, i) // All writes are guarded by the exclusive lock
			}
		}(g)
	}
	wg.Wait()
	elapsed := time.Since(start)

	fmt.Printf("len(m)=%d (expect %d), elapsed=%s\n", sm.Len(), G*N, elapsed)
}
