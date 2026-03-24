package handlers

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"proxy-service/internal/database"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupAdminTest(t *testing.T) (sqlmock.Sqlmock, *gin.Engine) {
	t.Helper()

	gin.SetMode(gin.TestMode)

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}

	h := &AdminHandlers{DBRepository: &database.DBRepository{DB: db}}
	r := gin.New()
	r.GET("/service", h.GetServices)
	r.GET("/service/:name", h.GetService)
	r.POST("/service", h.CreateService)
	r.PUT("/service/:name", h.UpdateService)
	r.DELETE("/service/:name", h.DeleteService)

	return mock, r
}

func TestGetServices(t *testing.T) {
	mock, router := setupAdminTest(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "services"`)).
		WillReturnRows(sqlmock.NewRows([]string{
			"name", "scheme", "host", "timeout", "retry_count", "retry_interval", "version",
		}).AddRow("mock", "http", "localhost:8080", 10.0, 3, 0.1, 1))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "targets" WHERE "targets"."service_name" = $1`)).
		WithArgs("mock").
		WillReturnRows(sqlmock.NewRows([]string{"service_name", "path", "method"}).AddRow("mock", "/mock", "GET"))

	req := httptest.NewRequest(http.MethodGet, "/service", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
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
				{Path: "/mock", Method: "GET"},
			},
		},
	}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("unexpected response body: %s", w.Body.String())
	}
}

func TestGetServiceNotFound(t *testing.T) {
	mock, router := setupAdminTest(t)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "services" WHERE name = $1 ORDER BY "services"."name" LIMIT $2`)).
		WithArgs("missing", 1).
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectRollback()

	req := httptest.NewRequest(http.MethodGet, "/service/missing", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestGetService(t *testing.T) {
	mock, router := setupAdminTest(t)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "services" WHERE name = $1 ORDER BY "services"."name" LIMIT $2`)).
		WithArgs("mock", 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"name", "scheme", "host", "timeout", "retry_count", "retry_interval", "version",
		}).AddRow("mock", "http", "localhost:8080", 1.5, 0, 0, 1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "targets" WHERE "targets"."service_name" = $1`)).
		WithArgs("mock").
		WillReturnRows(sqlmock.NewRows([]string{"service_name", "path", "method"}).AddRow("mock", "/mock", "GET"))
	mock.ExpectCommit()

	req := httptest.NewRequest(http.MethodGet, "/service/mock", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
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
			{Path: "/mock", Method: "GET"},
		},
	}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("unexpected response body: %s", w.Body.String())
	}
}

func TestCreateServiceDuplicate(t *testing.T) {
	mock, router := setupAdminTest(t)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "services" ("name","scheme","host","timeout","retry_count","retry_interval","version") VALUES ($1,$2,$3,$4,$5,$6,$7)`)).
		WithArgs("mock", "http", "localhost:8081", 10.0, 3, 0.5, 0).
		WillReturnError(gorm.ErrDuplicatedKey)
	mock.ExpectRollback()

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
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestUpdateServiceVersionMismatch(t *testing.T) {
	mock, router := setupAdminTest(t)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "services" WHERE name = $1 ORDER BY "services"."name" LIMIT $2`)).
		WithArgs("mock", 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"name", "scheme", "host", "timeout", "retry_count", "retry_interval", "version",
		}).AddRow("mock", "http", "localhost:8081", 10.0, 2, 0.5, 3))
	mock.ExpectRollback()

	body := []byte(`{
		"name":"mock",
		"scheme":"http",
		"host":"localhost:8081",
		"timeout":10.0,
		"retry_count":3,
		"retry_interval":0.5,
		"version":2,
		"targets":[{"path":"/new","method":"GET"}]
	}`)
	req := httptest.NewRequest(http.MethodPut, "/service/mock", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, w.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestDeleteService(t *testing.T) {
	mock, router := setupAdminTest(t)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "services" WHERE "services"."name" = $1`)).
		WithArgs("mock").
		WillReturnResult(driver.RowsAffected(1))
	mock.ExpectCommit()

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
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
