package database

import (
	"context"
	"database/sql/driver"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupRepositoryTest(t *testing.T) (*DBRepository, sqlmock.Sqlmock) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}

	return &DBRepository{DB: db}, mock
}

func TestDBRepositoryGetAll(t *testing.T) {
	repo, mock := setupRepositoryTest(t)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "services"`)).
		WillReturnRows(sqlmock.NewRows([]string{
			"name", "scheme", "host", "timeout", "retry_count", "retry_interval", "version",
		}).AddRow("mock", "http", "localhost:8080", 10.0, 3, 0.1, 1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "targets" WHERE "targets"."service_name" = $1`)).
		WithArgs("mock").
		WillReturnRows(sqlmock.NewRows([]string{"service_name", "path", "method"}).AddRow("mock", "/mock", "GET"))
	mock.ExpectCommit()

	services, err := repo.GetAll(context.Background())
	if err != nil {
		t.Fatalf("GetAll returned error: %v", err)
	}
	if len(services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(services))
	}
	if services[0].Name != "mock" {
		t.Fatalf("expected service name mock, got %s", services[0].Name)
	}
	if len(services[0].Targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(services[0].Targets))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestDBRepositoryGetNotFound(t *testing.T) {
	repo, mock := setupRepositoryTest(t)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "services" WHERE name = $1 ORDER BY "services"."name" LIMIT $2`)).
		WithArgs("missing", 1).
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectRollback()

	service, err := repo.Get(context.Background(), "missing")
	if service != nil {
		t.Fatalf("expected nil service, got %+v", service)
	}
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestDBRepositoryGetFiltered(t *testing.T) {
	repo, mock := setupRepositoryTest(t)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "services" WHERE name = $1 ORDER BY "services"."name" LIMIT $2`)).
		WithArgs("mock", 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"name", "scheme", "host", "timeout", "retry_count", "retry_interval", "version",
		}).AddRow("mock", "http", "localhost:8080", 1.5, 0, 0, 1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "targets" WHERE "targets"."service_name" = $1 AND (path = $2 AND method = $3 AND query = $4)`)).
		WithArgs("mock", "/mock", "GET", "query=param").
		WillReturnRows(sqlmock.NewRows([]string{"service_name", "path", "method", "query"}).AddRow("mock", "/mock", "GET", "query=param"))
	mock.ExpectCommit()

	service, err := repo.GetFiltered(context.Background(), "mock", "/mock", "GET", "query=param")
	if err != nil {
		t.Fatalf("GetFiltered returned error: %v", err)
	}
	if service.Name != "mock" {
		t.Fatalf("expected service name mock, got %s", service.Name)
	}
	if len(service.Targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(service.Targets))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestDBRepositoryCreateDuplicate(t *testing.T) {
	repo, mock := setupRepositoryTest(t)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "services" ("name","scheme","host","timeout","retry_count","retry_interval","version") VALUES ($1,$2,$3,$4,$5,$6,$7)`)).
		WithArgs("mock", "http", "localhost:8081", 10.0, 3, 0.5, 0).
		WillReturnError(gorm.ErrDuplicatedKey)
	mock.ExpectRollback()

	err := repo.Create(context.Background(), &Service{
		Name:          "mock",
		Scheme:        "http",
		Host:          "localhost:8081",
		Timeout:       10.0,
		RetryCount:    3,
		RetryInterval: 0.5,
	})
	if !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestDBRepositoryUpdateVersionMismatch(t *testing.T) {
	repo, mock := setupRepositoryTest(t)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "services" WHERE name = $1 ORDER BY "services"."name" LIMIT $2`)).
		WithArgs("mock", 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"name", "scheme", "host", "timeout", "retry_count", "retry_interval", "version",
		}).AddRow("mock", "http", "localhost:8081", 10.0, 2, 0.5, 3))
	mock.ExpectRollback()

	err := repo.Update(context.Background(), &Service{
		Name:          "mock",
		Scheme:        "http",
		Host:          "localhost:8081",
		Timeout:       10.0,
		RetryCount:    3,
		RetryInterval: 0.5,
		Version:       2,
		Targets:       []Target{{Path: "/new", Method: "GET"}},
	})
	if !errors.Is(err, ErrVersionMismatch) {
		t.Fatalf("expected ErrVersionMismatch, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestDBRepositoryDelete(t *testing.T) {
	repo, mock := setupRepositoryTest(t)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "services" WHERE name = $1`)).
		WithArgs("mock").
		WillReturnResult(driver.RowsAffected(1))
	mock.ExpectCommit()

	if err := repo.Delete(context.Background(), "mock"); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
