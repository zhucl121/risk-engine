.PHONY: all setup test lint bench proto build docker clean

BINARY     := riskengine
CMD_DIR    := ./cmd/server
BUILD_DIR  := ./bin
DOCKER_IMG := ghcr.io/yourorg/riskengine

all: lint test build

## setup: install development tools
setup:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/vektra/mockery/v2@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go mod download

## test: run unit tests with race detector
test:
	go test -race -count=1 -timeout=120s ./...

## test-integration: run integration tests (requires running Redis + Kafka)
test-integration:
	go test -race -count=1 -timeout=300s -tags=integration ./...

## lint: run golangci-lint
lint:
	golangci-lint run ./...

## bench: run benchmarks
bench:
	go test -bench=. -benchmem -benchtime=5s ./internal/... ./pkg/...

## proto: regenerate protobuf
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
	       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	       api/grpc/proto/*.proto

## build: build binary
build:
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=$(shell git describe --tags --always)" \
	  -o $(BUILD_DIR)/$(BINARY) $(CMD_DIR)

## docker: build Docker image
docker:
	docker build -t $(DOCKER_IMG):$(shell git describe --tags --always) .

## clean: remove build artifacts
clean:
	rm -rf $(BUILD_DIR)

## mock: regenerate mocks
mock:
	mockery --all --dir internal --output internal/mocks --case snake

## cover: generate coverage report
cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
