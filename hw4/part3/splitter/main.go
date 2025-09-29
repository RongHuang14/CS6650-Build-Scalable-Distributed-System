package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// SplitRequest represents the incoming request with text to split
type SplitRequest struct {
	Text string `json:"text"`
	Chunks int    `json:"chunks"`
}

// SplitResponse contains the split text chunks
type SplitResponse struct {
	Chunks []string `json:"chunks"`
	Count int `json:"total_chunks"`
}

// splitText divides text into roughly equal chunks
func splitText(text string, numChunks int) []string {
	words := strings.Fields(text)

	// Calculate chunk size
	totalWords := len(words)
	chunkSize := totalWords / numChunks
    if chunkSize == 0 {
        chunkSize = 1
    }

	chunks := make([]string, 0, numChunks)
	for i := 0; i < numChunks; i++ {
		start := i * chunkSize
		end := start + chunkSize

		// Last chunk gets remaining words
		if i == numChunks-1 {
			end = totalWords
		}
        if start < totalWords {
            chunk := strings.Join(words[start:end], " ")
            chunks = append(chunks, chunk)
        }
	}
	return chunks
}

// handleSplit processes the split request
func handleSplit(w http.ResponseWriter, r *http.Request) {
    // Parse request
    var req SplitRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    // Default to 3 chunks if not specified
    if req.Chunks == 0 {
        req.Chunks = 3
    }
    
    // Split the text
    chunks := splitText(req.Text, req.Chunks)
    
    // Send response
    response := SplitResponse{
        Chunks: chunks,
        Count:  len(chunks),
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func main() {
    fmt.Println("Splitter service starting on port 8080...")
    
    http.HandleFunc("/split", handleSplit)
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("OK"))
    })
    
    if err := http.ListenAndServe(":8080", nil); err != nil {
        fmt.Printf("Server error: %v\n", err)
    }
}

	