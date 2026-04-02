package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type StatusResponse struct {
	Status string `json:"status"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

// Ping godoc
//
//	@Summary	Ping
//	@Tags		Common API
//	@Produce	json
//	@Success	200	{object}	StatusResponse	"Success"
//	@Router		/ping [get]
func Ping(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, StatusResponse{Status: "ok"})
}
