all: check

check:
	staticcheck ./...
	go test --race ./...

cover:
	go test ./... -covermode=count -coverprofile=coverage.out
	go tool cover -html=coverage.out
