package models

import (
	"proxy-service/internal/database"
)

func TargetDTOFromDBModel(obj database.Target) TargetDTO {
	return TargetDTO{
		Path:          obj.Path,
		Method:        obj.Method,
		Query:         obj.Query,
		CacheInterval: obj.CacheInterval,
	}
}

func ServiceDTOFromDBModel(obj database.Service) ServiceDTO {
	targets := make([]TargetDTO, len(obj.Targets))
	for i, target := range obj.Targets {
		targets[i] = TargetDTOFromDBModel(target)
	}

	return ServiceDTO{
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

func TargetDBModelFromDTO(serviceName string, dto TargetDTO) database.Target {
	return database.Target{
		ServiceName:   serviceName,
		Path:          dto.Path,
		Method:        dto.Method,
		Query:         dto.Query,
		CacheInterval: dto.CacheInterval,
	}
}

func ServiceDBModelFromDTO(dto ServiceDTO) database.Service {
	targets := make([]database.Target, len(dto.Targets))
	for i, target := range dto.Targets {
		targets[i] = TargetDBModelFromDTO(dto.Name, target)
	}

	return database.Service{
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
