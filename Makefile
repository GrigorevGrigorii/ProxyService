.PHONY: test butify start-proxy start-mock start-containers start-containers-with-build stop-containers migrate-local

butify:
	go fmt ./internal ./cmd ./scripts
	golint -set_exit_status ./internal ./cmd ./scripts
	gofmt -w -e ./internal ./cmd ./scripts
	goimports -w -e ./internal ./cmd ./scripts

start-proxy:
	go run cmd/proxy-api-server/main.go

start-mock:
	go run cmd/mock-api-server/main.go

start-admin:
	go run cmd/admin-api-server/main.go

test:
	go test ./...

start-containers:
	docker compose -f test/docker-compose.yaml up

start-containers-with-build:
	docker compose -f test/docker-compose.yaml up --build

stop-containers:
	docker compose -f test/docker-compose.yaml down

migrate-local:
	migrate -source file:internal/database/migrations -database 'postgresql://proxy_service_user:proxy_service_password@127.0.0.1:5432/proxy_service_db?sslmode=disable' up
