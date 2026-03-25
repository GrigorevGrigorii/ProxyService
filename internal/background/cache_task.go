package background

import (
	"context"
	"encoding/json"
	"io"
	"net/url"
	"proxy-service/internal/httpclient"
	"proxy-service/internal/models"
	"proxy-service/internal/redis_repositories"
	"time"

	"github.com/hibiken/asynq"
)

type CacheTask struct {
	HTTPClient      httpclient.HTTPClient
	RedisRepository redis_repositories.Repository
}

func (t *CacheTask) Run(ctx context.Context, task *asynq.Task) error {
	var service models.ServiceDTO
	if err := json.Unmarshal(task.Payload(), &service); err != nil {
		return err
	}
	target := service.Targets[0]

	targetUrl := url.URL{
		Scheme:   service.Scheme,
		Host:     service.Host,
		Path:     target.Path,
		RawQuery: target.Query,
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := t.RedisRepository.Set(ctx, service, target, string(body), resp.StatusCode, resp.Header.Get("Content-Type")); err != nil {
		return err
	}

	return nil
}
