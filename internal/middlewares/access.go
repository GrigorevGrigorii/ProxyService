package middlewares

import (
	"errors"
	"net/http"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
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

		logSuccess(c)
		c.Next()
	}
}

func getUserRoles(c *gin.Context) ([]string, int, error) {
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

func logSuccess(c *gin.Context) {
	emailInterface, exists := c.Get("user_email")
	if !exists {
		return
	}
	email, ok := emailInterface.(string)
	if !ok {
		return
	}

	log.Info().Msgf("User %s has enough permission to process request", email)
}
