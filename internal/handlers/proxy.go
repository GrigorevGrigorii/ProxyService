package handlers

import (
	"net/http"
	"proxy-service/internal/config"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	Services []config.Service
}

func Ping(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handlers) ProxyGetRequest(c *gin.Context) {
	if !h.allowedToProxy(c.Param("service"), config.HTTPMethod(c.Request.Method), c.Param("path")) {
		c.IndentedJSON(http.StatusForbidden, gin.H{"message": "service or path is not allowed"})
		return
	}
	c.IndentedJSON(http.StatusNotImplemented, gin.H{"message": "not_implemented_error"})
}

func (h *Handlers) ProxyPostRequest(c *gin.Context) {
	if !h.allowedToProxy(c.Param("service"), config.HTTPMethod(c.Request.Method), c.Param("path")) {
		c.IndentedJSON(http.StatusForbidden, gin.H{"message": "service or path is not allowed"})
		return
	}
	c.IndentedJSON(http.StatusNotImplemented, gin.H{"message": "not_implemented_error"})
}

func (h *Handlers) ProxyPutRequest(c *gin.Context) {
	if !h.allowedToProxy(c.Param("service"), config.HTTPMethod(c.Request.Method), c.Param("path")) {
		c.IndentedJSON(http.StatusForbidden, gin.H{"message": "service or path is not allowed"})
		return
	}
	c.IndentedJSON(http.StatusNotImplemented, gin.H{"message": "not_implemented_error"})
}

func (h *Handlers) ProxyDeleteRequest(c *gin.Context) {
	if !h.allowedToProxy(c.Param("service"), config.HTTPMethod(c.Request.Method), c.Param("path")) {
		c.IndentedJSON(http.StatusForbidden, gin.H{"message": "service or path is not allowed"})
		return
	}
	c.IndentedJSON(http.StatusNotImplemented, gin.H{"message": "not_implemented_error"})
}

func (h *Handlers) allowedToProxy(service string, method config.HTTPMethod, path string) bool {
	for _, allowesService := range h.Services {
		if allowesService.Name == service {
			for _, allowedTarget := range allowesService.Targets {
				if allowedTarget.Method == method && allowedTarget.Path == path {
					return true
				}
			}
			return false
		}
	}
	return false
}
