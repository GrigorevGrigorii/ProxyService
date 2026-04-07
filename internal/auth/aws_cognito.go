package auth

import (
	"errors"
	"net/http"
	"proxy-service/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

type userClaims struct {
	CognitoGroups []string `json:"cognito:groups"`
}

// it has already been validated by the ALB itself, so skipping signature validation is acceptable in this context
func (c userClaims) Valid() error { return nil }

type AWSCognitoAuthChecker struct{}

func (ac AWSCognitoAuthChecker) GetDebaggingRoles() []string {
	return []string{"proxy-service-admin-access", "proxy-service-proxy-access"}
}

func (ac AWSCognitoAuthChecker) Check(c *gin.Context) (roles []string, errStatusCode int, err error) {
	log := utils.GetLogger(c.Request.Context())

	oidcData := c.GetHeader("X-Amzn-Oidc-Accesstoken")
	if oidcData == "" {
		return nil, http.StatusUnauthorized, errors.New("No authentication data provided")
	}

	claims, err := ac.decodeUserClaims(oidcData)
	if err != nil {
		log.Error().Msgf("Got error while decoding AWS user claims: %s", err.Error())
		return nil, http.StatusUnauthorized, errors.New("Invalid authentication data")
	}

	return claims.CognitoGroups, 0, nil
}

func (ac AWSCognitoAuthChecker) decodeUserClaims(oidcData string) (*userClaims, error) {
	parser := jwt.Parser{SkipClaimsValidation: true}
	token, _, err := parser.ParseUnverified(oidcData, &userClaims{})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*userClaims); ok {
		return claims, nil
	}

	return nil, jwt.ErrInvalidKey
}
