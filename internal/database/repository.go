package database

import (
	"context"
	"errors"
	"proxy-service/internal/models"
	"proxy-service/internal/repository"

	"gorm.io/gorm"
)

type DBRepository struct {
	DB *gorm.DB
}

func (r *DBRepository) GetAll(ctx context.Context) ([]models.ServiceDTO, error) {
	services, err := gorm.G[Service](r.DB).Preload("Targets", nil).Find(ctx)
	if err != nil {
		return nil, convertErr(err)
	}

	result := make([]models.ServiceDTO, len(services))
	for i, service := range services {
		result[i] = ServiceDTOFromDBModel(service)
	}
	return result, nil
}

func (r *DBRepository) Get(ctx context.Context, name string) (*models.ServiceDTO, error) {
	service, err := gorm.G[Service](r.DB).Preload("Targets", nil).Where("name = ?", name).First(ctx)
	if err != nil {
		return nil, convertErr(err)
	}

	result := ServiceDTOFromDBModel(service)
	return &result, nil
}

func (r *DBRepository) Resolve(ctx context.Context, name, path, method, query string) (*models.ServiceDTO, error) {
	targetsFilter := func(db gorm.PreloadBuilder) error {
		db.Where("path = ? AND method = ? AND (query = ? OR query = '*')", path, method, query)
		return nil
	}
	service, err := gorm.G[Service](r.DB).Preload("Targets", targetsFilter).Where("name = ?", name).First(ctx)
	if err != nil {
		return nil, convertErr(err)
	}
	if len(service.Targets) == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	result := ServiceDTOFromDBModel(service)
	return &result, nil
}

func (r *DBRepository) LoadCacheable(ctx context.Context) ([]models.ServiceDTO, error) {
	var result []models.ServiceDTO

	targetsFilter := func(db gorm.PreloadBuilder) error {
		db.Where("cache_interval IS NOT NULL")
		return nil
	}
	services, err := gorm.G[Service](r.DB).Preload("Targets", targetsFilter).Find(ctx)
	if err != nil {
		return nil, convertErr(err)
	}
	for _, service := range services {
		if len(service.Targets) != 0 {
			result = append(result, ServiceDTOFromDBModel(service))
		}
	}

	return result, nil
}

func (r *DBRepository) Create(ctx context.Context, service *models.ServiceDTO) error {
	serviceRow := ServiceDBModelFromDTO(*service)
	err := r.DB.Transaction(func(tx *gorm.DB) error {
		return gorm.G[Service](tx).Create(ctx, &serviceRow)
	})
	return convertErr(err)
}

func (r *DBRepository) Update(ctx context.Context, service *models.ServiceDTO) error {
	serviceRow := ServiceDBModelFromDTO(*service)
	err := r.DB.Transaction(func(tx *gorm.DB) error {
		tmpService, err := gorm.G[Service](tx).Raw("SELECT * FROM \"services\" WHERE name = ? FOR UPDATE", service.Name).First(ctx)
		if err != nil {
			return err
		}

		if tmpService.Version != service.Version {
			return repository.ErrVersionMismatch
		}

		service.Version = tmpService.Version + 1
		if _, err := gorm.G[Service](tx).Omit("Targets").Updates(ctx, serviceRow); err != nil {
			return err
		}

		if _, err := gorm.G[Target](tx).Where("service_name = ?", service.Name).Delete(ctx); err != nil {
			return err
		}

		if len(service.Targets) > 0 {
			if err := gorm.G[Target](tx).CreateInBatches(ctx, &serviceRow.Targets, 10); err != nil {
				return err
			}
		}

		return nil
	})
	return convertErr(err)
}

func (r *DBRepository) Delete(ctx context.Context, name string) error {
	_, err := gorm.G[Service](r.DB).Where("name = ?", name).Delete(ctx)
	return convertErr(err)
}

func convertErr(err error) error {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return repository.ErrNotFound
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return repository.ErrAlreadyExists
	default:
		return err
	}
}
