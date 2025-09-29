// mapper/main.go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "strings"
    "time"
    
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
)

// S3MapRequest contains S3 URL of text chunk to process
type S3MapRequest struct {
    ChunkURL string `json:"chunk_url"`
}

// S3MapResponse contains S3 URL of the word count result
type S3MapResponse struct {
    ResultURL string `json:"result_url"`
}

// WordCountResult represents the word count output from mapper
type WordCountResult struct {
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
        // Clean punctuation (improved version)
        word = strings.Trim(word, ".,!?;:\\\"'()[]{}—–-")
        if word != "" {
            wordCount[word]++
        }
    }
    
    return wordCount
}

// handleS3Map processes the map request for S3-stored chunks
func handleS3Map(w http.ResponseWriter, r *http.Request) {
    startTime := time.Now()
    
    // Parse request
    var req S3MapRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    // Parse S3 URL (format: s3://bucket/key)
    parts := strings.Split(strings.TrimPrefix(req.ChunkURL, "s3://"), "/")
    if len(parts) < 2 {
        http.Error(w, "Invalid S3 URL format", http.StatusBadRequest)
        return
    }
    bucket := parts[0]
    key := strings.Join(parts[1:], "/")
    
    // Create S3 session
    sess := session.Must(session.NewSession(&aws.Config{
        Region: aws.String("us-west-2"),
    }))
    svc := s3.New(sess)
    
    // Download chunk from S3
    downloadStart := time.Now()
    fmt.Printf("Downloading chunk from s3://%s/%s\n", bucket, key)
    result, err := svc.GetObject(&s3.GetObjectInput{
        Bucket: aws.String(bucket),
        Key:    aws.String(key),
    })
    if err != nil {
        http.Error(w, fmt.Sprintf("Failed to download from S3: %v", err), http.StatusInternalServerError)
        return
    }
    defer result.Body.Close()
    
    // Read chunk content
    content, err := io.ReadAll(result.Body)
    if err != nil {
        http.Error(w, "Failed to read chunk content", http.StatusInternalServerError)
        return
    }
    fmt.Printf("[TIMING] Download chunk: %.2f seconds\n", time.Since(downloadStart).Seconds())
    fmt.Printf("[METRICS] Chunk size: %d bytes (%.2f MB)\n", len(content), float64(len(content))/(1024*1024)) 
    
    // Count words in the text chunk
    processStart := time.Now()
    wordCount := countWords(string(content))
    fmt.Printf("[TIMING] Process words: %.2f seconds\n", time.Since(processStart).Seconds())
    
    // Calculate totals
    totalWords := 0
    for _, count := range wordCount {
        totalWords += count
    }
    
    // Create result object
    mapResult := WordCountResult{
        WordCount:   wordCount,
        TotalWords:  totalWords,
        UniqueWords: len(wordCount),
    }
    
    // Convert result to JSON
    resultJSON, err := json.Marshal(mapResult)
    if err != nil {
        http.Error(w, "Failed to marshal result", http.StatusInternalServerError)
        return
    }
    
    // Generate unique key for result
    timestamp := time.Now().Unix()
    chunkName := strings.TrimSuffix(strings.Split(key, "/")[len(strings.Split(key, "/"))-1], ".txt")
    resultKey := fmt.Sprintf("map-results/%d/%s_result.json", timestamp, chunkName)
    
    // Upload result to S3
    uploadStart := time.Now()
    fmt.Printf("Uploading result to s3://%s/%s\n", bucket, resultKey)
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
    fmt.Printf("[TIMING] Upload result: %.2f seconds\n", time.Since(uploadStart).Seconds())
    
    fmt.Printf("[TIMING] TOTAL Map Time: %.2f seconds\n", time.Since(startTime).Seconds())
    fmt.Printf("[METRICS] Words processed: %d unique, %d total\n", len(wordCount), totalWords)
    
    // Send response with result URL
    response := S3MapResponse{
        ResultURL: fmt.Sprintf("s3://%s/%s", bucket, resultKey),
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
    fmt.Printf("Mapper completed successfully, result at: %s\n", response.ResultURL)
}

func main() {
    fmt.Println("Mapper service starting on port 8080...")
    
    http.HandleFunc("/map-s3", handleS3Map)
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("OK"))
    })
    
    if err := http.ListenAndServe(":8080", nil); err != nil {
        fmt.Printf("Server error: %v\n", err)
    }
}