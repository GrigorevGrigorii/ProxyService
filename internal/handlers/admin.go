package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AdminHandlers struct{}

func (h *AdminHandlers) GetServices(c *gin.Context) {
	c.IndentedJSON(http.StatusNotImplemented, gin.H{"message": "not_implemented_error"})
}

func (h *AdminHandlers) GetService(c *gin.Context) {
	c.IndentedJSON(http.StatusNotImplemented, gin.H{"message": "not_implemented_error"})
}

func (h *AdminHandlers) CreateService(c *gin.Context) {
	c.IndentedJSON(http.StatusNotImplemented, gin.H{"message": "not_implemented_error"})
}

func (h *AdminHandlers) UpdateService(c *gin.Context) {
	c.IndentedJSON(http.StatusNotImplemented, gin.H{"message": "not_implemented_error"})
}

func (h *AdminHandlers) DeleteService(c *gin.Context) {
	c.IndentedJSON(http.StatusNotImplemented, gin.H{"message": "not_implemented_error"})
}
