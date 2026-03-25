package background

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"proxy-service/internal/client"
	"proxy-service/internal/models"
	"time"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

type CacheTask struct {
	HTTPClient client.HTTPClient
	Redis      *redis.ClusterClient
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

	t.Redis.HSet(
		ctx,
		fmt.Sprintf("%s:%s:%s:%s", service.Name, target.Path, target.Method, target.Query),
		map[string]any{
			"data":         string(body),
			"status_code":  resp.StatusCode,
			"content_type": resp.Header.Get("Content-Type"),
		},
	)

	return nil
}
