package auth

import (
	"github.com/gin-gonic/gin"
)

type AuthChecker interface {
	Check(c *gin.Context) (roles []string, errStatusCode int, err error) // check auth and return user roles
	GetDebaggingRoles() []string                                         // get roles that used in debug mode without real auth check
}
