package main

import (
	"proxy-service/internal/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.SetTrustedProxies(nil)

	router.GET("/ping", handlers.PingHandler)

	router.GET("/api/v1/:service/*path", handlers.GetProxyHandler)
	router.POST("/api/v1/:service/*path", handlers.PostProxyHandler)
	router.PUT("/api/v1/:service/*path", handlers.PutProxyHandler)
	router.DELETE("/api/v1/:service/*path", handlers.DeleteProxyHandler)

	router.Run("localhost:8080")
}
