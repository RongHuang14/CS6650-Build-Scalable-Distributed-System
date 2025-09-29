package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "sort"
)

// ReduceRequest contains multiple word count results to merge
type ReduceRequest struct {
    Results []map[string]int `json:"results"`
}

// WordCount represents a word and its count for sorting
type WordCount struct {
    Word  string `json:"word"`
    Count int    `json:"count"`
}

// ReduceResponse contains the final aggregated results
type ReduceResponse struct {
    FinalCount  map[string]int `json:"final_count"`
    TotalWords  int           `json:"total_words"`
    UniqueWords int           `json:"unique_words"`
    TopWords    []WordCount   `json:"top_10_words"`
}

// mergeResults combines multiple word count maps
func mergeResults(results []map[string]int) map[string]int {
    finalCount := make(map[string]int)
    
    // Merge all results
    for _, result := range results {
        for word, count := range result {
            finalCount[word] += count
        }
    }
    
    return finalCount
}

// getTopWords returns the top N most frequent words
func getTopWords(wordCount map[string]int, n int) []WordCount {
    // Convert map to slice for sorting
    words := make([]WordCount, 0, len(wordCount))
    for word, count := range wordCount {
        words = append(words, WordCount{Word: word, Count: count})
    }
    
    // Sort by count (descending)
    sort.Slice(words, func(i, j int) bool {
        return words[i].Count > words[j].Count
    })
    
    // Return top N
    if len(words) > n {
        return words[:n]
    }
    return words
}

// handleReduce processes the reduce request
func handleReduce(w http.ResponseWriter, r *http.Request) {
    // Parse request
    var req ReduceRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    // Merge all word counts
    finalCount := mergeResults(req.Results)
    
    // Calculate totals
    totalWords := 0
    for _, count := range finalCount {
        totalWords += count
    }
    
    // Get top 10 words
    topWords := getTopWords(finalCount, 10)
    
    // Send response
    response := ReduceResponse{
        FinalCount:  finalCount,
        TotalWords:  totalWords,
        UniqueWords: len(finalCount),
        TopWords:    topWords,
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func main() {
    fmt.Println("Reducer service starting on port 8080...")
    
    http.HandleFunc("/reduce", handleReduce)
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("OK"))
    })
    
    if err := http.ListenAndServe(":8080", nil); err != nil {
        fmt.Printf("Server error: %v\n", err)
    }
}