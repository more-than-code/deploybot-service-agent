# Development Guide

## Prerequisites

| Tool | Version | Purpose |
|---|---|---|
| Go | 1.21+ | Build and test |
| Docker | 24+ | Required to run the service |
| `sha256sum` | any | Checksum generation in `build.sh` |

## Repository Layout

```
deploybot-service-agent/
├── main.go                  # Entry point, config, HTTP server, route registration
├── main_test.go             # Integration-style smoke tests
├── go.mod / go.sum          # Module dependencies
├── build.sh                 # Cross-compile script (x86_64 + aarch64)
├── certgen.sh               # Self-signed TLS certificate generator
├── Dockerfile               # Multi-stage Docker image build
├── docker-compose.yaml      # Example stack for local testing
├── api/
│   ├── api.go               # REST API handlers
│   └── scheduler.go         # Webhook handler, build/deploy task execution
├── deploybot-types/         # Vendored shared types from deploybot-types module
├── model/
│   └── models.go            # Request/response structs
├── util/
│   ├── containerHelper.go   # Docker client wrapper
│   └── utils.go             # Git, tar, filesystem utilities
└── docs/                    # This documentation
```

## Building

### Development build (current platform)

```bash
go build -o bot_agent ./main.go
```

### Release build (Linux, both architectures)

```bash
./build.sh 1.2.3
```

This produces:
- `bot_agent-linux-x86_64`
- `bot_agent-linux-x86_64.sha256`
- `bot_agent-linux-aarch64`
- `bot_agent-linux-aarch64.sha256`

The version string is embedded via `-ldflags "-X main.Version=<version>"` and printed by `app version`.

### Docker image build

```bash
docker build -t deploybot-service-agent:latest .
```

The Dockerfile uses a two-stage build: Go compiler in `golang:1.21-alpine`, runtime in `alpine:3.18`.

## Running Locally

### 1. Start a Docker daemon

The service needs a reachable Docker daemon. On macOS, Docker Desktop provides `unix:///var/run/docker.sock`.

### 2. Set environment variables

```bash
export SERVICE_PORT=8002
export API_BASE_URL=http://localhost:3000   # or a real DeployBot API
export API_KEY=dev-secret
export DOCKER_HOST=unix:///var/run/docker.sock
```

See [configuration.md](configuration.md) for the full variable reference.

### 3. Start the service

```bash
go run main.go start
# or
./bot_agent start
```

### 4. Verify

```bash
curl http://localhost:8002/healthCheck
```

### TLS (optional)

```bash
./certgen.sh                  # generates cert.pem and key.pem
export SERVICE_CRT=cert.pem
export SERVICE_KEY=key.pem
go run main.go start
```

## Testing

### Smoke tests

`main_test.go` contains basic HTTP integration tests that expect a running instance on port `8083`. Start the service first, then:

```bash
go test ./...
```

### Manual testing with test.http

The `test.http` file contains pre-built HTTP requests for common operations and can be run directly in VS Code with the REST Client extension or converted to `curl` commands.

## Dependencies

Key dependencies and their roles:

| Package | Role |
|---|---|
| `github.com/gin-gonic/gin` | HTTP router and middleware framework |
| `github.com/gin-contrib/cors` | CORS middleware for the Gin server |
| `github.com/docker/docker` | Docker Engine API client |
| `github.com/docker/go-connections` | Port binding utilities |
| `github.com/go-git/go-git/v5` | Pure-Go Git client for cloning repos |
| `github.com/kelseyhightower/envconfig` | Environment variable config binding |
| `gopkg.in/mgo.v2` | BSON utilities (used for ObjectId serialisation) |

Install/update dependencies:

```bash
go mod tidy
```

## Code Conventions

- Error handling: errors are propagated up to handlers and returned as HTTP 500 responses with the error string as the body or `msg` field.
- Docker operations are synchronous in `containerHelper.go` and called from handlers or goroutines in `scheduler.go`.
- The `resolveParam` helper in `api.go` checks both path parameters and query parameters, giving path parameters priority. This preserves backward compatibility as routes migrated from query-only to path-parameter style.
- Build tasks run fully synchronously in their goroutine (clone → build → push) before reporting status.
- Deploy tasks call `StartContainer` inside a nested goroutine (fire-and-forget) because image pulls can be slow; the outer goroutine reports `Done` after launching the inner one.

## Release Process

1. Update version argument to `build.sh`.
2. Run `./build.sh <version>`.
3. Verify checksums with `sha256sum -c *.sha256`.
4. Distribute binaries to target hosts via the DeployBot install mechanism (`install.sh`).
