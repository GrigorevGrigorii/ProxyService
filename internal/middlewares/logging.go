package middlewares

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func ZerologMiddleware(logPings bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		if !logPings && strings.HasSuffix(c.Request.URL.Path, "/ping") {
			return
		}

		log.Info().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Str("request-id", c.GetHeader("X-Request-ID")).
			Int("status", c.Writer.Status()).
			Dur("latency", time.Since(start)).
			Str("client_ip", c.ClientIP()).
			Msg("request")
	}
}
