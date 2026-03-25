package background

import (
	"context"
	"encoding/json"
	"net/url"
	"proxy-service/internal/client"
	"proxy-service/internal/models"
	"time"

	"github.com/hibiken/asynq"
)

type CacheTask struct {
	HTTPClient client.HTTPClient
}

func (t *CacheTask) Run(ctx context.Context, task *asynq.Task) error {
	var service models.ServiceDTO
	if err := json.Unmarshal(task.Payload(), &service); err != nil {
		return err
	}

	targetUrl := url.URL{
		Scheme:   service.Scheme,
		Host:     service.Host,
		Path:     service.Targets[0].Path,
		RawQuery: service.Targets[0].Query,
	}

	resp, err := t.HTTPClient.Get(
		ctx,
		targetUrl.String(),
		time.Duration(service.Timeout*float32(time.Second)),
		uint8(service.RetryCount),
		time.Duration(service.RetryInterval*float32(time.Second)),
	)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	// todo: save to cache and read from cache in proxy.go

	return nil
}
