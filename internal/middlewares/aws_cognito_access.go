package middlewares

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type UserClaims struct {
	CognitoGroups []string `json:"cognito:groups"`
}

func AWSCognitoMiddleware(isDebugging bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if isDebugging {
			c.Set("user_roles", []string{"proxy-service-admin-access", "proxy-service-proxy-access"})
			c.Next()
			return
		}

		if strings.HasSuffix(c.Request.URL.Path, "/ping") {
			c.Next()
			return
		}

		oidcData := c.GetHeader("X-Amzn-Oidc-Accesstoken")
		if oidcData == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "No authentication data provided"})
			c.Abort()
			return
		}

		claims, err := decodeUserClaims(oidcData)
		if err != nil {
			log.Error().Msgf("Got error while decoding AWS user claims: %s", err.Error())
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid authentication data"})
			c.Abort()
			return
		}

		c.Set("user_roles", claims.CognitoGroups)
		c.Next()
	}
}

func decodeUserClaims(oidcData string) (*UserClaims, error) {
	parts := strings.Split(oidcData, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid JWT format")
	}

	payload := parts[1]

	if l := len(payload) % 4; l > 0 {
		payload += strings.Repeat("=", 4-l)
	}

	decoded, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		return nil, err
	}

	var claims UserClaims
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return nil, err
	}

	return &claims, nil
}
