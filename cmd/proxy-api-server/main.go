package main

import (
	"fmt"
	"os"
	"proxy-service/internal/client"
	"proxy-service/internal/config"
	"proxy-service/internal/database"
	"proxy-service/internal/handlers"
	"proxy-service/internal/middlewares"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if gin.Mode() == gin.DebugMode {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	// Configs
	cfg, err := config.LoadProxyServer()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	// Postgres
	db, err := database.InitDB(&cfg.PGConfig)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	// Router
	router := gin.New()
	router.SetTrustedProxies(nil)

	// Middlewares
	router.Use(gin.Recovery())
	router.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "DELETE"},
	}))
	router.Use(middlewares.RequestIDMiddleware())
	router.Use(middlewares.ZerologMiddleware())

	// Handlers
	proxyHandlers := handlers.ProxyHandlers{
		DBRepository: &database.DBRepository{DB: db},
		HTTPClient:   &client.Client{},
	}

	router.GET("/ping", handlers.Ping)

	router.GET("/api/proxy/v1/:service/*path", proxyHandlers.ProxyGetRequest)
	router.POST("/api/proxy/v1/:service/*path", proxyHandlers.ProxyPostRequest)
	router.PUT("/api/proxy/v1/:service/*path", proxyHandlers.ProxyPutRequest)
	router.DELETE("/api/proxy/v1/:service/*path", proxyHandlers.ProxyDeleteRequest)

	// Server
	router.Run(fmt.Sprintf(":%d", cfg.Port))
}
