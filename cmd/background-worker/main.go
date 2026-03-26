package main

import (
	"os"
	"proxy-service/internal/background"
	"proxy-service/internal/cache"
	"proxy-service/internal/config"
	"proxy-service/internal/httpclient"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
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

	// Redis
	rdb := redis.NewFailoverClusterClient(&redis.FailoverOptions{
		MasterName:    cfg.RedisConfig.MasterName,
		SentinelAddrs: cfg.RedisConfig.Hosts,
		Password:      cfg.RedisConfig.Password,
		DB:            cfg.RedisConfig.Database,
		PoolSize:      cfg.RedisConfig.PoolSize,
	})

	// Task Handlers
	cacheTask := background.CacheTask{
		HTTPClient:      &httpclient.Client{},
		CacheRepository: &cache.CacheRepository{Redis: rdb},
	}

	// Asynq Redis
	var redisOpt = &asynq.RedisFailoverClientOpt{
		MasterName:    cfg.RedisConfig.MasterName,
		SentinelAddrs: cfg.RedisConfig.Hosts,
		Password:      cfg.RedisConfig.Password,
		DB:            cfg.RedisConfig.Database,
		PoolSize:      cfg.RedisConfig.PoolSize,
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
