APP_NAME=redisclient
BIN_DIR=bin

.PHONY: build run test clean

build:
	go build -o $(BIN_DIR)/$(APP_NAME) ./cmd/tui

run:
	go run ./cmd/tui

test:
	go test ./...

clean:
	rm -f $(BIN_DIR)/$(APP_NAME)
