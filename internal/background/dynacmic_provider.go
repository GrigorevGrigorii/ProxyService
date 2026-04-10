package background

import (
	"context"
	"encoding/json"
	"proxy-service/internal/models"
	"proxy-service/internal/repository"
	"time"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

type DynamicProvider struct {
	ServiceRepository repository.ServiceRepository
}

func (p *DynamicProvider) GetConfigs() ([]*asynq.PeriodicTaskConfig, error) {
	var tasks []*asynq.PeriodicTaskConfig

	servicesToCache, err := p.ServiceRepository.GetForCaching(context.Background())
	if err != nil {
		return nil, err
	}

	for _, service := range servicesToCache {
		for _, target := range service.Targets {
			if target.CacheInterval == nil {
				continue
			}

			payload, err := getPayload(service, target)
			if err != nil {
				return nil, err
			}

			cacheIntervalDuration, err := time.ParseDuration(*target.CacheInterval)
			if err != nil {
				return nil, err
			}

			task := asynq.NewTask("cache_task", payload)
			tasks = append(
				tasks,
				&asynq.PeriodicTaskConfig{
					Cronspec: "@every " + *target.CacheInterval,
					Task:     task,
					Opts:     []asynq.Option{asynq.Unique(cacheIntervalDuration)},
				},
			)
			log.Info().Msgf("Added task for service %s (path=%s, query=%s)", service.Name, target.Path, target.Query)
		}
	}

	log.Info().Msgf("Return %d tasks for execution", len(tasks))
	return tasks, nil
}

// Return dumped service with the only target that required for this caching task
func getPayload(service models.ServiceDTO, target models.TargetDTO) ([]byte, error) {
	service.Targets = []models.TargetDTO{target}
	return json.Marshal(service)
}
