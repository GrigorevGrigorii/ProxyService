package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"proxy-service/internal/cache"
	"proxy-service/internal/database"
	"proxy-service/internal/httpclient"
	"proxy-service/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type ProxyHandlers struct {
	DBRepository    database.Repository
	HTTPClient      httpclient.HTTPClient
	CacheRepository cache.Repository
}

func (h *ProxyHandlers) ProxyGetRequest(c *gin.Context) {
	service, err := h.getAllowedService(c.Request.Context(), c.Param("service"), http.MethodGet, c.Param("path"), c.Request.URL.Query().Encode())
	if errors.Is(err, database.ErrNotFound) {
		c.IndentedJSON(http.StatusForbidden, gin.H{"message": "service, path or query params are not allowed"})
		return
	}
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	target := service.Targets[0]

	if target.CacheInterval != nil {
		data, statusCode, contentType, err := h.CacheRepository.Get(
			c.Request.Context(), models.ServiceDTOFromDBModel(*service), models.TargetDTOFromDBModel(target),
		)
		if err == nil {
			log.Info().Msgf("Use cache response for %s, %s, %s, %s", service.Name, target.Path, http.MethodGet, target.Query)
			c.Data(statusCode, contentType, []byte(data))
			return
		} else {
			log.Warn().Msgf("Got error from redis for %s, %s, %s, %s: %s", service.Name, target.Path, http.MethodGet, target.Query, err.Error())
		}
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
	_, err := h.getAllowedService(c.Request.Context(), c.Param("service"), http.MethodGet, c.Param("path"), "")
	if errors.Is(err, database.ErrNotFound) {
		c.IndentedJSON(http.StatusForbidden, gin.H{"message": "service, path or query params are not allowed"})
		return
	}
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusNotImplemented, gin.H{"message": "not_implemented_error"})
}

func (h *ProxyHandlers) ProxyPutRequest(c *gin.Context) {
	_, err := h.getAllowedService(c.Request.Context(), c.Param("service"), http.MethodGet, c.Param("path"), "")
	if errors.Is(err, database.ErrNotFound) {
		c.IndentedJSON(http.StatusForbidden, gin.H{"message": "service, path or query params are not allowed"})
		return
	}
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusNotImplemented, gin.H{"message": "not_implemented_error"})
}

func (h *ProxyHandlers) ProxyDeleteRequest(c *gin.Context) {
	_, err := h.getAllowedService(c.Request.Context(), c.Param("service"), http.MethodGet, c.Param("path"), "")
	if errors.Is(err, database.ErrNotFound) {
		c.IndentedJSON(http.StatusForbidden, gin.H{"message": "service, path or query params are not allowed"})
		return
	}
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusNotImplemented, gin.H{"message": "not_implemented_error"})
}

func (h *ProxyHandlers) getAllowedService(ctx context.Context, service, method, path, query string) (*database.Service, error) {
	allowedService, err := h.DBRepository.GetFiltered(ctx, service, path, method, query)
	if err != nil {
		return nil, err
	}

	return allowedService, nil
}
