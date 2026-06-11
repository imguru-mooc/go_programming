BINARY := healthcheck

.PHONY: all build run test clean

all: build

build:
	go build -o bin/$(BINARY) ./cmd/$(BINARY)

run: build
	./bin/$(BINARY) -w 5 -t 10s

test:
	go test -race ./...

clean:
	rm -rf bin/
