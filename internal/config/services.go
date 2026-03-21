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

func LoadServices(path string) ([]Service, error) {
	var services []Service

	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return services, err
	}

	err = yaml.Unmarshal(yamlFile, &services)
	if err != nil {
		return services, err
	}

	return services, nil
}
