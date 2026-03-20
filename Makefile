butify-code:
	go fmt .
	golint -set_exit_status .
	gofmt -w -e .
	goimports -w -e .
