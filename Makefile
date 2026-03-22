butify:
	go fmt ./internal ./cmd
	golint -set_exit_status ./internal ./cmd
	gofmt -w -e ./internal ./cmd
	goimports -w -e ./internal ./cmd

start-proxy:
	go run cmd/proxy-api-server/main.go

start-mock:
	go run cmd/mock-api-server/main.go

test:
	go test ./...

start-containers:
	docker compose -f test/docker-compose.yaml up

start-containers-with-build:
	docker compose -f test/docker-compose.yaml up --build

stop-containers:
	docker compose -f test/docker-compose.yaml down
