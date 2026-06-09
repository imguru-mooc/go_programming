APP_NAME = myapp
MAIN_FILE = main.go

.PHONY : run build clean test fmt tidy

all: build

run:
	go run $(MAIN_FILE) 

build:
	go build -o $(APP_NAME)  $(MAIN_FILE)

clean:
	rm -rf $(APP_NAME)

test:
	go test ./...

fmt:
	go fmt ./...

tidy: 
	go mod tidy
