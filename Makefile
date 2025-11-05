APP_NAME=redisclient
BIN_DIR=bin
LOG_FILE=/tmp/redisclient.log
n ?= 1000

.PHONY: build run test clean fake-data

run:
	go run ./cmd/tui

build:
	go build -o $(BIN_DIR)/$(APP_NAME) ./cmd/tui

test:
	go test ./...

clean:
	rm -f $(BIN_DIR)/$(APP_NAME)
	
fake-data: # Generate fake data for testing
	go run ./script/data.go -n $(n) -b $(big)

log: # View application logs
	less $(LOG_FILE)

release: # Build & upload binary to GitHub Releases
	goreleaser release --clean

