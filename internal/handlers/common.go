package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type MessageResponse struct {
	Message string `json:"message"`
}

// Ping godoc
//
//	@Summary	Ping
//	@Tags		Common API
//	@Produce	json
//	@Success	200	{object}	MessageResponse	"Success"
//	@Router		/ping [get]
func Ping(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, MessageResponse{Message: "ok"})
}
