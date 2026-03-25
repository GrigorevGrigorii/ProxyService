package main

import (
	"fmt"
	"os"
	"proxy-service/internal/config"
	"proxy-service/internal/database"
	"proxy-service/internal/handlers"
	"proxy-service/internal/httpclient"
	"proxy-service/internal/middlewares"
	"proxy-service/internal/redis_repositories"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
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
	cfg, err := config.LoadConfig[config.ProxyServerConfig]()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	// Postgres
	db, err := database.InitDB(&cfg.PGConfig)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	// Redis
	rdb := redis.NewFailoverClusterClient(&redis.FailoverOptions{
		MasterName:    cfg.RedisConfig.MasterName,
		SentinelAddrs: cfg.RedisConfig.Hosts,
		Password:      cfg.RedisConfig.Password,
		DB:            cfg.RedisConfig.Database,
		ReplicaOnly:   true,
	})

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
		DBRepository:    &database.DBRepository{DB: db},
		HTTPClient:      &httpclient.Client{},
		RedisRepository: &redis_repositories.RedisRepository{Redis: rdb},
	}

	router.GET("/ping", handlers.Ping)

	router.GET("/api/proxy/v1/:service/*path", proxyHandlers.ProxyGetRequest)
	router.POST("/api/proxy/v1/:service/*path", proxyHandlers.ProxyPostRequest)
	router.PUT("/api/proxy/v1/:service/*path", proxyHandlers.ProxyPutRequest)
	router.DELETE("/api/proxy/v1/:service/*path", proxyHandlers.ProxyDeleteRequest)

	// Server
	router.Run(fmt.Sprintf(":%d", cfg.Port))
}
