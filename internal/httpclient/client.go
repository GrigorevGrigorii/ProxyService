package httpclient

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

type HTTPClient interface {
	Get(ctx context.Context, url string, timeout time.Duration, retryCount uint8, retryInterval time.Duration) (resp *http.Response, err error)
}

type Client struct{}

func (c *Client) Get(ctx context.Context, url string, timeout time.Duration, retryCount uint8, retryInterval time.Duration) (*http.Response, error) {
	var reqErr error
	var resp *http.Response
	var currentRetryCount uint8 = 0
	httpClient := &http.Client{Timeout: timeout}

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		log.Info().Msgf("Making GET request to %s", url)
		resp, reqErr = httpClient.Do(req)
		if reqErr != nil {
			log.Error().Msgf("Error while making request to %s: %s", url, reqErr.Error())
		} else if resp.StatusCode >= 400 {
			log.Error().Msgf("Error %d while making request to %s", resp.StatusCode, url)
		}

		if currentRetryCount < retryCount-1 && (reqErr != nil || shouldRetry(resp)) {

			if resp != nil {
				resp.Body.Close()
			}
			currentRetryCount++
			if err := waitRetry(ctx, retryInterval); err != nil {
				return nil, err
			}
		} else {
			break
		}
	}

	return resp, reqErr
}

func waitRetry(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return nil
		}
	}
	t := time.NewTimer(d)
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		if !t.Stop() {
			select {
			case <-t.C:
			default:
			}
		}
		return ctx.Err()
	}
}

func shouldRetry(resp *http.Response) bool {
	if resp.StatusCode == http.StatusBadGateway ||
		resp.StatusCode == http.StatusServiceUnavailable ||
		resp.StatusCode == http.StatusGatewayTimeout {
		return true
	}
	return false
}
