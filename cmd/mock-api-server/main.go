package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"proxy-service/internal/config"

	"github.com/gin-gonic/gin"
)

type MockHandlers struct {
	cfg *config.MockServerConfig
}

func (h *MockHandlers) mockHandler(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		body = []byte(err.Error())
	}

	resp := gin.H{
		"method": c.Request.Method,
		"query":  c.Request.URL.RawQuery,
		"body":   string(body),
	}
	c.IndentedJSON(h.cfg.ResponseStatusCode, resp)
}

func main() {
	cfg, err := config.LoadMockServer()
	if err != nil {
		log.Fatal(err)
	}

	mockHandlers := MockHandlers{
		cfg: cfg,
	}

	router := gin.Default()
	router.SetTrustedProxies(nil)

	router.GET("/ping", func(c *gin.Context) { c.IndentedJSON(http.StatusOK, gin.H{"status": "ok"}) })

	router.GET("/mock", mockHandlers.mockHandler)
	router.POST("/mock", mockHandlers.mockHandler)
	router.PUT("/mock", mockHandlers.mockHandler)
	router.DELETE("/mock", mockHandlers.mockHandler)

	router.Run(fmt.Sprintf(":%d", cfg.Port))
}
