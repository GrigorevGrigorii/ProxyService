package handlers

import (
	"proxy-service/internal/database"
)

type TargetDTO struct {
	Path   string `json:"path"`
	Method string `json:"method"`
}

type ServiceDTO struct {
	Name          string      `json:"name"`
	Scheme        string      `json:"scheme"`
	Host          string      `json:"host"`
	Timeout       float32     `json:"timeout"`
	RetryCount    int         `json:"retry_count"`
	RetryInterval float32     `json:"retry_interval"`
	Version       int         `json:"version"`
	Targets       []TargetDTO `json:"targets"`
}

func targetDTOFromDBModel(obj database.Target) TargetDTO {
	return TargetDTO{
		Path:   obj.Path,
		Method: obj.Method,
	}
}

func serviceDTOFromDBModel(obj database.Service) ServiceDTO {
	targets := make([]TargetDTO, len(obj.Targets))
	for i, target := range obj.Targets {
		targets[i] = targetDTOFromDBModel(target)
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

func targetDBModelFromDTO(serviceName string, dto TargetDTO) database.Target {
	return database.Target{
		ServiceName: serviceName,
		Path:        dto.Path,
		Method:      dto.Method,
	}
}

func serviceDBModelFromDTO(dto ServiceDTO) database.Service {
	targets := make([]database.Target, len(dto.Targets))
	for i, target := range dto.Targets {
		targets[i] = targetDBModelFromDTO(dto.Name, target)
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
