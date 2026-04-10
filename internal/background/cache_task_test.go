package background

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"proxy-service/internal/models"
	"strings"
	"testing"
	"time"

	"github.com/hibiken/asynq"
)

func testService() models.ServiceDTO {
	var cacheInterval string = "1m"
	return models.ServiceDTO{
		Name:          "mock",
		Scheme:        "http",
		Host:          "upstream.example",
		Targets:       []models.TargetDTO{{Path: "/mock", Method: "GET", Query: "query=param", CacheInterval: &cacheInterval}},
		Timeout:       1,
		RetryCount:    1,
		RetryInterval: 0,
	}
}

type stubHttpClient struct {
	passedCtx           context.Context
	passedURL           string
	passedTimeout       time.Duration
	passedRetryCount    int
	passedRetryInterval time.Duration
	resp                *http.Response
	err                 error
}

func (s *stubHttpClient) Get(ctx context.Context, url string, timeout time.Duration, retryCount int, retryInterval time.Duration) (*http.Response, error) {
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

type stubCacheRepository struct {
	setFunc func(ctx context.Context, service models.ServiceDTO, target models.TargetDTO, data string, statusCode int, contentType string) error
	getFunc func(ctx context.Context, service models.ServiceDTO, target models.TargetDTO) (string, int, string, error)
}

func (s *stubCacheRepository) Set(ctx context.Context, service models.ServiceDTO, target models.TargetDTO, data string, statusCode int, contentType string) error {
	if s.setFunc == nil {
		return errors.New("unexpected Set call")
	}
	return s.setFunc(ctx, service, target, data, statusCode, contentType)
}

func (s *stubCacheRepository) Get(ctx context.Context, service models.ServiceDTO, target models.TargetDTO) (string, int, string, error) {
	if s.getFunc == nil {
		return "", 0, "", errors.New("unexpected Get call")
	}
	return s.getFunc(ctx, service, target)
}

func TestCacheTask_Success(t *testing.T) {
	svc := testService()

	stub := &stubHttpClient{resp: httpResp(http.StatusOK, "upstream-body")}
	cacheTask := CacheTask{
		HTTPClient: stub,
		CacheRepository: &stubCacheRepository{
			setFunc: func(ctx context.Context, service models.ServiceDTO, target models.TargetDTO, data string, statusCode int, contentType string) error {
				if service.Name != svc.Name || target.Path != svc.Targets[0].Path || target.Query != svc.Targets[0].Query || data != "upstream-body" || statusCode != http.StatusOK {
					t.Fatalf("unexpected args: %v, %v, %s %d", service, target, data, statusCode)
				}
				return nil
			},
		},
	}

	payload, _ := json.Marshal(svc)
	task := asynq.NewTask("cache_task", payload)
	err := cacheTask.Run(context.Background(), task)

	if err != nil {
		t.Fatalf("got err from task: %s", err.Error())
	}

	wantURL := "http://upstream.example/mock?query=param"
	if stub.passedURL != wantURL {
		t.Fatalf("got URL = %q, want %q", stub.passedURL, wantURL)
	}
	if stub.passedTimeout != time.Duration(svc.Timeout*float32(time.Second)) {
		t.Fatalf("got timeout = %q, want %q", stub.passedURL, time.Duration(svc.Timeout*float32(time.Second)))
	}
	if stub.passedRetryCount != svc.RetryCount {
		t.Fatalf("got retry count = %q, want %q", stub.passedURL, svc.RetryCount)
	}
	if stub.passedRetryInterval != time.Duration(svc.RetryInterval*float32(time.Second)) {
		t.Fatalf("got retry interval = %q, want %q", stub.passedURL, time.Duration(svc.RetryInterval*float32(time.Second)))
	}
}

func TestCacheTask_BadStatusCode(t *testing.T) {
	svc := testService()

	stub := &stubHttpClient{resp: httpResp(http.StatusInternalServerError, "upstream-body")}
	cacheTask := CacheTask{
		HTTPClient:      stub,
		CacheRepository: &stubCacheRepository{},
	}

	payload, _ := json.Marshal(svc)
	task := asynq.NewTask("cache_task", payload)
	err := cacheTask.Run(context.Background(), task)

	if err.Error() != "Unsuccessful http response" {
		t.Fatalf("got unexpected err from task: %s", err.Error())
	}
}

func TestCacheTask_ErrorFromHttpClient(t *testing.T) {
	svc := testService()
	httpError := fmt.Errorf("connection refused")

	stub := &stubHttpClient{err: httpError}
	cacheTask := CacheTask{
		HTTPClient:      stub,
		CacheRepository: &stubCacheRepository{},
	}

	payload, _ := json.Marshal(svc)
	task := asynq.NewTask("cache_task", payload)
	err := cacheTask.Run(context.Background(), task)

	if !errors.Is(err, httpError) {
		t.Fatalf("got unexpected err from task: %s", err.Error())
	}
}
