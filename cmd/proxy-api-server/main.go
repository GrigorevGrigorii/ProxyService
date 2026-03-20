package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) { c.IndentedJSON(http.StatusOK, gin.H{"status": "ok"}) })

	router.Run("localhost:8080")
}
