package auth

import (
	"errors"
	"proxy-service/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

var awsCognitoDebuggingRoles = []string{"proxy-service-admin-access", "proxy-service-proxy-access"}

type userClaims struct {
	CognitoGroups []string `json:"cognito:groups"`
}

func (c userClaims) Valid() error { return nil } // it has already been validated by the ALB itself, so skipping signature validation is acceptable in this context

type AWSCognitoAuthChecker struct {
	IsDebugging bool
}

func (ac AWSCognitoAuthChecker) Check(c *gin.Context) (roles []string, err error) {
	if ac.IsDebugging {
		return awsCognitoDebuggingRoles, nil
	}

	log := utils.GetLogger(c.Request.Context())

	oidcData := c.GetHeader("X-Amzn-Oidc-Accesstoken")
	if oidcData == "" {
		return nil, errors.New("No authentication data provided")
	}

	claims, err := ac.decodeUserClaims(oidcData)
	if err != nil {
		log.Error().Msgf("Got error while decoding AWS user claims: %s", err.Error())
		return nil, errors.New("Invalid authentication data")
	}

	return claims.CognitoGroups, nil
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
