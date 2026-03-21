package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type HTTPMethod string

const (
	MethodGet    HTTPMethod = "GET"
	MethodPost   HTTPMethod = "POST"
	MethodPut    HTTPMethod = "PUT"
	MethodDelete HTTPMethod = "DELETE"
)

type Target struct {
	Path   string     `yaml:"path"`
	Method HTTPMethod `yaml:"method"`
}

type Service struct {
	Name          string   `yaml:"name"`
	Host          string   `yaml:"host"`
	Targets       []Target `yaml:"targets"`
	RetryCount    uint8    `yaml:"retry_count"`
	RetryInterval float32  `yaml:"retry_interval"`
}

func LoadServices(path string) (map[string]Service, error) {
	var services []Service
	var servicesMap map[string]Service

	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return servicesMap, err
	}

	err = yaml.Unmarshal(yamlFile, &services)
	if err != nil {
		return servicesMap, err
	}

	servicesMap = make(map[string]Service, len(services))
	for _, service := range services {
		servicesMap[service.Name] = service
	}

	return servicesMap, nil
}
