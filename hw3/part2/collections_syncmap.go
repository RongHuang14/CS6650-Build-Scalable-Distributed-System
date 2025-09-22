// collections_syncmap.go
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var m sync.Map // Changed from a regular map to sync.Map to ensure safe concurrent access
	const G, N = 50, 1000 // 50 goroutines, each writing 1000 entries

	start := time.Now()

	var wg sync.WaitGroup
	wg.Add(G)
	for g := 0; g < G; g++ {
		go func(g int) {
			defer wg.Done()
			// Using sync.Map's Store method to store key-value pairs instead of direct map assignment
			for i := 0; i < N; i++ {
				m.Store(g*N+i, i) // Store key-value pairs in sync.Map
			}
		}(g)
	}
	wg.Wait()

	// Using sync.Map's Range method to count the entries in the map
	var count int
	m.Range(func(key, value interface{}) bool {
		count++ // Increment count for each entry in sync.Map
		return true
	})

	elapsed := time.Since(start)
	// Print the total number of entries and the elapsed time
	fmt.Printf("len(m)=%d (expect %d), elapsed=%s\n", count, G*N, elapsed)
}
