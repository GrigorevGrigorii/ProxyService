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
	GetAll() ([]Service, error)
	Get(name string) (*Service, error)
	GetFiltered(name, path, method string) (*Service, error)
	Create(service *Service) error
	Update(service *Service) error
	Delete(name string) error
}

type DBRepository struct {
	DB *gorm.DB
}

func (r *DBRepository) GetAll() ([]Service, error) {
	ctx := context.Background()
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

func (r *DBRepository) Get(name string) (*Service, error) {
	ctx := context.Background()
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

func (r *DBRepository) GetFiltered(name, path, method string) (*Service, error) {
	ctx := context.Background()
	var result *Service

	err := r.DB.Transaction(func(tx *gorm.DB) error {
		service, err := gorm.G[Service](tx).Preload("Targets", func(db gorm.PreloadBuilder) error {
			db.Where("path = ? AND method = ?", path, method)
			return nil
		}).Where("name = ?", name).First(ctx)
		if err != nil {
			return err
		}
		result = &service
		return nil
	})

	return result, err
}

func (r *DBRepository) Create(service *Service) error {
	ctx := context.Background()
	return r.DB.Transaction(func(tx *gorm.DB) error {
		return gorm.G[Service](tx).Create(ctx, service)
	})
}

func (r *DBRepository) Update(service *Service) error {
	ctx := context.Background()
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

func (r *DBRepository) Delete(name string) error {
	ctx := context.Background()
	_, err := gorm.G[Service](r.DB).Where("name = ?", name).Delete(ctx)
	return err
}
