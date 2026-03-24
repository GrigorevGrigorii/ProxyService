package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"proxy-service/internal/database"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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

func newMockRepository(t *testing.T) (*database.DBRepository, sqlmock.Sqlmock) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	gdb, err := gorm.Open(postgres.New(postgres.Config{
		Conn:                 sqlDB,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open: %v", err)
	}

	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet sql expectations: %v", err)
		}
	})

	return &database.DBRepository{DB: gdb}, mock
}

func expectGetFilteredSuccess(mock sqlmock.Sqlmock, svc database.Service, path, method string) {
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "services" WHERE name = \$1 ORDER BY "services"\."name" LIMIT \$2`).
		WithArgs(svc.Name, 1).
		WillReturnRows(sqlmock.NewRows([]string{"name", "scheme", "host", "timeout", "retry_count", "retry_interval", "version"}).
			AddRow(svc.Name, svc.Scheme, svc.Host, svc.Timeout, svc.RetryCount, svc.RetryInterval, svc.Version))
	mock.ExpectQuery(`SELECT \* FROM "targets" WHERE "targets"\."service_name" = \$1 AND \(path = \$2 AND method = \$3\)`).
		WithArgs(svc.Name, path, method).
		WillReturnRows(sqlmock.NewRows([]string{"service_name", "path", "method"}).
			AddRow(svc.Name, path, method))
	mock.ExpectCommit()
}

func expectGetFilteredError(mock sqlmock.Sqlmock, serviceName string, err error) {
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "services" WHERE name = \$1 ORDER BY "services"\."name" LIMIT \$2`).
		WithArgs(serviceName, 1).
		WillReturnError(err)
	mock.ExpectRollback()
}

func TestProxyGetRequest_NotAllowedService(t *testing.T) {
	repo, mock := newMockRepository(t)
	expectGetFilteredError(mock, "not-allowed", database.ErrNotFound)

	h := ProxyHandlers{
		DBRepository: repo,
		HTTPClient:   &stubHttpClient{resp: httpResp(http.StatusOK, "no")},
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
	repo, mock := newMockRepository(t)
	expectGetFilteredSuccess(mock, svc, "/mock", http.MethodGet)

	stub := &stubHttpClient{resp: httpResp(http.StatusNotFound, "upstream-body")}
	h := ProxyHandlers{
		DBRepository: repo,
		HTTPClient:   stub,
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
	repo, mock := newMockRepository(t)
	expectGetFilteredSuccess(mock, svc, "/mock", http.MethodGet)

	h := ProxyHandlers{
		DBRepository: repo,
		HTTPClient:   &stubHttpClient{resp: httpResp(http.StatusOK, "ok")},
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
	repo, mock := newMockRepository(t)
	expectGetFilteredSuccess(mock, svc, "/mock", http.MethodGet)

	stub := &stubHttpClient{err: fmt.Errorf("connection refused")}
	h := ProxyHandlers{
		DBRepository: repo,
		HTTPClient:   stub,
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
