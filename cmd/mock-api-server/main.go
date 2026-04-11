package main

import (
	"fmt"
	"io"
	"os"
	"proxy-service/internal/config"
	"proxy-service/internal/handlers"
	"proxy-service/internal/middlewares"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type MockHandlers struct {
	cfg config.MockServerConfig
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
	// Logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if gin.Mode() == gin.DebugMode {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// Configs
	cfg, err := config.LoadConfig[config.MockServerConfig]()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	// Router
	router := gin.New()
	router.SetTrustedProxies(nil)

	// Middlewares
	router.Use(middlewares.RequestIDMiddleware())
	router.Use(middlewares.ZerologMiddleware())

	// Handlers
	mockHandlers := MockHandlers{
		cfg: cfg,
	}

	router.GET("/mock/ping", handlers.Ping)

	router.GET("/mock", mockHandlers.mockHandler)
	router.POST("/mock", mockHandlers.mockHandler)
	router.PUT("/mock", mockHandlers.mockHandler)
	router.DELETE("/mock", mockHandlers.mockHandler)

	// Server
	router.Run(fmt.Sprintf(":%d", cfg.Port))
}
