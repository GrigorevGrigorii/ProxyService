package middlewares

import (
	"proxy-service/internal/auth"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(IsDebugging bool, checker auth.AuthChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasSuffix(c.Request.URL.Path, "/ping") {
			c.Next()
			return
		}

		if IsDebugging {
			c.Set("user_roles", checker.GetDebaggingRoles())
			c.Next()
			return
		}

		roles, errStatusCode, err := checker.Check(c)
		if err != nil {
			c.JSON(errStatusCode, gin.H{"message": err.Error()})
			c.Abort()
			return
		}

		c.Set("user_roles", roles)
		c.Next()
	}
}
