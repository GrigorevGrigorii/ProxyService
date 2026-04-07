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
	"proxy-service/internal/utils"
	"time"

	"github.com/gin-gonic/gin"
)

type ProxyHandlers struct {
	DBRepository    database.Repository
	HTTPClient      httpclient.HTTPClient
	CacheRepository cache.Repository
}

// ProxyGetRequest godoc
//
//	@Summary		Proxy GET request
//	@Tags			Proxy API
//	@Description	Proxy GET request to service with all provided query params
//	@Param			service	path	string	true	"Service name"
//	@Param			path	path	string	true	"Path of target to proxy"
//	@Router			/v1/{service}/{path} [get]
func (h *ProxyHandlers) ProxyGetRequest(c *gin.Context) {
	log := utils.GetLogger(c.Request.Context())

	service, err := h.getAllowedService(c.Request.Context(), c.Param("service"), http.MethodGet, c.Param("path"), c.Request.URL.Query().Encode())
	if errors.Is(err, database.ErrNotFound) {
		c.JSON(http.StatusForbidden, MessageResponse{Message: "service, path or query params are not allowed"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Message: err.Error()})
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
		c.JSON(http.StatusInternalServerError, MessageResponse{Message: err.Error()})
		return
	}

	defer resp.Body.Close()

	c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
}

// ProxyPostRequest godoc
//
//	@Summary		[not_implemented] Proxy POST request
//	@Tags			Proxy API
//	@Description	Proxy POST request to service
//	@Param			service	path	string	true	"Service name"
//	@Param			path	path	string	true	"Path of target to proxy"
//	@Router			/v1/{service}/{path} [post]
func (h *ProxyHandlers) ProxyPostRequest(c *gin.Context) {
	_, err := h.getAllowedService(c.Request.Context(), c.Param("service"), http.MethodGet, c.Param("path"), "")
	if errors.Is(err, database.ErrNotFound) {
		c.JSON(http.StatusForbidden, MessageResponse{Message: "service, path or query params are not allowed"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Message: err.Error()})
		return
	}
	c.JSON(http.StatusNotImplemented, MessageResponse{Message: "not_implemented_error"})
}

// ProxyPutRequest godoc
//
//	@Summary		[not_implemented] Proxy PUT request
//	@Tags			Proxy API
//	@Description	Proxy PUT request to service
//	@Param			service	path	string	true	"Service name"
//	@Param			path	path	string	true	"Path of target to proxy"
//	@Router			/v1/{service}/{path} [put]
func (h *ProxyHandlers) ProxyPutRequest(c *gin.Context) {
	_, err := h.getAllowedService(c.Request.Context(), c.Param("service"), http.MethodGet, c.Param("path"), "")
	if errors.Is(err, database.ErrNotFound) {
		c.JSON(http.StatusForbidden, MessageResponse{Message: "service, path or query params are not allowed"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Message: err.Error()})
		return
	}
	c.JSON(http.StatusNotImplemented, MessageResponse{Message: "not_implemented_error"})
}

// ProxyDeleteRequest godoc
//
//	@Summary		[not_implemented] Proxy DELETE request
//	@Tags			Proxy API
//	@Description	Proxy DELETE request to service
//	@Param			service	path	string	true	"Service name"
//	@Param			path	path	string	true	"Path of target to proxy"
//	@Router			/v1/{service}/{path} [delete]
func (h *ProxyHandlers) ProxyDeleteRequest(c *gin.Context) {
	_, err := h.getAllowedService(c.Request.Context(), c.Param("service"), http.MethodGet, c.Param("path"), "")
	if errors.Is(err, database.ErrNotFound) {
		c.JSON(http.StatusForbidden, MessageResponse{Message: "service, path or query params are not allowed"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Message: err.Error()})
		return
	}
	c.JSON(http.StatusNotImplemented, MessageResponse{Message: "not_implemented_error"})
}

func (h *ProxyHandlers) getAllowedService(ctx context.Context, service, method, path, query string) (*database.Service, error) {
	allowedService, err := h.DBRepository.GetFiltered(ctx, service, path, method, query)
	if err != nil {
		return nil, err
	}

	return allowedService, nil
}
