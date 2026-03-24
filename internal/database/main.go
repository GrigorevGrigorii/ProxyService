package database

import (
	"fmt"
	"net/url"
	"proxy-service/internal/config"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	maxIdleConns    = 10
	maxOpenConns    = 100
	connMaxLifetime = time.Hour
	connMaxIdleTime = 10 * time.Minute
)

func InitDB(cfg *config.PGConfig) (*gorm.DB, error) {
	log.Info().
		Msgf(
			"Init DB with user=%s, host=%s, port=%d, database=%s, sslmode=%s",
			cfg.Username,
			cfg.Host,
			cfg.Port,
			cfg.Database,
			cfg.SSLMode,
		)

	pgDsn := url.URL{
		Scheme:   "postgresql",
		User:     url.UserPassword(cfg.Username, cfg.Password),
		Host:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Path:     cfg.Database,
		RawQuery: "sslmode=" + cfg.SSLMode,
	}
	db, err := gorm.Open(postgres.Open(pgDsn.String()), &gorm.Config{TranslateError: true})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)
	sqlDB.SetConnMaxIdleTime(connMaxIdleTime)

	return db, nil
}
