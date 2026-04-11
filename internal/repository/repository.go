package repository

import (
	"context"
	"proxy-service/internal/models"
)

type ServiceRepository interface {
	GetAll(ctx context.Context) ([]models.ServiceDTO, error)
	Get(ctx context.Context, name string) (models.ServiceDTO, error)
	Create(ctx context.Context, service models.ServiceDTO) error
	Update(ctx context.Context, service models.ServiceDTO) error
	Delete(ctx context.Context, name string) error
}

type ServiceResolver interface {
	Resolve(ctx context.Context, name, path, method, query string) (models.ServiceDTO, error)
}

type CacheableServicesLoader interface {
	LoadCacheable(ctx context.Context) ([]models.ServiceDTO, error)
}
