package models

import (
	"net/http"
	"testing"
)

func TestValidate(t *testing.T) {
	serviceTemplate := ServiceDTO{
		Name:    "mock",
		Scheme:  "http",
		Host:    "localhost:8080",
		Timeout: 10.0,
	}
	cacheInterval := "1m"
	incorrectCacheInterval := "incoddect"

	testCases := []struct {
		Target        TargetDTO
		ExpectedError string
	}{
		{
			Target:        TargetDTO{Path: "/", Method: http.MethodGet, Query: "*", CacheInterval: nil},
			ExpectedError: "",
		},
		{
			Target:        TargetDTO{Path: "/", Method: http.MethodGet, Query: "a=b", CacheInterval: nil},
			ExpectedError: "",
		},
		{
			Target:        TargetDTO{Path: "/", Method: http.MethodGet, Query: "*", CacheInterval: &cacheInterval},
			ExpectedError: "Key: 'ServiceDTO.Targets[0].CacheInterval' Error:Field validation for 'CacheInterval' failed on the 'must_be_nil_for_*_query' tag",
		},
		{
			Target:        TargetDTO{Path: "/", Method: http.MethodGet, Query: "a=b", CacheInterval: &incorrectCacheInterval},
			ExpectedError: "Key: 'ServiceDTO.Targets[0].CacheInterval' Error:Field validation for 'CacheInterval' failed on the 'duration' tag",
		},
		{
			Target:        TargetDTO{Path: "/", Method: http.MethodPost, Query: "a=b", CacheInterval: &cacheInterval},
			ExpectedError: "Key: 'ServiceDTO.Targets[0].CacheInterval' Error:Field validation for 'CacheInterval' failed on the 'must_be_nil_for_non_get_methods' tag",
		},
	}

	for _, tc := range testCases {
		serviceTemplate.Targets = []TargetDTO{tc.Target}

		err := Validate.Struct(serviceTemplate)
		errStr := ""
		if err != nil {
			errStr = err.Error()
		}

		if errStr != tc.ExpectedError {
			t.Fatalf("Expected error '%s', got '%s' for target %v", tc.ExpectedError, errStr, tc.Target)
		}
	}
}
