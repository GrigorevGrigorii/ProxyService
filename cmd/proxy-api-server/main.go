package main

import (
	"fmt"
	"log"
	"proxy-service/internal/config"
	"proxy-service/internal/handlers"

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

	proxy_handlers := handlers.Handlers{
		Services: services,
	}

	router := gin.Default()
	router.SetTrustedProxies(nil)

	router.GET("/ping", handlers.Ping)

	router.GET("/api/v1/:service/*path", proxy_handlers.ProxyGetRequest)
	router.POST("/api/v1/:service/*path", proxy_handlers.ProxyPostRequest)
	router.PUT("/api/v1/:service/*path", proxy_handlers.ProxyPutRequest)
	router.DELETE("/api/v1/:service/*path", proxy_handlers.ProxyDeleteRequest)

	router.Run(fmt.Sprintf(":%d", cfg.ProxyServer.Port))
}
