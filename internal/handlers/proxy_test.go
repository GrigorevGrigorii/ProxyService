package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"proxy-service/internal/config"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestAllowedToProxy(t *testing.T) {
	// config that allows only requests to "mock" service with "GET" method and "/mock" path
	service := config.Service{
		Name: "mock",
		Targets: []config.Target{
			{
				Path:   "/mock",
				Method: "GET",
			},
		},
	}
	service.FillTargetsSet()
	var servicesMap = map[string]config.Service{"mock": service}

	handlers := ProxyHandlers{
		Services: servicesMap,
	}

	tests := []struct {
		serviceName string
		method      string
		path        string
		want        *config.Service
	}{
		{
			serviceName: "mock",
			method:      http.MethodGet,
			path:        "/mock",
			want:        &service,
		},
		{
			serviceName: "unknown", // not allowed by config
			method:      http.MethodGet,
			path:        "/mock",
			want:        nil,
		},
		{
			serviceName: "mock",
			method:      http.MethodPost, // not allowed by config
			path:        "/mock",
			want:        nil,
		},
		{
			serviceName: "mock",
			method:      http.MethodGet,
			path:        "/another/path", // not allowed by config
			want:        nil,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("Test %d", i), func(t *testing.T) {
			got := handlers.getAllowedService(tt.serviceName, tt.method, tt.path)
			if (tt.want == nil) != (got == nil) { // both nil or both not nil
				t.Errorf("For service=%v, method=%v, path=%v got %v, but expected %v", tt.serviceName, tt.method, tt.path, got, tt.want)
				return
			}
			if tt.want != nil && !reflect.DeepEqual(*got, *tt.want) {
				t.Errorf("For service=%v, method=%v, path=%v got %v, but expected %v", tt.serviceName, tt.method, tt.path, got, tt.want)
			}
		})
	}
}

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

func testServiceForProxy() config.Service {
	s := config.Service{
		Name:          "mock",
		Scheme:        "http",
		Host:          "upstream.example",
		Targets:       []config.Target{{Path: "/mock", Method: "GET"}},
		Timeout:       1,
		RetryCount:    1,
		RetryInterval: 0,
	}
	s.FillTargetsSet()
	return s
}

func TestProxyGetRequest_NotAllowedService(t *testing.T) {
	svc := testServiceForProxy()
	services := map[string]config.Service{"mock": svc}
	h := ProxyHandlers{
		Services:   services,
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
	service := testServiceForProxy()
	services := map[string]config.Service{service.Name: service}
	stub := &stubHttpClient{resp: httpResp(http.StatusNotFound, "upstream-body")}
	h := ProxyHandlers{
		Services:   services,
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
	if stub.passedTimeout != time.Duration(service.Timeout*float32(time.Second)) {
		t.Fatalf("got timeout = %q, want %q", stub.passedURL, wantURL)
	}
	if stub.passedRetryCount != service.RetryCount {
		t.Fatalf("got retry count = %q, want %q", stub.passedURL, wantURL)
	}
	if stub.passedRetryInterval != time.Duration(service.RetryInterval*float32(time.Second)) {
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
	services := map[string]config.Service{"mock": svc}
	h := ProxyHandlers{
		Services:   services,
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
	services := map[string]config.Service{"mock": svc}
	stub := &stubHttpClient{err: fmt.Errorf("connection refused")}
	h := ProxyHandlers{
		Services:   services,
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
