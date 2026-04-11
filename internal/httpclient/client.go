package httpclient

import (
	"context"
	"net/http"
	"proxy-service/internal/utils"
	"sync"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

var (
	sharedTransport *http.Transport
	transportOnce   sync.Once
)

func getSharedTransport() *http.Transport {
	transportOnce.Do(func() {
		sharedTransport = &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		}
	})
	return sharedTransport
}

func newRetryableClient(ctx context.Context, timeout time.Duration, retryCount int, retryInterval time.Duration) *retryablehttp.Client {
	var logger retryablehttp.LeveledLogger = &retryableHTTPLeveledLoggerAgapter{logger: utils.GetLogger(ctx)}
	return &retryablehttp.Client{
		HTTPClient: &http.Client{
			Timeout:   timeout,
			Transport: getSharedTransport(),
		},
		Logger:       logger,
		RetryMax:     retryCount,
		RetryWaitMin: retryInterval,
		RetryWaitMax: retryInterval,
		CheckRetry:   retryablehttp.DefaultRetryPolicy,
		Backoff:      retryablehttp.DefaultBackoff,
	}
}

type HTTPClient interface {
	Get(ctx context.Context, url string, timeout time.Duration, retryCount int, retryInterval time.Duration) (resp *http.Response, err error)
}

type Client struct{}

func (c Client) Get(ctx context.Context, url string, timeout time.Duration, retryCount int, retryInterval time.Duration) (*http.Response, error) {
	retryableClient := newRetryableClient(ctx, timeout, retryCount, retryInterval)

	req, err := retryablehttp.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	return retryableClient.Do(req)
}
