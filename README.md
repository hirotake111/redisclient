# redisclient

This repository is a Redis client program that provides both a Terminal User Interface (TUI) and a Web User Interface (Web UI) for managing and interacting with Redis servers.

## Specifying Redis Connection Parameters

This application connects to a Redis server using connection parameters specified via environment variables:

- `REDIS_URL`: The URL or address of the Redis server. If not set, defaults to `redis://localhost:6379`.
- `REDIS_PASSWORD`: The password for the Redis server, if required. If not set, no password is used.

You can set these environment variables before running the application to connect to a different Redis server or to use authentication.

### Example usage

```sh
# Connect to a local Redis server (default)
go run ./cmd/tui

# Connect to a remote Redis server with a password
export REDIS_URL=redis://your-redis-host:6379
export REDIS_PASSWORD=yourpassword
go run ./cmd/tui
```
