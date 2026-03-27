package main

import (
	"crypto/tls"
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
	var redisTLSConfig *tls.Config
	if cfg.RedisConfig.EnableTLS {
		redisTLSConfig = &tls.Config{}
	}
	rdb := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:      cfg.RedisConfig.Addrs,
		MasterName: cfg.RedisConfig.MasterName,
		Password:   cfg.RedisConfig.Password,
		DB:         cfg.RedisConfig.Database,
		PoolSize:   cfg.RedisConfig.PoolSize,
		TLSConfig:  redisTLSConfig,
		ReadOnly:   cfg.RedisConfig.ReadOnly,
	})

	// Task Handlers
	cacheTask := background.CacheTask{
		HTTPClient:      &httpclient.Client{},
		CacheRepository: &cache.CacheRepository{Redis: &rdb},
	}

	// Asinq
	srv := asynq.NewServerFromRedisClient(rdb, asynq.Config{Concurrency: cfg.Concurrency})

	mux := asynq.NewServeMux()
	mux.HandleFunc("cache_task", cacheTask.Run)

	if err := srv.Run(mux); err != nil {
		log.Fatal().Msg(err.Error())
	}
}
