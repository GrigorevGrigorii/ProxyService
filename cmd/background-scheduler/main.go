package main

import (
	"context"
	"crypto/tls"
	"os"
	"proxy-service/internal/background"
	"proxy-service/internal/config"
	"proxy-service/internal/database"
	"time"

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

	// Redis
	var redisTLSConfig *tls.Config
	if cfg.RedisConfig.EnableTLS {
		redisTLSConfig = &tls.Config{}
	}
	rdb := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:      cfg.RedisConfig.Addrs,
		MasterName: cfg.RedisConfig.GetMasterName(),
		Password:   cfg.RedisConfig.GetPassword(),
		DB:         cfg.RedisConfig.Database,
		PoolSize:   cfg.RedisConfig.PoolSize,
		TLSConfig:  redisTLSConfig,
		ReadOnly:   cfg.RedisConfig.ReadOnly,
	})

	// Asinq
	provider := &background.DynamicProvider{DBRepository: dbRepository}
	mgr, err := asynq.NewPeriodicTaskManager(
		asynq.PeriodicTaskManagerOpts{
			RedisUniversalClient:       rdb,
			PeriodicTaskConfigProvider: provider,
			SyncInterval:               30 * time.Second,
			SchedulerOpts: &asynq.SchedulerOpts{
				Location: time.UTC,
			},
		},
	)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	leaderElection := background.NewLeaderElection(&rdb, mgr)
	leaderElection.RunWithElection(context.Background())
}
