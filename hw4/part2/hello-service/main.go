package main

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

// album represents data about a record album
type album struct {
    ID     string  `json:"id"`
    Title  string  `json:"title"`
    Artist string  `json:"artist"`
    Price  float64 `json:"price"`
}

// albums slice to seed record album data
var albums = []album{
    {ID: "1", Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
    {ID: "2", Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
    {ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
}

// getAlbums responds with the list of all albums as JSON
func getAlbums(c *gin.Context) {
    c.IndentedJSON(http.StatusOK, albums)
}

// postAlbums adds an album from JSON received in the request body
func postAlbums(c *gin.Context) {
    var newAlbum album
    
    // Bind the received JSON to newAlbum
    if err := c.BindJSON(&newAlbum); err != nil {
        return
    }
    
    // Add the new album to the slice
    albums = append(albums, newAlbum)
    c.IndentedJSON(http.StatusCreated, newAlbum)
}

// getAlbumByID locates the album whose ID value matches the id parameter
func getAlbumByID(c *gin.Context) {
    id := c.Param("id")
    
    // Loop through albums to find matching ID
    for _, a := range albums {
        if a.ID == id {
            c.IndentedJSON(http.StatusOK, a)
            return
        }
    }
    c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
}

// updateAlbumByID updates an album by ID
func updateAlbumByID(c *gin.Context) {
    id := c.Param("id")
    var updatedAlbum album
    
    if err := c.BindJSON(&updatedAlbum); err != nil {
        c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid JSON"})
        return
    }
    
    // Find and update the album
    for i, a := range albums {
        if a.ID == id {
            albums[i] = updatedAlbum
            c.IndentedJSON(http.StatusOK, updatedAlbum)
            return
        }
    }
    c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
}

// deleteAlbumByID deletes an album by ID
func deleteAlbumByID(c *gin.Context) {
    id := c.Param("id")
    
    // Find and remove the album
    for i, a := range albums {
        if a.ID == id {
            albums = append(albums[:i], albums[i+1:]...)
            c.IndentedJSON(http.StatusOK, gin.H{"message": "album deleted"})
            return
        }
    }
    c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
}

func main() {
    router := gin.Default()
    router.GET("/albums", getAlbums)
    router.POST("/albums", postAlbums)
    router.GET("/albums/:id", getAlbumByID)
    router.PUT("/albums/:id", updateAlbumByID)
    router.DELETE("/albums/:id", deleteAlbumByID)
    
    router.Run(":8080")
}