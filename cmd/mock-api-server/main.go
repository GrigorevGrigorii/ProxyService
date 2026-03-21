package main

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func mockHandler(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		body = []byte(err.Error())
	}

	resp := gin.H{
		"status": "ok",
		"method": c.Request.Method,
		"query":  c.Request.URL.Query(),
		"body":   string(body),
	}
	c.IndentedJSON(http.StatusOK, resp)
}

func main() {
	router := gin.Default()
	router.SetTrustedProxies(nil)

	router.GET("/ping", func(c *gin.Context) { c.IndentedJSON(http.StatusOK, gin.H{"status": "ok"}) })

	router.GET("/mock", mockHandler)
	router.POST("/mock", mockHandler)
	router.PUT("/mock", mockHandler)
	router.DELETE("/mock", mockHandler)

	router.Run(":8081")
}
