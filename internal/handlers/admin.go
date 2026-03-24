package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"proxy-service/internal/database"

	"github.com/gin-gonic/gin"
)

type AdminHandlers struct {
	DBRepository database.Repository
}

func (h *AdminHandlers) GetServices(c *gin.Context) {
	services, err := h.DBRepository.GetAll(c.Request.Context())
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	response := make([]ServiceDTO, len(services))
	for i, service := range services {
		response[i] = serviceDTOFromDBModel(service)
	}

	c.IndentedJSON(http.StatusOK, response)
}

func (h *AdminHandlers) GetService(c *gin.Context) {
	service, err := h.DBRepository.Get(c.Request.Context(), c.Param("name"))
	if errors.Is(err, database.ErrNotFound) {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("Service '%s' not found", c.Param("name"))})
		return
	}
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, serviceDTOFromDBModel(*service))
}

func (h *AdminHandlers) CreateService(c *gin.Context) {
	var request ServiceDTO

	if err := c.BindJSON(&request); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	service := serviceDBModelFromDTO(request)
	err := h.DBRepository.Create(c.Request.Context(), &service)
	if errors.Is(err, database.ErrAlreadyExists) {
		c.IndentedJSON(http.StatusConflict, gin.H{"message": fmt.Sprintf("Service '%s' already exists", service.Name)})
		return
	}
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "ok"})
}

func (h *AdminHandlers) UpdateService(c *gin.Context) {
	var request ServiceDTO

	if err := c.BindJSON(&request); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if request.Name != c.Param("name") {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Connot update name"})
		return
	}

	service := serviceDBModelFromDTO(request)
	err := h.DBRepository.Update(c.Request.Context(), &service)
	if errors.Is(err, database.ErrNotFound) {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("Service '%s' not found", c.Param("name"))})
		return
	}
	if errors.Is(err, database.ErrVersionMismatch) {
		c.IndentedJSON(http.StatusConflict, gin.H{"message": err.Error()})
		return
	}
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "ok"})
}

func (h *AdminHandlers) DeleteService(c *gin.Context) {
	err := h.DBRepository.Delete(c.Request.Context(), c.Param("name"))
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "ok"})
}
