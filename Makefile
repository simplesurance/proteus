all: check cover

check:
	staticcheck ./...

cover:
	go test -v ./... -covermode=count -coverprofile=coverage.out
	#go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
