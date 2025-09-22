// server.go - Simple HTTP server for load testing
package main

import (
    "encoding/json"
    "log"
    "net/http"
    "sync"
)

// Thread-safe storage with RWMutex
var storage = struct {
    sync.RWMutex
    m map[string]string
}{m: make(map[string]string)}

func handleGet(w http.ResponseWriter, r *http.Request) {
    key := r.URL.Query().Get("key")
    
    // RLock allows multiple concurrent reads
    storage.RLock()
    value := storage.m[key]
    storage.RUnlock()
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "value": value,
    })
}

func handlePost(w http.ResponseWriter, r *http.Request) {
    var data map[string]string
    if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
        http.Error(w, "invalid JSON", http.StatusBadRequest)
        return
    }
    
    // Lock for exclusive write access
    storage.Lock()
    storage.m[data["key"]] = data["value"]
    storage.Unlock()
    
    w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(data)
}

func main() {
    http.HandleFunc("/get", handleGet)
    http.HandleFunc("/post", handlePost)
    log.Println("HTTP server listening on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}