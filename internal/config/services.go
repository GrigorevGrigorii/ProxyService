package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Target struct {
	Path   string `yaml:"path"`
	Method string `yaml:"method"`
}

type Service struct {
	Name          string              `yaml:"name"`
	Scheme        string              `yaml:"scheme"`
	Host          string              `yaml:"host"`
	Targets       []Target            `yaml:"targets"`
	Timeout       float32             `yaml:"timeout"`
	RetryCount    uint8               `yaml:"retry_count"`
	RetryInterval float32             `yaml:"retry_interval"`
	TargetsSet    map[Target]struct{} `yaml:"-"`
}

func (s *Service) FillTargetsSet() {
	s.TargetsSet = make(map[Target]struct{}, len(s.Targets))
	for i := range s.Targets {
		s.TargetsSet[s.Targets[i]] = struct{}{}
	}
}

func LoadServices(path string) (*map[string]Service, error) {
	var services []Service
	var servicesMap map[string]Service

	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(yamlFile, &services)
	if err != nil {
		return nil, err
	}

	servicesMap = make(map[string]Service, len(services))
	for _, service := range services {
		service.FillTargetsSet()
		servicesMap[service.Name] = service
	}

	return &servicesMap, nil
}
