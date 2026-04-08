package auth

import (
	"github.com/gin-gonic/gin"
)

type AuthChecker interface {
	Check(c *gin.Context) (roles []string, err error)
}
