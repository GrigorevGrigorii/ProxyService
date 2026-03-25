package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"proxy-service/internal/database"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
)

type stubAdminRepository struct {
	getAllFn func(ctx context.Context) ([]database.Service, error)
	getFn    func(ctx context.Context, name string) (*database.Service, error)
	createFn func(ctx context.Context, service *database.Service) error
	updateFn func(ctx context.Context, service *database.Service) error
	deleteFn func(ctx context.Context, name string) error
}

func (s *stubAdminRepository) GetAll(ctx context.Context) ([]database.Service, error) {
	if s.getAllFn == nil {
		return nil, errors.New("unexpected GetAll call")
	}
	return s.getAllFn(ctx)
}

func (s *stubAdminRepository) Get(ctx context.Context, name string) (*database.Service, error) {
	if s.getFn == nil {
		return nil, errors.New("unexpected Get call")
	}
	return s.getFn(ctx, name)
}

func (s *stubAdminRepository) GetFiltered(ctx context.Context, name, path, method, query string) (*database.Service, error) {
	return nil, errors.New("unexpected GetFiltered call")
}

func (s *stubAdminRepository) Create(ctx context.Context, service *database.Service) error {
	if s.createFn == nil {
		return errors.New("unexpected Create call")
	}
	return s.createFn(ctx, service)
}

func (s *stubAdminRepository) Update(ctx context.Context, service *database.Service) error {
	if s.updateFn == nil {
		return errors.New("unexpected Update call")
	}
	return s.updateFn(ctx, service)
}

func (s *stubAdminRepository) Delete(ctx context.Context, name string) error {
	if s.deleteFn == nil {
		return errors.New("unexpected Delete call")
	}
	return s.deleteFn(ctx, name)
}

func setupAdminTest(repo database.Repository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := &AdminHandlers{DBRepository: repo}
	r := gin.New()
	r.GET("/service", h.GetServices)
	r.GET("/service/:name", h.GetService)
	r.POST("/service", h.CreateService)
	r.PUT("/service/:name", h.UpdateService)
	r.DELETE("/service/:name", h.DeleteService)
	return r
}

func TestGetServices(t *testing.T) {
	router := setupAdminTest(&stubAdminRepository{
		getAllFn: func(ctx context.Context) ([]database.Service, error) {
			return []database.Service{
				{
					Name:          "mock",
					Scheme:        "http",
					Host:          "localhost:8080",
					Timeout:       10.0,
					RetryCount:    3,
					RetryInterval: 0.1,
					Version:       1,
					Targets: []database.Target{
						{ServiceName: "mock", Path: "/mock", Method: "GET", Query: "query=param"},
					},
				},
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/service", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	// Check resonse body
	var got []ServiceDTO
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}
	expected := []ServiceDTO{
		{
			Name:          "mock",
			Scheme:        "http",
			Host:          "localhost:8080",
			Timeout:       10.0,
			RetryCount:    3,
			RetryInterval: 0.1,
			Version:       1,
			Targets: []TargetDTO{
				{Path: "/mock", Method: "GET", Query: "query=param"},
			},
		},
	}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("unexpected response body: %s", w.Body.String())
	}
}

func TestGetServiceNotFound(t *testing.T) {
	router := setupAdminTest(&stubAdminRepository{
		getFn: func(ctx context.Context, name string) (*database.Service, error) {
			if name != "missing" {
				t.Fatalf("expected service name %q, got %q", "missing", name)
			}
			return nil, database.ErrNotFound
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/service/missing", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestGetService(t *testing.T) {
	router := setupAdminTest(&stubAdminRepository{
		getFn: func(ctx context.Context, name string) (*database.Service, error) {
			if name != "mock" {
				t.Fatalf("expected service name %q, got %q", "mock", name)
			}
			return &database.Service{
				Name:          "mock",
				Scheme:        "http",
				Host:          "localhost:8080",
				Timeout:       1.5,
				RetryCount:    0,
				RetryInterval: 0,
				Version:       1,
				Targets: []database.Target{
					{ServiceName: "mock", Path: "/mock", Method: "GET", Query: "query=param"},
				},
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/service/mock", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	// Check resonse body
	var got ServiceDTO
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}
	expected := ServiceDTO{
		Name:          "mock",
		Scheme:        "http",
		Host:          "localhost:8080",
		Timeout:       1.5,
		RetryCount:    0,
		RetryInterval: 0,
		Version:       1,
		Targets: []TargetDTO{
			{Path: "/mock", Method: "GET", Query: "query=param"},
		},
	}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("unexpected response body: %s", w.Body.String())
	}
}

func TestCreateServiceDuplicate(t *testing.T) {
	router := setupAdminTest(&stubAdminRepository{
		createFn: func(ctx context.Context, service *database.Service) error {
			if service.Name != "mock" {
				t.Fatalf("unexpected service name: %s", service.Name)
			}
			return database.ErrAlreadyExists
		},
	})

	body := []byte(`{
		"name":"mock",
		"scheme":"http",
		"host":"localhost:8081",
		"timeout":10.0,
		"retry_count":3,
		"retry_interval":0.5
	}`)
	req := httptest.NewRequest(http.MethodPost, "/service", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, w.Code)
	}
}

func TestUpdateServiceVersionMismatch(t *testing.T) {
	router := setupAdminTest(&stubAdminRepository{
		updateFn: func(ctx context.Context, service *database.Service) error {
			if service.Name != "mock" {
				t.Fatalf("unexpected service name: %s", service.Name)
			}
			return database.ErrVersionMismatch
		},
	})

	body := []byte(`{
		"name":"mock",
		"scheme":"http",
		"host":"localhost:8081",
		"timeout":10.0,
		"retry_count":3,
		"retry_interval":0.5,
		"version":2,
		"targets":[{"path":"/new","method":"GET","query":"query=param","cache_interval":"1m"}]
	}`)
	req := httptest.NewRequest(http.MethodPut, "/service/mock", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, w.Code)
	}
}

func TestDeleteService(t *testing.T) {
	router := setupAdminTest(&stubAdminRepository{
		deleteFn: func(ctx context.Context, name string) error {
			if name != "mock" {
				t.Fatalf("expected service name %q, got %q", "mock", name)
			}
			return nil
		},
	})

	req := httptest.NewRequest(http.MethodDelete, "/service/mock", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	var got map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}
	expected := map[string]string{"message": "ok"}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("unexpected response body: %s", w.Body.String())
	}
}
