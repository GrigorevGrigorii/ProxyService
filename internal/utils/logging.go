package utils

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func GetLogger(ctx context.Context) *zerolog.Logger {
	logger := zerolog.Ctx(ctx)
	if logger != nil {
		return logger
	}

	return &log.Logger
}
