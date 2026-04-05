package middlewares

import (
	"proxy-service/internal/utils"
	"time"

	"github.com/gin-gonic/gin"
)

func ZerologMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		log := utils.GetLogger(c.Request.Context())
		log.Info().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", c.Writer.Status()).
			Dur("latency", time.Since(start)).
			Str("client_ip", c.ClientIP()).
			Msg("request")
	}
}
