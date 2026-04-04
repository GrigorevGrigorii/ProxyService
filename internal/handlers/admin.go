package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"proxy-service/internal/database"
	"proxy-service/internal/models"

	"github.com/gin-gonic/gin"
)

type AdminHandlers struct {
	DBRepository database.Repository
}

// GetServices godoc
//
//	@Summary	Get all services with targets
//	@Tags		Admin API
//	@Produce	json
//	@Success	200	{object}	[]models.ServiceDTO	"Success"
//	@Router		/v1/service [get]
func (h *AdminHandlers) GetServices(c *gin.Context) {
	services, err := h.DBRepository.GetAll(c.Request.Context())
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, MessageResponse{Message: err.Error()})
		return
	}

	response := make([]models.ServiceDTO, len(services))
	for i, service := range services {
		response[i] = models.ServiceDTOFromDBModel(service)
	}

	c.IndentedJSON(http.StatusOK, response)
}

// GetService godoc
//
//	@Summary	Get service with targets by name
//	@Tags		Admin API
//	@Produce	json
//	@Param		name	path		string				true	"Service name"
//	@Success	200		{object}	models.ServiceDTO	"Success"
//	@Failure	404		{object}	MessageResponse		"Service not found"
//	@Router		/v1/service/{name} [get]
func (h *AdminHandlers) GetService(c *gin.Context) {
	service, err := h.DBRepository.Get(c.Request.Context(), c.Param("name"))
	if errors.Is(err, database.ErrNotFound) {
		c.IndentedJSON(http.StatusNotFound, MessageResponse{Message: fmt.Sprintf("Service '%s' not found", c.Param("name"))})
		return
	}
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, MessageResponse{Message: err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, models.ServiceDTOFromDBModel(*service))
}

// CreateService godoc
//
//	@Summary	Create service
//	@Tags		Admin API
//	@Accept		json
//	@Produce	json
//	@Param		request	body		models.ServiceDTO	true	"Service data (with targets)"
//	@Success	200		{object}	MessageResponse		"Success"
//	@Failure	400		{object}	MessageResponse		"Bad request"
//	@Failure	409		{object}	MessageResponse		"Service already exists"
//	@Router		/v1/service [post]
func (h *AdminHandlers) CreateService(c *gin.Context) {
	var request models.ServiceDTO

	if err := c.BindJSON(&request); err != nil {
		c.IndentedJSON(http.StatusBadRequest, MessageResponse{Message: err.Error()})
		return
	}
	if err := models.Validate.StructCtx(c.Request.Context(), request); err != nil {
		c.IndentedJSON(http.StatusBadRequest, MessageResponse{Message: fmt.Sprintf("Cannot parse query: %s", err.Error())})
		return
	}
	for i := range request.Targets {
		request.Targets[i].SortQuery()
	}

	request.Version = 0
	service := models.ServiceDBModelFromDTO(request)
	err := h.DBRepository.Create(c.Request.Context(), &service)
	if errors.Is(err, database.ErrAlreadyExists) {
		c.IndentedJSON(http.StatusConflict, MessageResponse{Message: fmt.Sprintf("Service '%s' already exists", service.Name)})
		return
	}
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, MessageResponse{Message: err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, MessageResponse{Message: "ok"})
}

// UpdateService godoc
//
//	@Summary	Update service
//	@Tags		Admin API
//	@Accept		json
//	@Produce	json
//	@Param		name	path		string				true	"Service name"
//	@Param		request	body		models.ServiceDTO	true	"Service data (with targets)"
//	@Success	200		{object}	MessageResponse		"Success"
//	@Failure	400		{object}	MessageResponse		"Bad request"
//	@Failure	404		{object}	MessageResponse		"Service not found"
//	@Router		/v1/service/{name} [put]
func (h *AdminHandlers) UpdateService(c *gin.Context) {
	var request models.ServiceDTO

	if err := c.BindJSON(&request); err != nil {
		c.IndentedJSON(http.StatusBadRequest, MessageResponse{Message: err.Error()})
		return
	}
	if err := models.Validate.StructCtx(c.Request.Context(), request); err != nil {
		c.IndentedJSON(http.StatusBadRequest, MessageResponse{Message: fmt.Sprintf("Cannot parse query: %s", err.Error())})
		return
	}
	for i := range request.Targets {
		request.Targets[i].SortQuery()
	}

	if request.Name != c.Param("name") {
		c.IndentedJSON(http.StatusBadRequest, MessageResponse{Message: "Connot update name"})
		return
	}

	service := models.ServiceDBModelFromDTO(request)
	err := h.DBRepository.Update(c.Request.Context(), &service)
	if errors.Is(err, database.ErrNotFound) {
		c.IndentedJSON(http.StatusNotFound, MessageResponse{Message: fmt.Sprintf("Service '%s' not found", c.Param("name"))})
		return
	}
	if errors.Is(err, database.ErrVersionMismatch) {
		c.IndentedJSON(http.StatusConflict, MessageResponse{Message: err.Error()})
		return
	}
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, MessageResponse{Message: err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, MessageResponse{Message: "ok"})
}

// DeleteService godoc
//
//	@Summary	Delete service
//	@Tags		Admin API
//	@Produce	json
//	@Param		name	path		string			true	"Service name"
//	@Success	200		{object}	MessageResponse	"Success"
//	@Router		/v1/service/{name} [delete]
func (h *AdminHandlers) DeleteService(c *gin.Context) {
	err := h.DBRepository.Delete(c.Request.Context(), c.Param("name"))
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, MessageResponse{Message: err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, MessageResponse{Message: "ok"})
}
