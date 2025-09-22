// collections_plain.go
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	m := make(map[int]int)        // shared map (not thread-safe for concurrent writes)
	const G, N = 50, 1000         // 50 goroutines, 1000 writes each
	start := time.Now()

	var wg sync.WaitGroup
	wg.Add(G)
	for g := 0; g < G; g++ {
		go func(g int) {
			defer wg.Done()
			// concurrent writes to a plain map can cause data races / crashes
			for i := 0; i < N; i++ {
				m[g*N+i] = i // write distinct keys
			}
		}(g)
	}
	wg.Wait()

	elapsed := time.Since(start)
	// If it didn't crash, you might still see races or an incorrect length.
	fmt.Printf("len(m)=%d, elapsed=%s (expect %d)\n", len(m), elapsed, G*N)
}
