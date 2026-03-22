package main

import (
	"fmt"
	"log"
	"proxy-service/internal/client"
	"proxy-service/internal/config"
	"proxy-service/internal/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	services, err := config.LoadServices(cfg.ProxyServer.ServicesPath)
	if err != nil {
		log.Fatal(err)
	}

	proxyHandlers := handlers.ProxyHandlers{
		Services:   services,
		HTTPClient: &client.Client{},
	}

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "DELETE"},
	}))
	router.SetTrustedProxies(nil)

	router.GET("/ping", handlers.Ping)

	router.GET("/api/v1/:service/*path", proxyHandlers.ProxyGetRequest)
	router.POST("/api/v1/:service/*path", proxyHandlers.ProxyPostRequest)
	router.PUT("/api/v1/:service/*path", proxyHandlers.ProxyPutRequest)
	router.DELETE("/api/v1/:service/*path", proxyHandlers.ProxyDeleteRequest)

	router.Run(fmt.Sprintf(":%d", cfg.ProxyServer.Port))
}
