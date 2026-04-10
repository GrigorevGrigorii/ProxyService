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
	ServiceRepository ServiceRepository
}

// GetServices godoc
//
//	@Summary	Get all services with targets
//	@Tags		Admin API
//	@Produce	json
//	@Success	200	{object}	[]models.ServiceDTO	"Success"
//	@Router		/v1/service [get]
func (h *AdminHandlers) GetServices(c *gin.Context) {
	services, err := h.ServiceRepository.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, services)
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
	service, err := h.ServiceRepository.Get(c.Request.Context(), c.Param("name"))
	if errors.Is(err, database.ErrNotFound) {
		c.JSON(http.StatusNotFound, MessageResponse{Message: fmt.Sprintf("Service '%s' not found", c.Param("name"))})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, *service)
}

// CreateService godoc
//
//	@Summary	Create service
//	@Tags		Admin API
//	@Accept		json
//	@Produce	json
//	@Param		request	body		createServiceRequest	true	"Service data (with targets)"
//	@Success	200		{object}	MessageResponse			"Success"
//	@Failure	400		{object}	MessageResponse			"Bad request"
//	@Failure	409		{object}	MessageResponse			"Service already exists"
//	@Router		/v1/service [post]
func (h *AdminHandlers) CreateService(c *gin.Context) {
	var request createServiceRequest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, MessageResponse{Message: err.Error()})
		return
	}
	if err := Validate.StructCtx(c.Request.Context(), request); err != nil {
		c.JSON(http.StatusBadRequest, MessageResponse{Message: fmt.Sprintf("Cannot parse query: %s", err.Error())})
		return
	}

	service := models.ServiceDTO{
		Name:          *request.Name,
		Scheme:        *request.Scheme,
		Host:          *request.Host,
		Timeout:       *request.Timeout,
		RetryCount:    *request.RetryCount,
		RetryInterval: *request.RetryInterval,
		Version:       0,
		Targets:       make([]models.TargetDTO, len(request.Targets)),
	}
	for i, t := range request.Targets {
		targetDTO := models.TargetDTO{Path: *t.Path, Method: *t.Method, Query: *t.Query, CacheInterval: t.CacheInterval}
		targetDTO.SortQuery()
		service.Targets[i] = targetDTO
	}

	err := h.ServiceRepository.Create(c.Request.Context(), &service)
	switch {
	case err == nil:
		c.JSON(http.StatusOK, MessageResponse{Message: "ok"})
	case errors.Is(err, database.ErrAlreadyExists):
		c.JSON(http.StatusConflict, MessageResponse{Message: fmt.Sprintf("Service '%s' already exists", service.Name)})
	default:
		c.JSON(http.StatusInternalServerError, MessageResponse{Message: err.Error()})
	}
}

// UpdateService godoc
//
//	@Summary	Update service
//	@Tags		Admin API
//	@Accept		json
//	@Produce	json
//	@Param		name	path		string					true	"Service name"
//	@Param		request	body		updateServiceRequest	true	"Service data"
//	@Success	200		{object}	MessageResponse			"Success"
//	@Failure	400		{object}	MessageResponse			"Bad request"
//	@Failure	404		{object}	MessageResponse			"Service not found"
//	@Router		/v1/service/{name} [put]
func (h *AdminHandlers) UpdateService(c *gin.Context) {
	var request updateServiceRequest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, MessageResponse{Message: err.Error()})
		return
	}
	if err := Validate.StructCtx(c.Request.Context(), request); err != nil {
		c.JSON(http.StatusBadRequest, MessageResponse{Message: fmt.Sprintf("Cannot parse query: %s", err.Error())})
		return
	}

	service := models.ServiceDTO{
		Name:          c.Param("name"),
		Scheme:        *request.Scheme,
		Host:          *request.Host,
		Timeout:       *request.Timeout,
		RetryCount:    *request.RetryCount,
		RetryInterval: *request.RetryInterval,
		Version:       *request.Version,
		Targets:       make([]models.TargetDTO, len(request.Targets)),
	}
	for i, t := range request.Targets {
		targetDTO := models.TargetDTO{Path: *t.Path, Method: *t.Method, Query: *t.Query, CacheInterval: t.CacheInterval}
		targetDTO.SortQuery()
		service.Targets[i] = targetDTO
	}

	err := h.ServiceRepository.Update(c.Request.Context(), &service)
	switch {
	case err == nil:
		c.JSON(http.StatusOK, MessageResponse{Message: "ok"})
	case errors.Is(err, database.ErrNotFound):
		c.JSON(http.StatusNotFound, MessageResponse{Message: fmt.Sprintf("Service '%s' not found", c.Param("name"))})
	case errors.Is(err, database.ErrVersionMismatch):
		c.JSON(http.StatusConflict, MessageResponse{Message: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, MessageResponse{Message: err.Error()})
	}
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
	err := h.ServiceRepository.Delete(c.Request.Context(), c.Param("name"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, MessageResponse{Message: "ok"})
}
