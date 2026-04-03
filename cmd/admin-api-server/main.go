package main

import (
	"fmt"
	"os"
	"proxy-service/api/admin-api-server/docs"
	"proxy-service/internal/config"
	"proxy-service/internal/database"
	"proxy-service/internal/handlers"
	"proxy-service/internal/middlewares"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title			Proxy Service Admin API
// @version		1.0
// @description	API for managing info about allowed services and targets to proxy
// @BasePath		/api/admin
func main() {
	// Logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if gin.Mode() == gin.DebugMode {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// Configs
	cfg, err := config.LoadConfig[config.AdminServerConfig]()
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
	router.Use(middlewares.ZerologMiddleware(cfg.LogPings))
	if cfg.AWSCognitoGroup != "" {
		router.Use(middlewares.AWSCognitoAccessMiddleware(cfg.AWSCognitoGroup))
	}

	// Swagger
	docs.SwaggerInfo.Host = cfg.SwaggerHost
	router.GET("/api/admin/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Handlers
	adminHandlers := handlers.AdminHandlers{
		DBRepository: &database.DBRepository{DB: db},
	}

	router.GET("/api/admin/ping", handlers.Ping)

	router.GET("/api/admin/v1/service", adminHandlers.GetServices)
	router.GET("/api/admin/v1/service/:name", adminHandlers.GetService)
	router.POST("/api/admin/v1/service", adminHandlers.CreateService)
	router.PUT("/api/admin/v1/service/:name", adminHandlers.UpdateService)
	router.DELETE("/api/admin/v1/service/:name", adminHandlers.DeleteService)

	// Server
	router.Run(fmt.Sprintf(":%d", cfg.Port))
}
