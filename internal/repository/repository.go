package repository

import (
	"context"
	"proxy-service/internal/models"
)

type ServiceRepository interface {
	GetAll(ctx context.Context) ([]models.ServiceDTO, error)
	Get(ctx context.Context, name string) (*models.ServiceDTO, error)
	GetFiltered(ctx context.Context, name, path, method, query string) (*models.ServiceDTO, error)
	GetForCaching(ctx context.Context) ([]models.ServiceDTO, error)
	Create(ctx context.Context, service *models.ServiceDTO) error
	Update(ctx context.Context, service *models.ServiceDTO) error
	Delete(ctx context.Context, name string) error
}
