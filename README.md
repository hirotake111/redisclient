# redisclient

This repository is a Redis client program that provides a Terminal User Interface (TUI) for managing and interacting with Redis servers.

## Specifying Redis Connection Parameters

This application connects to a Redis server using connection parameters specified via environment variables:

- `REDIS_URL`: The URL or address of the Redis server. If not set, defaults to `redis://localhost:6379`.

You can set this environment variable before running the application to connect to a different Redis server.

### Example usage

```sh
# Connect to a local Redis server (default)
go run ./cmd/tui

# Connect to a remote Redis server
export REDIS_URL=redis://your-redis-host:6379
go run ./cmd/tui
```

### TODOs

- Bulk delete keys
- Update readme and help window
- Allow to type characters only in input field
- Navigation with arrow keys on key list
