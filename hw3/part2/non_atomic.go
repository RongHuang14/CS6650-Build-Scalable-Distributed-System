package main

import (
	"fmt"
	"sync"
)

func main() {
	// normal counter (not atomic)
	var ops uint64
	var wg sync.WaitGroup

	workers := 50
	loops := 1000

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < loops; j++ {
				ops++ // not safe: race condition
			}
		}()
	}

	wg.Wait()
	fmt.Printf("non-atomic ops (expect %d): %d\n", workers*loops, ops)
}
