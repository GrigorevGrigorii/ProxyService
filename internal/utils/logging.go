package utils

import (
	"context"

	"github.com/rs/zerolog"
)

func GetLogger(ctx context.Context) *zerolog.Logger {
	logger := zerolog.Ctx(ctx)
	if logger != nil {
		return logger
	}

	globalLogger := zerolog.Logger{}
	return &globalLogger
}
