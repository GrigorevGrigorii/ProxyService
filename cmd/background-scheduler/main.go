package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"proxy-service/internal/config"
	"proxy-service/internal/database"
	"time"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type DynamicProvider struct {
	dbRepository database.DBRepository
}

func (p *DynamicProvider) GetConfigs() ([]*asynq.PeriodicTaskConfig, error) {
	var tasks []*asynq.PeriodicTaskConfig

	servicesToCache, err := p.dbRepository.GetForCaching(context.Background())
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

func main() {
	// Logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if os.Getenv("MODE") == "debug" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	// Configs
	cfg, err := config.LoadConfig[config.BackgroundSchedulerConfig]()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	// Postgres
	db, err := database.InitDB(&cfg.PGConfig)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	dbRepository := database.DBRepository{DB: db}

	// Asynq Redis
	var redis = &asynq.RedisFailoverClientOpt{
		MasterName:    cfg.RedisConfig.MasterName,
		SentinelAddrs: cfg.RedisConfig.Hosts,
		Password:      cfg.RedisConfig.Password,
		DB:            cfg.RedisConfig.Database,
	}

	// Asinq
	provider := &DynamicProvider{dbRepository: dbRepository}
	mgr, err := asynq.NewPeriodicTaskManager(
		asynq.PeriodicTaskManagerOpts{
			RedisConnOpt:               redis,
			PeriodicTaskConfigProvider: provider,
			SyncInterval:               30 * time.Second,
		},
	)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	if err := mgr.Start(); err != nil {
		log.Fatal().Msg(err.Error())
	}

	select {}
}
