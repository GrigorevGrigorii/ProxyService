package database

import (
	"proxy-service/internal/models"
)

func TargetDTOFromDBModel(obj Target) models.TargetDTO {
	return models.TargetDTO{
		Path:          obj.Path,
		Method:        obj.Method,
		Query:         obj.Query,
		CacheInterval: obj.CacheInterval,
	}
}

func ServiceDTOFromDBModel(obj Service) models.ServiceDTO {
	targets := make([]models.TargetDTO, len(obj.Targets))
	for i, target := range obj.Targets {
		targets[i] = TargetDTOFromDBModel(target)
	}

	return models.ServiceDTO{
		Name:          obj.Name,
		Scheme:        obj.Scheme,
		Host:          obj.Host,
		Timeout:       obj.Timeout,
		RetryCount:    obj.RetryCount,
		RetryInterval: obj.RetryInterval,
		Version:       obj.Version,
		Targets:       targets,
	}
}

func TargetDBModelFromDTO(serviceName string, dto models.TargetDTO) Target {
	return Target{
		ServiceName:   serviceName,
		Path:          dto.Path,
		Method:        dto.Method,
		Query:         dto.Query,
		CacheInterval: dto.CacheInterval,
	}
}

func ServiceDBModelFromDTO(dto models.ServiceDTO) Service {
	targets := make([]Target, len(dto.Targets))
	for i, target := range dto.Targets {
		targets[i] = TargetDBModelFromDTO(dto.Name, target)
	}

	return Service{
		Name:          dto.Name,
		Scheme:        dto.Scheme,
		Host:          dto.Host,
		Timeout:       dto.Timeout,
		RetryCount:    dto.RetryCount,
		RetryInterval: dto.RetryInterval,
		Version:       dto.Version,
		Targets:       targets,
	}
}
