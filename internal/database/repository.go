package database

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

var (
	ErrVersionMismatch = errors.New("Version mismatch")
	ErrNotFound        = gorm.ErrRecordNotFound
	ErrAlreadyExists   = gorm.ErrDuplicatedKey
)

type Repository interface {
	GetAll(ctx context.Context) ([]Service, error)
	Get(ctx context.Context, name string) (*Service, error)
	GetFiltered(ctx context.Context, name, path, method, query string) (*Service, error)
	Create(ctx context.Context, service *Service) error
	Update(ctx context.Context, service *Service) error
	Delete(ctx context.Context, name string) error
}

type DBRepository struct {
	DB *gorm.DB
}

func (r *DBRepository) GetAll(ctx context.Context) ([]Service, error) {
	var result []Service

	err := r.DB.Transaction(func(tx *gorm.DB) error {
		services, err := gorm.G[Service](tx).Preload("Targets", nil).Find(ctx)
		if err != nil {
			return err
		}
		result = services
		return nil
	})

	return result, err
}

func (r *DBRepository) Get(ctx context.Context, name string) (*Service, error) {
	var result *Service

	err := r.DB.Transaction(func(tx *gorm.DB) error {
		service, err := gorm.G[Service](tx).Preload("Targets", nil).Where("name = ?", name).First(ctx)
		if err != nil {
			return err
		}
		result = &service
		return nil
	})

	return result, err
}

func (r *DBRepository) GetFiltered(ctx context.Context, name, path, method, query string) (*Service, error) {
	var result *Service

	err := r.DB.Transaction(func(tx *gorm.DB) error {
		service, err := gorm.G[Service](tx).Preload("Targets", func(db gorm.PreloadBuilder) error {
			db.Where("path = ? AND method = ? AND query = ?", path, method, query)
			return nil
		}).Where("name = ?", name).First(ctx)
		if err != nil {
			return err
		}
		if len(service.Targets) == 0 {
			return gorm.ErrRecordNotFound
		}

		result = &service
		return nil
	})

	return result, err
}

func (r *DBRepository) GetForCaching(ctx context.Context) ([]Service, error) {
	var result []Service

	err := r.DB.Transaction(func(tx *gorm.DB) error {
		services, err := gorm.G[Service](tx).Preload("Targets", func(db gorm.PreloadBuilder) error {
			db.Where("cache_interval IS NOT NULL")
			return nil
		}).Find(ctx)
		if err != nil {
			return err
		}
		for _, service := range services {
			if len(service.Targets) != 0 {
				result = append(result, service)
			}
		}
		return nil
	})

	return result, err
}

func (r *DBRepository) Create(ctx context.Context, service *Service) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		return gorm.G[Service](tx).Create(ctx, service)
	})
}

func (r *DBRepository) Update(ctx context.Context, service *Service) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		tmpService, err := gorm.G[Service](tx).Where("name = ?", service.Name).First(ctx)
		if err != nil {
			return err
		}

		if tmpService.Version != service.Version {
			return ErrVersionMismatch
		}

		service.Version = tmpService.Version + 1
		if _, err := gorm.G[Service](tx).Omit("Targets").Updates(ctx, *service); err != nil {
			return err
		}

		if _, err := gorm.G[Target](tx).Where("service_name = ?", service.Name).Delete(ctx); err != nil {
			return err
		}

		if len(service.Targets) > 0 {
			if err := gorm.G[Target](tx).CreateInBatches(ctx, &service.Targets, 10); err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *DBRepository) Delete(ctx context.Context, name string) error {
	_, err := gorm.G[Service](r.DB).Where("name = ?", name).Delete(ctx)
	return err
}
