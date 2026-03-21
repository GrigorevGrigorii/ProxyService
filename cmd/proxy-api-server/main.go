package main

import (
	"proxy-service/internal/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.SetTrustedProxies(nil)

	router.GET("/ping", handlers.Ping)

	router.GET("/api/v1/:service/*path", handlers.ProxyGetRequest)
	router.POST("/api/v1/:service/*path", handlers.ProxyPostRequest)
	router.PUT("/api/v1/:service/*path", handlers.ProxyPutRequest)
	router.DELETE("/api/v1/:service/*path", handlers.ProxyDeleteRequest)

	router.Run("localhost:8080")
}
