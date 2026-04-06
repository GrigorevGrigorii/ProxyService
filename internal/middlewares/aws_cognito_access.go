package middlewares

import (
	"net/http"
	"proxy-service/internal/utils"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

type UserClaims struct {
	CognitoGroups []string `json:"cognito:groups"`
	jwt.StandardClaims
}

func AWSCognitoMiddleware(isDebugging bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := utils.GetLogger(c.Request.Context())

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
	// it has already been validated by the ALB itself, so skipping signature validation is acceptable in this context
	parser := jwt.Parser{SkipClaimsValidation: true}
	token, _, err := parser.ParseUnverified(oidcData, &UserClaims{})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*UserClaims); ok {
		return claims, nil
	}

	return nil, jwt.ErrInvalidKey
}
