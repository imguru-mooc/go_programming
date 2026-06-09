BINARY     := calc
VERSION    := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS    := -ldflags="-s -w -X 'main.Version=$(VERSION)'"

.PHONY: all
all: fmt vet test build

.PHONY: build
build:
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/$(BINARY)

.PHONY: build-all
build-all:
	GOOS=linux   GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY)-linux-amd64 ./cmd/$(BINARY)
	GOOS=darwin  GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY)-darwin-arm64 ./cmd/$(BINARY)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY)-windows-amd64.exe ./cmd/$(BINARY)

.PHONY: test
test:
	go test -v -race -cover ./...

.PHONY: bench
bench:
	go test -bench=. -benchmem ./internal/calculator

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: install
install:
	go install $(LDFLAGS) ./cmd/$(BINARY)

.PHONY: run
run:
	go run ./cmd/$(BINARY) add 3 5

.PHONY: clean
clean:
	rm -rf bin/

.PHONY: help
help:
	@echo "타깃: all build build-all test fmt vet tidy install run clean"
