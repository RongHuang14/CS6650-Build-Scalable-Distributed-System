package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
)

// MapRequest contains text chunk to process
type MapRequest struct {
    Text string `json:"text"`
}

// MapResponse contains word count results
type MapResponse struct {
    WordCount   map[string]int `json:"word_count"`
    TotalWords  int           `json:"total_words"`
    UniqueWords int           `json:"unique_words"`
}

// countWords processes text and counts word occurrences
func countWords(text string) map[string]int {
    wordCount := make(map[string]int)
    
    // Convert to lowercase and split into words
    text = strings.ToLower(text)
    words := strings.Fields(text)
    
    // Count each word
    for _, word := range words {
        // Clean punctuation (simple version)
        word = strings.Trim(word, ".,!?;:")
        if word != "" {
            wordCount[word]++
        }
    }
    
    return wordCount
}

// handleMap processes the map request
func handleMap(w http.ResponseWriter, r *http.Request) {
    // Parse request
    var req MapRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    // Count words in the text chunk
    wordCount := countWords(req.Text)
    
    // Calculate totals
    totalWords := 0
    for _, count := range wordCount {
        totalWords += count
    }
    
    // Send response
    response := MapResponse{
        WordCount:   wordCount,
        TotalWords:  totalWords,
        UniqueWords: len(wordCount),
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func main() {
    fmt.Println("Mapper service starting on port 8080...")
    
    http.HandleFunc("/map", handleMap)
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("OK"))
    })
    
    if err := http.ListenAndServe(":8080", nil); err != nil {
        fmt.Printf("Server error: %v\n", err)
    }
}