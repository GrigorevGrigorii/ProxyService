package handlers

import (
	"net/http"
	"testing"
)

var (
	// service
	name          string  = "mock"
	scheme        string  = "http"
	host          string  = "localhost"
	timeout       float32 = 10.0
	retryCount    int     = 0
	retryInterval float32 = 0
	version       int     = 0
	// target
	path                 string = "/"
	methodGet            string = http.MethodGet
	methodPost           string = http.MethodPost
	query                string = "a=b"
	anyQuery             string = "*"
	cacheInterval        string = "1m"
	invalidCacheInterval string = "invalid"
)

func TestCreateServiceRequest(t *testing.T) {
	testCases := []struct {
		request     createServiceRequest
		expectedErr string
	}{
		{ // incorrect with empty fields
			request: createServiceRequest{},
			expectedErr: `Key: 'createServiceRequest.Name' Error:Field validation for 'Name' failed on the 'required' tag
Key: 'createServiceRequest.Scheme' Error:Field validation for 'Scheme' failed on the 'required' tag
Key: 'createServiceRequest.Host' Error:Field validation for 'Host' failed on the 'required' tag
Key: 'createServiceRequest.Timeout' Error:Field validation for 'Timeout' failed on the 'required' tag
Key: 'createServiceRequest.RetryCount' Error:Field validation for 'RetryCount' failed on the 'required' tag
Key: 'createServiceRequest.RetryInterval' Error:Field validation for 'RetryInterval' failed on the 'required' tag
Key: 'createServiceRequest.Targets' Error:Field validation for 'Targets' failed on the 'required' tag`,
		},
		{ // correct with empty targets
			request: createServiceRequest{
				Name:          &name,
				Scheme:        &scheme,
				Host:          &host,
				Timeout:       &timeout,
				RetryCount:    &retryCount,
				RetryInterval: &retryInterval,
				Targets:       []target{},
			},
			expectedErr: "",
		},
		{ // correct with empty cache interval
			request: createServiceRequest{
				Name:          &name,
				Scheme:        &scheme,
				Host:          &host,
				Timeout:       &timeout,
				RetryCount:    &retryCount,
				RetryInterval: &retryInterval,
				Targets:       []target{{Path: &path, Method: &methodGet, Query: &query}},
			},
			expectedErr: "",
		},
		{ // correct with valid cache interval
			request: createServiceRequest{
				Name:          &name,
				Scheme:        &scheme,
				Host:          &host,
				Timeout:       &timeout,
				RetryCount:    &retryCount,
				RetryInterval: &retryInterval,
				Targets:       []target{{Path: &path, Method: &methodGet, Query: &query, CacheInterval: &cacheInterval}},
			},
			expectedErr: "",
		},
		{ // incorrect with invalid cache interval
			request: createServiceRequest{
				Name:          &name,
				Scheme:        &scheme,
				Host:          &host,
				Timeout:       &timeout,
				RetryCount:    &retryCount,
				RetryInterval: &retryInterval,
				Targets:       []target{{Path: &path, Method: &methodGet, Query: &query, CacheInterval: &invalidCacheInterval}},
			},
			expectedErr: "Key: 'createServiceRequest.Targets[0].CacheInterval' Error:Field validation for 'CacheInterval' failed on the 'duration' tag",
		},
		{ // incorrect: caching for '*' query
			request: createServiceRequest{
				Name:          &name,
				Scheme:        &scheme,
				Host:          &host,
				Timeout:       &timeout,
				RetryCount:    &retryCount,
				RetryInterval: &retryInterval,
				Targets:       []target{{Path: &path, Method: &methodGet, Query: &anyQuery, CacheInterval: &cacheInterval}},
			},
			expectedErr: "Key: 'createServiceRequest.Targets[0].CacheInterval' Error:Field validation for 'CacheInterval' failed on the 'must_be_nil_for_*_query' tag",
		},
		{ // incorrect: caching for non-GET request
			request: createServiceRequest{
				Name:          &name,
				Scheme:        &scheme,
				Host:          &host,
				Timeout:       &timeout,
				RetryCount:    &retryCount,
				RetryInterval: &retryInterval,
				Targets:       []target{{Path: &path, Method: &methodPost, Query: &query, CacheInterval: &cacheInterval}},
			},
			expectedErr: "Key: 'createServiceRequest.Targets[0].CacheInterval' Error:Field validation for 'CacheInterval' failed on the 'must_be_nil_for_non_get_methods' tag",
		},
	}

	for i, tc := range testCases {
		err := Validate.Struct(tc.request)
		errStr := ""
		if err != nil {
			errStr = err.Error()
		}

		if errStr != tc.expectedErr {
			t.Fatalf("Expected error '%s', got '%s' for test case #%d", tc.expectedErr, errStr, i)
		}
	}
}

