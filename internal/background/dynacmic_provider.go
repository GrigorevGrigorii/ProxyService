package background

import (
	"context"
	"encoding/json"
	"fmt"
	"proxy-service/internal/database"
	"time"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

type DynamicProvider struct {
	DBRepository database.DBRepository
}

func (p *DynamicProvider) GetConfigs() ([]*asynq.PeriodicTaskConfig, error) {
	var tasks []*asynq.PeriodicTaskConfig

	servicesToCache, err := p.DBRepository.GetForCaching(context.Background())
	if err != nil {
		return nil, err
	}

	for _, service := range servicesToCache {
		for _, target := range service.Targets {
			duration, err := time.ParseDuration(*target.CacheInterval)
			if err != nil {
				return nil, err
			}

			payload, err := json.Marshal(service)
			if err != nil {
				return nil, err
			}

			task := asynq.NewTask(fmt.Sprintf("%s:%s:%s", service.Name, target.Path, target.Query), payload, asynq.Unique(duration))
			tasks = append(
				tasks,
				&asynq.PeriodicTaskConfig{
					Cronspec: fmt.Sprintf("@every %s", *target.CacheInterval),
					Task:     task,
				},
			)
			log.Info().Msgf("Added task for service %s (path=%s, query=%s)", service.Name, target.Path, target.Query)
		}
	}

	log.Info().Msgf("Return %d tasks for execution", len(tasks))
	return tasks, nil
}
