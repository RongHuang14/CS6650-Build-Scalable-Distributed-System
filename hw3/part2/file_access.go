// file_access.go
package main

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

func main() {
	const iterations = 100000
	data := "This is a test line for file writing experiment\n"
	
	// UNBUFFERED writes
	fmt.Println("Testing UNBUFFERED file writes...")
	f1, err := os.Create("unbuffered.txt")
	if err != nil {
		panic(err)
	}
	
	start := time.Now()
	for i := 0; i < iterations; i++ {
		f1.Write([]byte(data)) // Direct write to file
	}
	f1.Close()
	unbufferedTime := time.Since(start)
	
	// BUFFERED writes
	fmt.Println("Testing BUFFERED file writes...")
	f2, err := os.Create("buffered.txt")
	if err != nil {
		panic(err)
	}
	
	w := bufio.NewWriter(f2)
	start = time.Now()
	for i := 0; i < iterations; i++ {
		w.WriteString(data) // Write to buffer
	}
	w.Flush() // Flush buffer to file
	f2.Close()
	bufferedTime := time.Since(start)
	
	// Print results
	fmt.Printf("Unbuffered: %v\n", unbufferedTime)
	fmt.Printf("Buffered:   %v\n", bufferedTime)
	fmt.Printf("Speedup:    %.2fx faster\n", float64(unbufferedTime)/float64(bufferedTime))
	
	// Cleanup
	os.Remove("unbuffered.txt")
	os.Remove("buffered.txt")
}