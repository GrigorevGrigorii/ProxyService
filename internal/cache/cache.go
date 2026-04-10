package cache

import (
	"context"
	"proxy-service/internal/models"
)

type Repository interface {
	Set(ctx context.Context, service models.ServiceDTO, target models.TargetDTO, data string, statusCode int, contentType string) error
	Get(ctx context.Context, service models.ServiceDTO, target models.TargetDTO) (string, int, string, error)
}
