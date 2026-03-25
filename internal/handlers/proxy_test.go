package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"proxy-service/internal/database"
	"proxy-service/internal/models"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type stubHttpClient struct {
	passedCtx           context.Context
	passedURL           string
	passedTimeout       time.Duration
	passedRetryCount    uint8
	passedRetryInterval time.Duration
	resp                *http.Response
	err                 error
}

func (s *stubHttpClient) Get(ctx context.Context, url string, timeout time.Duration, retryCount uint8, retryInterval time.Duration) (*http.Response, error) {
	s.passedCtx = ctx
	s.passedURL = url
	s.passedTimeout = timeout
	s.passedRetryCount = retryCount
	s.passedRetryInterval = retryInterval
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if s.err != nil {
		return nil, s.err
	}
	return s.resp, nil
}

func httpResp(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func testProxyGETRouter(h ProxyHandlers) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/v1/:service/*path", h.ProxyGetRequest)
	return r
}

func testServiceForProxy() database.Service {
	var cacheInterval string = "1m"
	return database.Service{
		Name:          "mock",
		Scheme:        "http",
		Host:          "upstream.example",
		Targets:       []database.Target{{Path: "/mock", Method: "GET", Query: "query=param", CacheInterval: &cacheInterval}},
		Timeout:       1,
		RetryCount:    1,
		RetryInterval: 0,
	}
}

type stubDBRepository struct {
	getFilteredFn func(ctx context.Context, name, path, method, query string) (*database.Service, error)
}

func (s *stubDBRepository) GetAll(ctx context.Context) ([]database.Service, error) {
	return nil, errors.New("unexpected GetAll call")
}

func (s *stubDBRepository) Get(ctx context.Context, name string) (*database.Service, error) {
	return nil, errors.New("unexpected Get call")
}

func (s *stubDBRepository) GetFiltered(ctx context.Context, name, path, method, query string) (*database.Service, error) {
	if s.getFilteredFn == nil {
		return nil, errors.New("unexpected GetFiltered call")
	}
	return s.getFilteredFn(ctx, name, path, method, query)
}

func (s *stubDBRepository) Create(ctx context.Context, service *database.Service) error {
	return errors.New("unexpected Create call")
}

func (s *stubDBRepository) Update(ctx context.Context, service *database.Service) error {
	return errors.New("unexpected Update call")
}

func (s *stubDBRepository) Delete(ctx context.Context, name string) error {
	return errors.New("unexpected Delete call")
}

type stubRedisReposirory struct {
	setFunc func(ctx context.Context, service models.ServiceDTO, target models.TargetDTO, data string, statusCode int, contentType string) error
	getFunc func(ctx context.Context, service models.ServiceDTO, target models.TargetDTO) (string, int, string, error)
}

func (s *stubRedisReposirory) Set(ctx context.Context, service models.ServiceDTO, target models.TargetDTO, data string, statusCode int, contentType string) error {
	if s.setFunc == nil {
		return errors.New("unexpected Set call")
	}
	return s.setFunc(ctx, service, target, data, statusCode, contentType)
}

func (s *stubRedisReposirory) Get(ctx context.Context, service models.ServiceDTO, target models.TargetDTO) (string, int, string, error) {
	if s.getFunc == nil {
		return "", 0, "", errors.New("unexpected Get call")
	}
	return s.getFunc(ctx, service, target)
}

func TestProxyGetRequest_NotAllowedService(t *testing.T) {
	h := ProxyHandlers{
		DBRepository: &stubDBRepository{
			getFilteredFn: func(ctx context.Context, name, path, method, query string) (*database.Service, error) {
				if name != "not-allowed" || path != "/path" || method != http.MethodGet || query != "query=param" {
					t.Fatalf("unexpected args: %s %s %s %s", name, path, method, query)
				}
				return nil, database.ErrNotFound
			},
		},
		HTTPClient: &stubHttpClient{resp: httpResp(http.StatusOK, "no")},
		RedisRepository: &stubRedisReposirory{
			getFunc: func(ctx context.Context, service models.ServiceDTO, target models.TargetDTO) (string, int, string, error) {
				return "", 0, "", errors.New("Data not found")
			},
		},
	}
	r := testProxyGETRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/not-allowed/path?query=param", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestProxyGetRequest_ProxyResponseData(t *testing.T) {
	svc := testServiceForProxy()

	stub := &stubHttpClient{resp: httpResp(http.StatusNotFound, "upstream-body")}
	h := ProxyHandlers{
		DBRepository: &stubDBRepository{
			getFilteredFn: func(ctx context.Context, name, path, method, query string) (*database.Service, error) {
				if name != svc.Name || path != "/mock" || method != http.MethodGet || query != "query=param" {
					t.Fatalf("unexpected args: %s %s %s %s", name, path, method, query)
				}
				return &svc, nil
			},
		},
		HTTPClient: stub,
		RedisRepository: &stubRedisReposirory{
			getFunc: func(ctx context.Context, service models.ServiceDTO, target models.TargetDTO) (string, int, string, error) {
				if service.Name != svc.Name || target.Path != svc.Targets[0].Path || target.Query != svc.Targets[0].Query {
					t.Fatalf("unexpected args: %v %v", service, target)
				}
				return "", 0, "", errors.New("Data not found")
			},
		},
	}
	r := testProxyGETRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/mock/mock?query=param", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	wantURL := "http://upstream.example/mock?query=param"
	if stub.passedURL != wantURL {
		t.Fatalf("got URL = %q, want %q", stub.passedURL, wantURL)
	}
	if stub.passedTimeout != time.Duration(svc.Timeout*float32(time.Second)) {
		t.Fatalf("got timeout = %q, want %q", stub.passedURL, time.Duration(svc.Timeout*float32(time.Second)))
	}
	if stub.passedRetryCount != uint8(svc.RetryCount) {
		t.Fatalf("got retry count = %q, want %q", stub.passedURL, svc.RetryCount)
	}
	if stub.passedRetryInterval != time.Duration(svc.RetryInterval*float32(time.Second)) {
		t.Fatalf("got retry interval = %q, want %q", stub.passedURL, time.Duration(svc.RetryInterval*float32(time.Second)))
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	if got := rec.Body.String(); got != "upstream-body" {
		t.Fatalf("body = %q, want %q", got, "upstream-body")
	}
}

func TestProxyGetRequest_ContextCancelled(t *testing.T) {
	svc := testServiceForProxy()

	h := ProxyHandlers{
		DBRepository: &stubDBRepository{
			getFilteredFn: func(ctx context.Context, name, path, method, query string) (*database.Service, error) {
				if name != svc.Name || path != "/mock" || method != http.MethodGet || query != "query=param" {
					t.Fatalf("unexpected args: %s %s %s %s", name, path, method, query)
				}
				return &svc, nil
			},
		},
		HTTPClient: &stubHttpClient{resp: httpResp(http.StatusOK, "ok")},
		RedisRepository: &stubRedisReposirory{
			getFunc: func(ctx context.Context, service models.ServiceDTO, target models.TargetDTO) (string, int, string, error) {
				if service.Name != svc.Name || target.Path != svc.Targets[0].Path || target.Query != svc.Targets[0].Query {
					t.Fatalf("unexpected args: %v %v", service, target)
				}
				return "", 0, "", errors.New("Data not found")
			},
		},
	}
	r := testProxyGETRouter(h)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/mock/mock?query=param", nil)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	var payload map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json: %v", err)
	}
	if payload["error"] != context.Canceled.Error() {
		t.Fatalf("error field = %q, want canceled", payload["error"])
	}
}

func TestProxyGetRequest_ClientError(t *testing.T) {
	svc := testServiceForProxy()

	stub := &stubHttpClient{err: fmt.Errorf("connection refused")}
	h := ProxyHandlers{
		DBRepository: &stubDBRepository{
			getFilteredFn: func(ctx context.Context, name, path, method, query string) (*database.Service, error) {
				if name != svc.Name || path != "/mock" || method != http.MethodGet || query != "query=param" {
					t.Fatalf("unexpected args: %s %s %s %s", name, path, method, query)
				}
				return &svc, nil
			},
		},
		HTTPClient: stub,
		RedisRepository: &stubRedisReposirory{
			getFunc: func(ctx context.Context, service models.ServiceDTO, target models.TargetDTO) (string, int, string, error) {
				if service.Name != svc.Name || target.Path != svc.Targets[0].Path || target.Query != svc.Targets[0].Query {
					t.Fatalf("unexpected args: %v %v", service, target)
				}
				return "", 0, "", errors.New("Data not found")
			},
		},
	}
	r := testProxyGETRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/mock/mock?query=param", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	var payload map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json: %v", err)
	}
	if payload["error"] != "connection refused" {
		t.Fatalf("error field = %q", payload["error"])
	}
}

// todo
func TestProxyGetRequest_ReturnCachedData(t *testing.T) {
	svc := testServiceForProxy()

	stub := &stubHttpClient{resp: httpResp(http.StatusNotFound, "upstream-body")}
	h := ProxyHandlers{
		DBRepository: &stubDBRepository{
			getFilteredFn: func(ctx context.Context, name, path, method, query string) (*database.Service, error) {
				if name != svc.Name || path != "/mock" || method != http.MethodGet || query != "query=param" {
					t.Fatalf("unexpected args: %s %s %s %s", name, path, method, query)
				}
				return &svc, nil
			},
		},
		HTTPClient: stub,
		RedisRepository: &stubRedisReposirory{
			getFunc: func(ctx context.Context, service models.ServiceDTO, target models.TargetDTO) (string, int, string, error) {
				if service.Name != svc.Name || target.Path != svc.Targets[0].Path || target.Query != svc.Targets[0].Query {
					t.Fatalf("unexpected args: %v %v", service, target)
				}
				return "upstream-body", http.StatusNotFound, "application/json", nil
			},
		},
	}
	r := testProxyGETRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/mock/mock?query=param", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if stub.passedURL != "" {
		t.Fatalf("got URL = %q, want empty", stub.passedURL)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	if got := rec.Body.String(); got != "upstream-body" {
		t.Fatalf("body = %q, want %q", got, "upstream-body")
	}
}
