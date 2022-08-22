all: check

.PHONY: check
check:
	golangci-lint run ./...

.PHONY: test
test:
	go test --race ./...

.PHONY: cover
cover:
	go test ./... -covermode=count -coverprofile=coverage.out
	go tool cover -html=coverage.out
