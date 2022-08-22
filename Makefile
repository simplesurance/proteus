all: check

check:
	golangci-lint run ./...

test:
	go test --race ./...

cover:
	go test ./... -covermode=count -coverprofile=coverage.out
	go tool cover -html=coverage.out
