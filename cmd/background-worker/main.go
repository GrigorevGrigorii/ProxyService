package main

import (
	"os"
	"proxy-service/internal/background"
	"proxy-service/internal/client"
	"proxy-service/internal/config"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if os.Getenv("MODE") == "debug" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	// Configs
	cfg, err := config.LoadConfig[config.BackgroundWorkerConfig]()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	// Asynq Redis
	var redisOpt = &asynq.RedisFailoverClientOpt{
		MasterName:    cfg.RedisConfig.MasterName,
		SentinelAddrs: cfg.RedisConfig.Hosts,
		Password:      cfg.RedisConfig.Password,
		DB:            cfg.RedisConfig.Database,
	}

	// Task Handlers
	cacheTask := background.CacheTask{
		HTTPClient: &client.Client{},
	}

	// Asinq
	srv := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency: cfg.Concurrency,
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc("cache_task", cacheTask.Run)

	if err := srv.Run(mux); err != nil {
		log.Fatal().Msg(err.Error())
	}
}
