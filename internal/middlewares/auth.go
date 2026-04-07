package middlewares

import (
	"strings"

	"github.com/gin-gonic/gin"
)

type AuthChecker interface {
	Check(c *gin.Context) (roles []string, errStatusCode int, err error)
	GetDebaggingRoles() []string
}

func AuthMiddleware(IsDebugging bool, checker AuthChecker) gin.HandlerFunc {
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
