package models

import (
	"net/http"
	"testing"
)

func TestValidate(t *testing.T) {
	serviceTemplate := ServiceDTO{}
	cacheInterval := "1m"
	incorrectCacheInterval := "incoddect"

	testCases := []struct {
		Target        TargetDTO
		ExpectedError string
		ExpectedQuery string
	}{
		{
			Target:        TargetDTO{Path: "/", Method: http.MethodGet, Query: "*", CacheInterval: nil},
			ExpectedError: "",
			ExpectedQuery: "*",
		},
		{
			Target:        TargetDTO{Path: "/", Method: http.MethodGet, Query: "a=b", CacheInterval: nil},
			ExpectedError: "",
			ExpectedQuery: "a=b",
		},
		{
			Target:        TargetDTO{Path: "/", Method: http.MethodGet, Query: "*", CacheInterval: &cacheInterval},
			ExpectedError: "Not allowed to set cache interval for target with query='*'",
			ExpectedQuery: "*",
		},
		{
			Target:        TargetDTO{Path: "/", Method: http.MethodGet, Query: "a=b", CacheInterval: &incorrectCacheInterval},
			ExpectedError: "time: invalid duration \"incoddect\"",
			ExpectedQuery: "a=b",
		},
		{
			Target:        TargetDTO{Path: "/", Method: http.MethodPost, Query: "a=b", CacheInterval: &cacheInterval},
			ExpectedError: "'cache_interval' can be set only for GET requests",
			ExpectedQuery: "a=b",
		},
		{
			Target:        TargetDTO{Path: "/", Method: http.MethodGet, Query: "b=c&a=b", CacheInterval: nil},
			ExpectedError: "",
			ExpectedQuery: "a=b&b=c",
		},
	}

	for _, tc := range testCases {
		serviceTemplate.Targets = []TargetDTO{tc.Target}

		err := serviceTemplate.Validate()
		errStr := ""
		if err != nil {
			errStr = err.Error()
		}

		if errStr != tc.ExpectedError {
			t.Fatalf("Expected error '%s', got '%s' for target %v", tc.ExpectedError, errStr, tc.Target)
		}
		if serviceTemplate.Targets[0].Query != tc.ExpectedQuery {
			t.Fatalf("Expected query '%s', got '%s' for target %v", tc.ExpectedQuery, serviceTemplate.Targets[0].Query, tc.Target)
		}
	}
}
