package handlers

import (
	"io"
	"net/http"
	"net/url"
	"proxy-service/internal/client"
	"proxy-service/internal/config"
	"time"

	"github.com/gin-gonic/gin"
)

type ProxyHandlers struct {
	Services   *map[string]config.Service
	HttpClient client.HTTPClient
}

func Ping(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *ProxyHandlers) ProxyGetRequest(c *gin.Context) {
	service := h.getAllowedService(c.Param("service"), http.MethodGet, c.Param("path"))
	if service == nil {
		c.IndentedJSON(http.StatusForbidden, gin.H{"message": "service or path is not allowed"})
		return
	}

	targetUrl := url.URL{
		Scheme:   service.Scheme,
		Host:     service.Host,
		Path:     c.Param("path"),
		RawQuery: c.Request.URL.RawQuery,
	}

	resp, err := h.HttpClient.Get(
		targetUrl.String(),
		time.Duration(service.Timeout*float32(time.Second)),
		service.RetryCount,
		time.Duration(service.RetryInterval*float32(time.Second)),
	)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.String(resp.StatusCode, string(body))
}

func (h *ProxyHandlers) ProxyPostRequest(c *gin.Context) {
	service := h.getAllowedService(c.Param("service"), http.MethodPost, c.Param("path"))
	if service == nil {
		c.IndentedJSON(http.StatusForbidden, gin.H{"message": "service or path is not allowed"})
		return
	}
	c.IndentedJSON(http.StatusNotImplemented, gin.H{"message": "not_implemented_error"})
}

func (h *ProxyHandlers) ProxyPutRequest(c *gin.Context) {
	service := h.getAllowedService(c.Param("service"), http.MethodPut, c.Param("path"))
	if service == nil {
		c.IndentedJSON(http.StatusForbidden, gin.H{"message": "service or path is not allowed"})
		return
	}
	c.IndentedJSON(http.StatusNotImplemented, gin.H{"message": "not_implemented_error"})
}

func (h *ProxyHandlers) ProxyDeleteRequest(c *gin.Context) {
	service := h.getAllowedService(c.Param("service"), http.MethodDelete, c.Param("path"))
	if service == nil {
		c.IndentedJSON(http.StatusForbidden, gin.H{"message": "service or path is not allowed"})
		return
	}
	c.IndentedJSON(http.StatusNotImplemented, gin.H{"message": "not_implemented_error"})
}

func (h *ProxyHandlers) getAllowedService(service string, method string, path string) *config.Service {
	allowedService, ok := (*h.Services)[service]
	if !ok { // service not found
		return nil
	}

	_, ok = allowedService.TargetsSet[config.Target{Method: method, Path: path}]
	if !ok { // target in service not found
		return nil
	}

	return &allowedService
}
