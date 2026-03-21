butify:
	go fmt .
	golint -set_exit_status .
	gofmt -w -e .
	goimports -w -e .

start-proxy:
	go run cmd/proxy-api-server/main.go
