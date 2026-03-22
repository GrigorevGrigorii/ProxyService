package config

import (
	"os"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

const tmpTestServicesFileName string = "tmp_test_services.yaml"

func TestLoadServices(t *testing.T) {
	service := Service{
		Name: "mock",
		Targets: []Target{
			{
				Path:   "/mock",
				Method: "GET",
			},
		},
	}
	service.FillTargetsSet()

	var services = []Service{service}
	var servicesMap = map[string]Service{"mock": service}

	yamlData, err := yaml.Marshal(&services)
	if err != nil {
		t.Error("Failed to dump services to yaml format string")
		return
	}

	f, err := os.Create(tmpTestServicesFileName)
	if err != nil {
		t.Errorf("Failed to create %s", tmpTestServicesFileName)
		return
	}
	defer f.Close()
	defer os.Remove(tmpTestServicesFileName)

	_, err = f.Write(yamlData)
	if err != nil {
		t.Errorf("Failed to write data to %s", tmpTestServicesFileName)
		return
	}

	servicesMapGot, err := LoadServices(tmpTestServicesFileName)
	if err != nil {
		t.Error("Failed to load services")
		return
	}

	if !reflect.DeepEqual(servicesMapGot, servicesMap) {
		t.Errorf("servicesMapGot != servicesMap: %v != %v", servicesMapGot, servicesMap)
	}
}
