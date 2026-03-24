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
	return database.Service{
		Name:          "mock",
		Scheme:        "http",
		Host:          "upstream.example",
		Targets:       []database.Target{{Path: "/mock", Method: "GET"}},
		Timeout:       1,
		RetryCount:    1,
		RetryInterval: 0,
	}
}

type stubProxyRepository struct {
	getFilteredFn func(name, path, method string) (*database.Service, error)
}

func (s *stubProxyRepository) GetAll() ([]database.Service, error) {
	return nil, errors.New("unexpected GetAll call")
}

func (s *stubProxyRepository) Get(name string) (*database.Service, error) {
	return nil, errors.New("unexpected Get call")
}

func (s *stubProxyRepository) GetFiltered(name, path, method string) (*database.Service, error) {
	if s.getFilteredFn == nil {
		return nil, errors.New("unexpected GetFiltered call")
	}
	return s.getFilteredFn(name, path, method)
}

func (s *stubProxyRepository) Create(service *database.Service) error {
	return errors.New("unexpected Create call")
}

func (s *stubProxyRepository) Update(service *database.Service) error {
	return errors.New("unexpected Update call")
}

func (s *stubProxyRepository) Delete(name string) error {
	return errors.New("unexpected Delete call")
}

func TestProxyGetRequest_NotAllowedService(t *testing.T) {
	h := ProxyHandlers{
		DBRepository: &stubProxyRepository{
			getFilteredFn: func(name, path, method string) (*database.Service, error) {
				if name != "not-allowed" || path != "/path" || method != http.MethodGet {
					t.Fatalf("unexpected args: %s %s %s", name, path, method)
				}
				return nil, database.ErrNotFound
			},
		},
		HTTPClient: &stubHttpClient{resp: httpResp(http.StatusOK, "no")},
	}
	r := testProxyGETRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/not-allowed/path", nil)
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
		DBRepository: &stubProxyRepository{
			getFilteredFn: func(name, path, method string) (*database.Service, error) {
				if name != svc.Name || path != "/mock" || method != http.MethodGet {
					t.Fatalf("unexpected args: %s %s %s", name, path, method)
				}
				return &svc, nil
			},
		},
		HTTPClient: stub,
	}
	r := testProxyGETRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/mock/mock?x=1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	wantURL := "http://upstream.example/mock?x=1"
	if stub.passedURL != wantURL {
		t.Fatalf("got URL = %q, want %q", stub.passedURL, wantURL)
	}
	if stub.passedTimeout != time.Duration(svc.Timeout*float32(time.Second)) {
		t.Fatalf("got timeout = %q, want %q", stub.passedURL, wantURL)
	}
	if stub.passedRetryCount != uint8(svc.RetryCount) {
		t.Fatalf("got retry count = %q, want %q", stub.passedURL, wantURL)
	}
	if stub.passedRetryInterval != time.Duration(svc.RetryInterval*float32(time.Second)) {
		t.Fatalf("got retry interval = %q, want %q", stub.passedURL, wantURL)
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
		DBRepository: &stubProxyRepository{
			getFilteredFn: func(name, path, method string) (*database.Service, error) {
				if name != svc.Name || path != "/mock" || method != http.MethodGet {
					t.Fatalf("unexpected args: %s %s %s", name, path, method)
				}
				return &svc, nil
			},
		},
		HTTPClient: &stubHttpClient{resp: httpResp(http.StatusOK, "ok")},
	}
	r := testProxyGETRouter(h)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/mock/mock", nil)
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
		DBRepository: &stubProxyRepository{
			getFilteredFn: func(name, path, method string) (*database.Service, error) {
				if name != svc.Name || path != "/mock" || method != http.MethodGet {
					t.Fatalf("unexpected args: %s %s %s", name, path, method)
				}
				return &svc, nil
			},
		},
		HTTPClient: stub,
	}
	r := testProxyGETRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/mock/mock", nil)
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
