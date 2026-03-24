package database

import (
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
	var services []Service
	err := r.DB.Preload("Targets").Find(&services).Error
	if err != nil {
		return nil, err
	}

	return services, nil
}

func (r *DBRepository) Get(name string) (*Service, error) {
	var service Service

	err := r.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Preload("Targets").First(&service, "name = ?", name).Error
	})
	if err != nil {
		return nil, err
	}

	return &service, nil
}

func (r *DBRepository) GetFiltered(name, path, method string) (*Service, error) {
	var service Service

	err := r.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Preload("Targets", "path = ? AND method = ?", path, method).First(&service, "name = ?", name).Error
	})
	if err != nil {
		return nil, err
	}

	return &service, nil
}

func (r *DBRepository) Create(service *Service) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Create(service).Error
	})
}

func (r *DBRepository) Update(service *Service) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		var tmpService Service
		if err := tx.First(&tmpService, "name = ?", service.Name).Error; err != nil {
			return err
		}

		if tmpService.Version != service.Version {
			return ErrVersionMismatch
		}

		service.Version = tmpService.Version + 1
		if err := tx.Omit("Targets").Save(service).Error; err != nil {
			return err
		}

		if err := tx.Where("service_name = ?", service.Name).Delete(&Target{}).Error; err != nil {
			return err
		}

		if len(service.Targets) > 0 {
			for i := range service.Targets {
				service.Targets[i].ServiceName = service.Name
			}

			if err := tx.Create(&service.Targets).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *DBRepository) Delete(name string) error {
	result := r.DB.Delete(&Service{Name: name})
	return result.Error
}
