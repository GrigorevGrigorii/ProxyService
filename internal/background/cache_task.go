package background

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"proxy-service/internal/cache"
	"proxy-service/internal/httpclient"
	"proxy-service/internal/models"
	"time"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

type CacheTask struct {
	HTTPClient      httpclient.HTTPClient
	CacheRepository cache.Repository
}

func (t *CacheTask) Run(ctx context.Context, task *asynq.Task) error {
	log.Info().Msgf("Running task with payload %s", string(task.Payload()))

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

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Error().Msgf("Got %d for %s, %s, %s. Do not save to cache", resp.StatusCode, service.Name, target.Path, target.Query)
		return errors.New("Unsuccessful http response")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := t.CacheRepository.Set(ctx, service, target, string(body), resp.StatusCode, resp.Header.Get("Content-Type")); err != nil {
		return err
	}

	log.Info().Msgf("Successfully saved response to cache for %s, %s, %s", service.Name, target.Path, target.Query)
	return nil
}
