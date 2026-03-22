package client

import (
	"net/http"
	"time"
)

func Get(url string, timeout time.Duration, retryCount uint8, retryInterval time.Duration) (resp *http.Response, err error) {
	var currentRetryCount uint8 = 0
	var client = http.Client{Timeout: timeout}

	for {
		resp, err = client.Get(url)

		if currentRetryCount < retryCount-1 && (err != nil || shouldRetry(resp)) {
			if resp != nil {
				resp.Body.Close()
			}
			currentRetryCount++
			<-time.After(retryInterval)
		} else {
			break
		}
	}

	return resp, err
}

func shouldRetry(resp *http.Response) bool {
	if resp.StatusCode == http.StatusBadGateway ||
		resp.StatusCode == http.StatusServiceUnavailable ||
		resp.StatusCode == http.StatusGatewayTimeout {
		return true
	}
	return false
}
