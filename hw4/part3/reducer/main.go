// reducer/main.go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "sort"
    "strings"
    "time"
    
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
)

// S3ReduceRequest contains S3 URLs of mapper results to merge
type S3ReduceRequest struct {
    ResultURLs []string `json:"result_urls"`
}

// S3ReduceResponse contains S3 URL of final aggregated result
type S3ReduceResponse struct {
    FinalResultURL string `json:"final_result_url"`
}

// WordCountResult represents the word count output from mapper
type WordCountResult struct {
    WordCount   map[string]int `json:"word_count"`
    TotalWords  int           `json:"total_words"`
    UniqueWords int           `json:"unique_words"`
}

// WordCount represents a word and its count for sorting
type WordCount struct {
    Word  string `json:"word"`
    Count int    `json:"count"`
}

// FinalResult contains the final aggregated results
type FinalResult struct {
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

// handleS3Reduce processes the reduce request for S3-stored mapper results
func handleS3Reduce(w http.ResponseWriter, r *http.Request) {
    startTime := time.Now()
    
    // Parse request
    var req S3ReduceRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    // Create S3 session
    sess := session.Must(session.NewSession(&aws.Config{
        Region: aws.String("us-west-2"),
    }))
    svc := s3.New(sess)
    
    // Initialize final count map
    finalCount := make(map[string]int)
    
    // Download and merge all mapper results
    mergeStart := time.Now()
    fmt.Printf("Processing %d mapper results\n", len(req.ResultURLs))
    for i, url := range req.ResultURLs {
        // Parse S3 URL
        parts := strings.Split(strings.TrimPrefix(url, "s3://"), "/")
        if len(parts) < 2 {
            fmt.Printf("Skipping invalid URL: %s\n", url)
            continue
        }
        bucket := parts[0]
        key := strings.Join(parts[1:], "/")
        
        // Download mapper result from S3
        fmt.Printf("Downloading result %d from s3://%s/%s\n", i+1, bucket, key)
        result, err := svc.GetObject(&s3.GetObjectInput{
            Bucket: aws.String(bucket),
            Key:    aws.String(key),
        })
        if err != nil {
            http.Error(w, fmt.Sprintf("Failed to download result %d: %v", i+1, err), http.StatusInternalServerError)
            return
        }
        
        // Read result content
        content, err := io.ReadAll(result.Body)
        result.Body.Close()
        if err != nil {
            http.Error(w, fmt.Sprintf("Failed to read result %d", i+1), http.StatusInternalServerError)
            return
        }
        
        // Parse mapper result
        var mapResult WordCountResult
        if err := json.Unmarshal(content, &mapResult); err != nil {
            http.Error(w, fmt.Sprintf("Failed to parse result %d", i+1), http.StatusInternalServerError)
            return
        }
        
        // Merge word counts into final count
        for word, count := range mapResult.WordCount {
            finalCount[word] += count
        }
        fmt.Printf("Merged result %d: %d unique words\n", i+1, len(mapResult.WordCount))
    }
    fmt.Printf("[TIMING] Merge all results: %.2f seconds\n", time.Since(mergeStart).Seconds())
    
    // Calculate totals
    totalWords := 0
    for _, count := range finalCount {
        totalWords += count
    }
    
    // Get top 10 words
    topWords := getTopWords(finalCount, 10)
    
    // Create final result
    finalResult := FinalResult{
        FinalCount:  finalCount,
        TotalWords:  totalWords,
        UniqueWords: len(finalCount),
        TopWords:    topWords,
    }
    
    // Convert final result to JSON
    resultJSON, err := json.Marshal(finalResult)
    if err != nil {
        http.Error(w, "Failed to marshal final result", http.StatusInternalServerError)
        return
    }
    
    // Generate unique key for final result
    timestamp := time.Now().Unix()
    // Use bucket from first result URL
    bucket := strings.Split(strings.TrimPrefix(req.ResultURLs[0], "s3://"), "/")[0]
    resultKey := fmt.Sprintf("final-results/%d/final_word_count.json", timestamp)
    
    // Upload final result to S3
    fmt.Printf("Uploading final result to s3://%s/%s\n", bucket, resultKey)
    _, err = svc.PutObject(&s3.PutObjectInput{
        Bucket:      aws.String(bucket),
        Key:         aws.String(resultKey),
        Body:        bytes.NewReader(resultJSON),
        ContentType: aws.String("application/json"),
    })
    if err != nil {
        http.Error(w, fmt.Sprintf("Failed to upload final result: %v", err), http.StatusInternalServerError)
        return
    }
    
    fmt.Printf("[TIMING] TOTAL Reduce Time: %.2f seconds\n", time.Since(startTime).Seconds())
    
    // Send response with final result URL
    response := S3ReduceResponse{
        FinalResultURL: fmt.Sprintf("s3://%s/%s", bucket, resultKey),
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
    
    // Log summary
    fmt.Printf("Reduce completed successfully!\n")
    fmt.Printf("Total words: %d\n", totalWords)
    fmt.Printf("Unique words: %d\n", len(finalCount))
    fmt.Printf("Final result at: %s\n", response.FinalResultURL)
}

// handleSingleMachine processes entire file without MapReduce for comparison
func handleSingleMachine(w http.ResponseWriter, r *http.Request) {
    startTime := time.Now()
    
    var req struct {
        S3URL string `json:"s3_url"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    // Parse S3 URL
    parts := strings.Split(strings.TrimPrefix(req.S3URL, "s3://"), "/")
    if len(parts) < 2 {
        http.Error(w, "Invalid S3 URL", http.StatusBadRequest)
        return
    }
    bucket := parts[0]
    key := strings.Join(parts[1:], "/")
    
    // Create S3 session
    sess := session.Must(session.NewSession(&aws.Config{
        Region: aws.String("us-west-2"),
    }))
    svc := s3.New(sess)
    
    // Download entire file
    downloadStart := time.Now()
    result, err := svc.GetObject(&s3.GetObjectInput{
        Bucket: aws.String(bucket),
        Key:    aws.String(key),
    })
    if err != nil {
        http.Error(w, fmt.Sprintf("Failed to download: %v", err), http.StatusInternalServerError)
        return
    }
    content, err := io.ReadAll(result.Body)
    result.Body.Close()
    if err != nil {
        http.Error(w, "Failed to read file content", http.StatusInternalServerError)
        return
    }
    fmt.Printf("[SINGLE] Download Time: %.2f seconds\n", time.Since(downloadStart).Seconds())
    fmt.Printf("[SINGLE] File size: %d bytes (%.2f MB)\n", len(content), float64(len(content))/(1024*1024))
    
    // Process entire file
    processStart := time.Now()
    text := strings.ToLower(string(content))
    words := strings.Fields(text)
    wordCount := make(map[string]int)
    for _, word := range words {
        word = strings.Trim(word, ".,!?;:\\\"'()[]{}—–-")
        if word != "" {
            wordCount[word]++
        }
    }
    fmt.Printf("[SINGLE] Processing Time: %.2f seconds\n", time.Since(processStart).Seconds())
    
    // Calculate totals
    totalWords := 0
    for _, count := range wordCount {
        totalWords += count
    }
    
    // Get top 10 words
    topWords := getTopWords(wordCount, 10)
    
    // Upload result
    uploadStart := time.Now()
    finalResult := FinalResult{
        FinalCount:  wordCount,
        TotalWords:  totalWords,
        UniqueWords: len(wordCount),
        TopWords:    topWords,
    }
    resultJSON, _ := json.Marshal(finalResult)
    
    timestamp := time.Now().Unix()
    resultKey := fmt.Sprintf("single-machine/%d/result.json", timestamp)
    _, err = svc.PutObject(&s3.PutObjectInput{
        Bucket:      aws.String(bucket),
        Key:         aws.String(resultKey),
        Body:        bytes.NewReader(resultJSON),
        ContentType: aws.String("application/json"),
    })
    if err != nil {
        http.Error(w, fmt.Sprintf("Failed to upload result: %v", err), http.StatusInternalServerError)
        return
    }
    fmt.Printf("[SINGLE] Upload Time: %.2f seconds\n", time.Since(uploadStart).Seconds())
    
    totalTime := time.Since(startTime).Seconds()
    fmt.Printf("[SINGLE] *** TOTAL TIME: %.2f seconds ***\n", totalTime)
    fmt.Printf("[SINGLE] Total words: %d, Unique words: %d\n", totalWords, len(wordCount))
    
    // Send response
    response := map[string]interface{}{
        "total_time": totalTime,
        "total_words": totalWords,
        "unique_words": len(wordCount),
        "result_url": fmt.Sprintf("s3://%s/%s", bucket, resultKey),
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func main() {
    fmt.Println("Reducer service starting on port 8080...")
    
    http.HandleFunc("/reduce-s3", handleS3Reduce)
    http.HandleFunc("/single-machine", handleSingleMachine)
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("OK"))
    })
    
    if err := http.ListenAndServe(":8080", nil); err != nil {
        fmt.Printf("Server error: %v\n", err)
    }
}