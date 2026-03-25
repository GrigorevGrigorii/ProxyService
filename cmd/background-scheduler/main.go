package main

import (
	"os"
	"proxy-service/internal/background"
	"proxy-service/internal/config"
	"proxy-service/internal/database"
	"time"

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
	var redisOpt = &asynq.RedisFailoverClientOpt{
		MasterName:    cfg.RedisConfig.MasterName,
		SentinelAddrs: cfg.RedisConfig.Hosts,
		Password:      cfg.RedisConfig.Password,
		DB:            cfg.RedisConfig.Database,
	}

	// Asinq
	provider := &background.DynamicProvider{DBRepository: dbRepository}
	mgr, err := asynq.NewPeriodicTaskManager(
		asynq.PeriodicTaskManagerOpts{
			RedisConnOpt:               redisOpt,
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

	if err := mgr.Start(); err != nil {
		log.Fatal().Msg(err.Error())
	}

	select {}
}
