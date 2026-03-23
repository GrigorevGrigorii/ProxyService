package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"proxy-service/internal/database"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AdminHandlers struct {
	DB *gorm.DB
}

func (h *AdminHandlers) GetServices(c *gin.Context) {
	c.IndentedJSON(http.StatusNotImplemented, gin.H{"message": "not_implemented_error"})
}

func (h *AdminHandlers) GetService(c *gin.Context) {
	var service database.Service
	result := h.DB.First(&service, "name = ?", c.Param("name"))

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("Service with name '%s' not found", c.Param("name"))})
		return
	}
	if result.Error != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": result.Error.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, service)
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