func TestUpdateServiceRequest(t *testing.T) {
	testCases := []struct {
		request     updateServiceRequest
		expectedErr string
	}{
		{ // incorrect with empty fields
			request: updateServiceRequest{},
			expectedErr: `Key: 'updateServiceRequest.Scheme' Error:Field validation for 'Scheme' failed on the 'required' tag
Key: 'updateServiceRequest.Host' Error:Field validation for 'Host' failed on the 'required' tag
Key: 'updateServiceRequest.Timeout' Error:Field validation for 'Timeout' failed on the 'required' tag
Key: 'updateServiceRequest.RetryCount' Error:Field validation for 'RetryCount' failed on the 'required' tag
Key: 'updateServiceRequest.RetryInterval' Error:Field validation for 'RetryInterval' failed on the 'required' tag
Key: 'updateServiceRequest.Version' Error:Field validation for 'Version' failed on the 'required' tag
Key: 'updateServiceRequest.Targets' Error:Field validation for 'Targets' failed on the 'required' tag`,
		},
		{ // correct with empty targets
			request: updateServiceRequest{
				Scheme:        &scheme,
				Host:          &host,
				Timeout:       &timeout,
				RetryCount:    &retryCount,
				RetryInterval: &retryInterval,
				Version:       &version,
				Targets:       []target{},
			},
			expectedErr: "",
		},
		{ // correct with empty cache interval
			request: updateServiceRequest{
				Scheme:        &scheme,
				Host:          &host,
				Timeout:       &timeout,
				RetryCount:    &retryCount,
				RetryInterval: &retryInterval,
				Version:       &version,
				Targets:       []target{{Path: &path, Method: &methodGet, Query: &query}},
			},
			expectedErr: "",
		},
		{ // correct with valid cache interval
			request: updateServiceRequest{
				Scheme:        &scheme,
				Host:          &host,
				Timeout:       &timeout,
				RetryCount:    &retryCount,
				RetryInterval: &retryInterval,
				Version:       &version,
				Targets:       []target{{Path: &path, Method: &methodGet, Query: &query, CacheInterval: &cacheInterval}},
			},
			expectedErr: "",
		},
		{ // incorrect with invalid cache interval
			request: updateServiceRequest{
				Scheme:        &scheme,
				Host:          &host,
				Timeout:       &timeout,
				RetryCount:    &retryCount,
				RetryInterval: &retryInterval,
				Version:       &version,
				Targets:       []target{{Path: &path, Method: &methodGet, Query: &query, CacheInterval: &invalidCacheInterval}},
			},
			expectedErr: "Key: 'updateServiceRequest.Targets[0].CacheInterval' Error:Field validation for 'CacheInterval' failed on the 'duration' tag",
		},
		{ // incorrect: caching for '*' query
			request: updateServiceRequest{
				Scheme:        &scheme,
				Host:          &host,
				Timeout:       &timeout,
				RetryCount:    &retryCount,
				RetryInterval: &retryInterval,
				Version:       &version,
				Targets:       []target{{Path: &path, Method: &methodGet, Query: &anyQuery, CacheInterval: &cacheInterval}},
			},
			expectedErr: "Key: 'updateServiceRequest.Targets[0].CacheInterval' Error:Field validation for 'CacheInterval' failed on the 'must_be_nil_for_*_query' tag",
		},
		{ // incorrect: caching for non-GET request
			request: updateServiceRequest{
				Scheme:        &scheme,
				Host:          &host,
				Timeout:       &timeout,
				RetryCount:    &retryCount,
				RetryInterval: &retryInterval,
				Version:       &version,
				Targets:       []target{{Path: &path, Method: &methodPost, Query: &query, CacheInterval: &cacheInterval}},
			},
			expectedErr: "Key: 'updateServiceRequest.Targets[0].CacheInterval' Error:Field validation for 'CacheInterval' failed on the 'must_be_nil_for_non_get_methods' tag",
		},
	}

	for i, tc := range testCases {
		err := Validate.Struct(tc.request)
		errStr := ""
		if err != nil {
			errStr = err.Error()
		}

		if errStr != tc.expectedErr {
			t.Fatalf("Expected error '%s', got '%s' for test case #%d", tc.expectedErr, errStr, i)
		}
	}
}
