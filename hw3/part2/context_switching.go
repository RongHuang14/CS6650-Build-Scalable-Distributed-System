// context_switching.go
package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

func main() {
	const pingPongs = 1000000
	
	// TEST 1: Single OS thread
	fmt.Println("TEST 1: GOMAXPROCS(1) - Single OS thread")
	runtime.GOMAXPROCS(1)
	
	ping := make(chan struct{})
	pong := make(chan struct{})
	
	var wg sync.WaitGroup
	wg.Add(2)
	
	start := time.Now()
	
	// Goroutine 1: sends ping, waits for pong
	go func() {
		defer wg.Done()
		for i := 0; i < pingPongs; i++ {
			ping <- struct{}{}
			<-pong
		}
	}()
	
	// Goroutine 2: waits for ping, sends pong
	go func() {
		defer wg.Done()
		for i := 0; i < pingPongs; i++ {
			<-ping
			pong <- struct{}{}
		}
	}()
	
	wg.Wait()
	singleThreadTime := time.Since(start)
	avgSwitch1 := singleThreadTime / (2 * pingPongs) // 2 switches per round trip
	
	fmt.Printf("Total time: %v\n", singleThreadTime)
	fmt.Printf("Average switch time: %v\n", avgSwitch1)
	
	// TEST 2: Multiple OS threads
	fmt.Println("\nTEST 2: Default GOMAXPROCS - Multiple OS threads")
	runtime.GOMAXPROCS(runtime.NumCPU()) // Use all CPU cores
	
	ping2 := make(chan struct{})
	pong2 := make(chan struct{})
	
	wg.Add(2)
	
	start = time.Now()
	
	go func() {
		defer wg.Done()
		for i := 0; i < pingPongs; i++ {
			ping2 <- struct{}{}
			<-pong2
		}
	}()
	
	go func() {
		defer wg.Done()
		for i := 0; i < pingPongs; i++ {
			<-ping2
			pong2 <- struct{}{}
		}
	}()
	
	wg.Wait()
	multiThreadTime := time.Since(start)
	avgSwitch2 := multiThreadTime / (2 * pingPongs)
	
	fmt.Printf("Total time: %v\n", multiThreadTime)
	fmt.Printf("Average switch time: %v\n", avgSwitch2)
	
	// Comparison
	fmt.Println("\n=== COMPARISON ===")
	fmt.Printf("Single thread avg: %v\n", avgSwitch1)
	fmt.Printf("Multi thread avg:  %v\n", avgSwitch2)
	if avgSwitch1 < avgSwitch2 {
		fmt.Printf("Single thread is %.2fx FASTER\n", float64(avgSwitch2)/float64(avgSwitch1))
	} else {
		fmt.Printf("Multi thread is %.2fx FASTER\n", float64(avgSwitch1)/float64(avgSwitch2))
	}
}