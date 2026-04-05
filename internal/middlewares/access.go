package middlewares

import (
	"errors"
	"net/http"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
)

func AccessMiddleware(enforcer *casbin.Enforcer) gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasSuffix(c.Request.URL.Path, "/ping") {
			c.Next()
			return
		}

		userRoles, errStatusCode, err := getUserRoles(c)
		if err != nil {
			c.JSON(errStatusCode, gin.H{"message": err.Error()})
			c.Abort()
			return
		}

		allowed := false
		for _, role := range userRoles {
			ok, err := enforcer.Enforce(role, c.Request.URL.Path, c.Request.Method)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
				c.Abort()
				return
			}
			if ok {
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{"message": "Access denied"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func getUserRoles(c *gin.Context) (userRoles []string, errStatusCode int, err error) {
	userRolesInterface, exists := c.Get("user_roles")
	if !exists {
		return nil, http.StatusUnauthorized, errors.New("No user roles found in context")
	}
	userRoles, ok := userRolesInterface.([]string)
	if !ok {
		return nil, http.StatusInternalServerError, errors.New("Invalid user roles format")
	}
	return userRoles, 0, nil
}
