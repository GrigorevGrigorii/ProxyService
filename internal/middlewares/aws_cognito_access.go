package middlewares

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type UserClaims struct {
	CognitoGroups []string `json:"cognito:groups"`
	Email         string   `json:"email"`
}

func AWSCognitoAccessMiddleware(groupName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasSuffix(c.Request.URL.Path, "/ping") {
			c.Next()
			return
		}

		oidcData := c.GetHeader("X-Amzn-Oidc-Data")
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

		if !slices.Contains(claims.CognitoGroups, groupName) {
			c.JSON(http.StatusForbidden, gin.H{"message": "Forbidden"})
			c.Abort()
			return
		}

		log.Info().Msgf("User %s has enough permission to process request", claims.Email)
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
