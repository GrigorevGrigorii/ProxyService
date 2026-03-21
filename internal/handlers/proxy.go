package handlers

import (
	"net/http"
	"proxy-service/internal/config"

	"github.com/gin-gonic/gin"
)

type ProxyHandlers struct {
	Services map[string]config.Service
}

func Ping(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *ProxyHandlers) ProxyGetRequest(c *gin.Context) {
	if !h.allowedToProxy(c.Param("service"), config.MethodGet, c.Param("path")) {
		c.IndentedJSON(http.StatusForbidden, gin.H{"message": "service or path is not allowed"})
		return
	}
	c.IndentedJSON(http.StatusNotImplemented, gin.H{"message": "not_implemented_error"})
}

func (h *ProxyHandlers) ProxyPostRequest(c *gin.Context) {
	if !h.allowedToProxy(c.Param("service"), config.MethodPost, c.Param("path")) {
		c.IndentedJSON(http.StatusForbidden, gin.H{"message": "service or path is not allowed"})
		return
	}
	c.IndentedJSON(http.StatusNotImplemented, gin.H{"message": "not_implemented_error"})
}

func (h *ProxyHandlers) ProxyPutRequest(c *gin.Context) {
	if !h.allowedToProxy(c.Param("service"), config.MethodPut, c.Param("path")) {
		c.IndentedJSON(http.StatusForbidden, gin.H{"message": "service or path is not allowed"})
		return
	}
	c.IndentedJSON(http.StatusNotImplemented, gin.H{"message": "not_implemented_error"})
}

func (h *ProxyHandlers) ProxyDeleteRequest(c *gin.Context) {
	if !h.allowedToProxy(c.Param("service"), config.MethodDelete, c.Param("path")) {
		c.IndentedJSON(http.StatusForbidden, gin.H{"message": "service or path is not allowed"})
		return
	}
	c.IndentedJSON(http.StatusNotImplemented, gin.H{"message": "not_implemented_error"})
}

func (h *ProxyHandlers) allowedToProxy(service string, method config.HTTPMethod, path string) bool {
	allowesService, ok := h.Services[service]
	if !ok { // service not found
		return false
	}

	_, ok = allowesService.TargetsSet[config.Target{Method: method, Path: path}]
	if !ok { // target in service not found
		return false
	}

	return true
}
