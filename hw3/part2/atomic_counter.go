package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	// atomic counter
	var ops atomic.Uint64
	var wg sync.WaitGroup

	workers := 50
	loops := 1000

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < loops; j++ {
				ops.Add(1)
			}
		}()
	}

	wg.Wait()
	fmt.Println("ops:", ops.Load())
}
