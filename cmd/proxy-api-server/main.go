package main

import (
	"crypto/tls"
	"fmt"
	"os"
	"proxy-service/api/proxy-api-server/docs"
	"proxy-service/internal/auth"
	"proxy-service/internal/cache"
	"proxy-service/internal/config"
	"proxy-service/internal/database"
	"proxy-service/internal/handlers"
	"proxy-service/internal/httpclient"
	"proxy-service/internal/middlewares"

	"github.com/casbin/casbin/v2"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title			Proxy Service Proxy API
// @version		1.0
// @description	API for proxying requests
// @BasePath		/api/proxy
func main() {
	// Logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if gin.IsDebugging() {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
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
	var redisTLSConfig *tls.Config
	if cfg.RedisConfig.EnableTLS {
		redisTLSConfig = &tls.Config{}
	}
	rdb := redis.NewUniversalClient(&redis.UniversalOptions{
		MasterName: cfg.RedisConfig.GetMasterName(),
		Addrs:      cfg.RedisConfig.Addrs,
		Password:   cfg.RedisConfig.GetPassword(),
		DB:         cfg.RedisConfig.Database,
		PoolSize:   cfg.RedisConfig.PoolSize,
		TLSConfig:  redisTLSConfig,
		ReadOnly:   cfg.RedisConfig.ReadOnly,
	})

	// Casbin
	casbinEnforcer, err := casbin.NewEnforcer("configs/auth/authz_model.conf", "configs/auth/authz_policy.csv")
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
	router.Use(middlewares.AuthMiddleware(auth.AWSCognitoAuthChecker{IsDebugging: gin.IsDebugging()}))
	router.Use(middlewares.AccessMiddleware(casbinEnforcer))

	// Swagger
	docs.SwaggerInfo.Host = cfg.SwaggerHost
	router.GET("/api/proxy/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Handlers
	proxyHandlers := handlers.ProxyHandlers{
		DBRepository:    &database.DBRepository{DB: db},
		HTTPClient:      &httpclient.Client{},
		CacheRepository: &cache.CacheRepository{Redis: &rdb},
	}

	router.GET("/api/proxy/ping", handlers.Ping)

	router.GET("/api/proxy/v1/:service/*path", proxyHandlers.ProxyGetRequest)
	router.POST("/api/proxy/v1/:service/*path", proxyHandlers.ProxyPostRequest)
	router.PUT("/api/proxy/v1/:service/*path", proxyHandlers.ProxyPutRequest)
	router.DELETE("/api/proxy/v1/:service/*path", proxyHandlers.ProxyDeleteRequest)

	// Server
	router.Run(fmt.Sprintf(":%d", cfg.Port))
}
