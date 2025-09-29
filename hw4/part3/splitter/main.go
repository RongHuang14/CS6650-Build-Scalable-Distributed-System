// splitter/main.go
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

// S3SplitRequest for S3-based splitting
type S3SplitRequest struct {
    S3URL   string `json:"s3_url"`    // s3://bucket/key
    Chunks  int    `json:"chunks"`
}

// S3SplitResponse returns S3 URLs of chunks
type S3SplitResponse struct {
    ChunkURLs []string `json:"chunk_urls"`
    Count     int      `json:"total_chunks"`
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

// handleS3Split processes S3-based split request
func handleS3Split(w http.ResponseWriter, r *http.Request) {
    var req S3SplitRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    if req.Chunks == 0 {
        req.Chunks = 3
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
    
    // Download file from S3
    result, err := svc.GetObject(&s3.GetObjectInput{
        Bucket: aws.String(bucket),
        Key:    aws.String(key),
    })
    if err != nil {
        http.Error(w, fmt.Sprintf("Failed to download from S3: %v", err), http.StatusInternalServerError)
        return
    }
    defer result.Body.Close()
    
    // Read content
    content, err := io.ReadAll(result.Body)
    if err != nil {
        http.Error(w, "Failed to read file", http.StatusInternalServerError)
        return
    }
    
    // Split text using existing function
    chunks := splitText(string(content), req.Chunks)
    
    // Upload chunks to S3
    var chunkURLs []string
    timestamp := time.Now().Unix()
    
    for i, chunk := range chunks {
        // Upload chunk to S3
        chunkKey := fmt.Sprintf("chunks/%d/chunk_%d.txt", timestamp, i)
        _, err := svc.PutObject(&s3.PutObjectInput{
            Bucket: aws.String(bucket),
            Key:    aws.String(chunkKey),
            Body:   bytes.NewReader([]byte(chunk)),
        })
        if err != nil {
            http.Error(w, fmt.Sprintf("Failed to upload chunk: %v", err), http.StatusInternalServerError)
            return
        }
        
        chunkURLs = append(chunkURLs, fmt.Sprintf("s3://%s/%s", bucket, chunkKey))
    }
    
    response := S3SplitResponse{
        ChunkURLs: chunkURLs,
        Count:     len(chunkURLs),
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func main() {
    fmt.Println("Splitter service starting on port 8080...")
    
    http.HandleFunc("/split", handleSplit)
    http.HandleFunc("/split-s3", handleS3Split)  // New S3 endpoint
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("OK"))
    })
    
    if err := http.ListenAndServe(":8080", nil); err != nil {
        fmt.Printf("Server error: %v\n", err)
    }
}