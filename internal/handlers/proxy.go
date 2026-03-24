package handlers

import (
	"errors"
	"net/http"
	"net/url"
	"proxy-service/internal/client"
	"proxy-service/internal/database"
	"time"

	"github.com/gin-gonic/gin"
)

type ProxyHandlers struct {
	DBRepository database.Repository
	HTTPClient   client.HTTPClient
}

func (h *ProxyHandlers) ProxyGetRequest(c *gin.Context) {
	service, err := h.getAllowedService(c.Param("service"), http.MethodGet, c.Param("path"))
	if errors.Is(err, database.ErrNotFound) {
		c.IndentedJSON(http.StatusForbidden, gin.H{"message": "service or path is not allowed"})
		return
	}
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	targetUrl := url.URL{
		Scheme:   service.Scheme,
		Host:     service.Host,
		Path:     c.Param("path"),
		RawQuery: c.Request.URL.RawQuery,
	}

	resp, err := h.HTTPClient.Get(
		c.Request.Context(),
		targetUrl.String(),
		time.Duration(service.Timeout*float32(time.Second)),
		uint8(service.RetryCount),
		time.Duration(service.RetryInterval*float32(time.Second)),
	)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	defer resp.Body.Close()

	c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
}

func (h *ProxyHandlers) ProxyPostRequest(c *gin.Context) {
	_, err := h.getAllowedService(c.Param("service"), http.MethodGet, c.Param("path"))
	if errors.Is(err, database.ErrNotFound) {
		c.IndentedJSON(http.StatusForbidden, gin.H{"message": "service or path is not allowed"})
		return
	}
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusNotImplemented, gin.H{"message": "not_implemented_error"})
}

func (h *ProxyHandlers) ProxyPutRequest(c *gin.Context) {
	_, err := h.getAllowedService(c.Param("service"), http.MethodGet, c.Param("path"))
	if errors.Is(err, database.ErrNotFound) {
		c.IndentedJSON(http.StatusForbidden, gin.H{"message": "service or path is not allowed"})
		return
	}
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusNotImplemented, gin.H{"message": "not_implemented_error"})
}

func (h *ProxyHandlers) ProxyDeleteRequest(c *gin.Context) {
	_, err := h.getAllowedService(c.Param("service"), http.MethodGet, c.Param("path"))
	if errors.Is(err, database.ErrNotFound) {
		c.IndentedJSON(http.StatusForbidden, gin.H{"message": "service or path is not allowed"})
		return
	}
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusNotImplemented, gin.H{"message": "not_implemented_error"})
}

func (h *ProxyHandlers) getAllowedService(service string, method string, path string) (*database.Service, error) {
	allowedService, err := h.DBRepository.GetFiltered(service, path, method)
	if err != nil {
		return nil, err
	}

	return allowedService, nil
}
