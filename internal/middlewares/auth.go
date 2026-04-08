package middlewares

import (
	"net/http"
	"proxy-service/internal/auth"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(checker auth.AuthChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasSuffix(c.Request.URL.Path, "/ping") {
			c.Next()
			return
		}

		roles, err := checker.Check(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
			c.Abort()
			return
		}

		c.Set("user_roles", roles)
		c.Next()
	}
}
