# Architecture

## Overview

deploybot-service-agent is an agent process installed on each target deployment host. It connects the DeployBot orchestration platform to the local Docker daemon, handling the actual execution of build and deploy tasks.

```
DeployBot API ──► /streamWebhook ──► Scheduler ──► Docker Daemon
                                         │
                                         └──► Report task status back to API
```

## Components

### `main.go` — Entry Point

Loads configuration from environment variables via `envconfig`, initialises the Gin HTTP server with CORS middleware, registers all routes, and starts either HTTP or HTTPS depending on whether TLS credentials are supplied.

CLI interface:

| Command | Behaviour |
|---|---|
| `app start` | Start the HTTP service |
| `app version` | Print the compiled version string |
| `app env` | Print the resolved configuration (useful for debugging) |

### `api/scheduler.go` — Task Scheduler

The core orchestration component. Responsibilities:

- **`StreamWebhookHandler`** — receives task trigger payloads from the DeployBot API, fetches the full task definition, marks the task as `InProgress`, and dispatches execution in a background goroutine.
- **`DoBuildTask`** — clones a Git repository, tars the source tree, and invokes the Docker BuildKit API to build and push an image to a registry.
- **`DoDeployTask`** — writes any required config files to the host filesystem, then starts a container using the deploy configuration.
- **`updateTaskStatus` / `ProcessPostTask`** — call back to the DeployBot API to set a task's status to `Done`, `Failed`, or `TimedOut`.
- **`cleanUp`** — implements optional per-task timeouts; fires a `TimedOut` status update if the task exceeds its configured duration.

### `api/api.go` — REST API Handlers

Direct-management endpoints that wrap Docker operations. They do not go through the webhook flow and are intended for use by an operator or the DeployBot UI.

Handlers:

| Handler | Purpose |
|---|---|
| `CreateService` | Pull image and start a named container |
| `UpdateService` | Stop or restart a running container |
| `DeleteService` | Force-remove a container |
| `GetService` | Inspect a container |
| `GetServices` | List all containers |
| `GetServiceLog` | Tail or stream container logs |
| `GetDiskInfo` | Report filesystem usage for a mount path |
| `CreateNetwork` | Create a bridge network |
| `GetNetwork` | Look up a network by name |
| `GetNetworks` | List all networks |
| `DeleteNetwork` | Remove a network |
| `DeleteImages` | Remove all local Docker images |
| `DeleteBuilderCache` | Prune the BuildKit cache |
| `HealthCheckHandler` | Liveness probe |

### `util/containerHelper.go` — Docker Client Wrapper

Wraps the official `docker/docker` Go client. Translates service-level concepts (service name, volume mounts, port bindings, restart policies, log drivers) into Docker API calls.

Key behaviours:
- Before starting a container, any existing container with the same name is stopped and removed to ensure a clean replacement.
- Volume mounts are classified as **bind mounts** (host paths starting with `/` or `.`) or **named volumes** (everything else) and configured accordingly.
- Image pushes authenticate to Docker Hub using base64-encoded registry auth.

### `util/utils.go` — Utilities

| Function | Purpose |
|---|---|
| `CloneRepo` | Clone a Git repository at a specific branch using basic auth |
| `TarFiles` | Walk a directory and produce a tar archive suitable for the Docker build API |
| `WriteToFile` | Write string content to a host path, creating parent directories as needed |
| `CreateDirsIfNotExist` | Ensure a directory hierarchy exists |
| `GetDiskInfo` | Read filesystem statistics via `statfs` syscall |

### `model/models.go` — Request / Response Models

Defines the JSON schemas for all HTTP inputs and outputs:

- `BuildConfig` — parameters for a build task (image name/tag, Dockerfile path, repo URL/branch, build args)
- `DeployConfig` — parameters for a deploy task (image, ports, volumes, env vars, networks, restart policy, log config)
- `DeployConfig` is also accepted directly by `POST /service` for ad-hoc deployments
- `ApiResponse` — generic envelope with `code`, `msg`, and `payload` fields

### `deploybot-types/` — Shared Types

A vendored copy of the `more-than-code/deploybot-types` module. Contains:

- `Task` / `TaskType` — task definition and type constants (`build`, `deploy`)
- `StreamWebhook` — webhook payload shape
- `ObjectId` — BSON-compatible 12-byte ID with JSON/hex marshalling
- Status constants: `Pending`, `InProgress`, `Done`, `Failed`, `Canceled`, `TimedOut`
- Error code and message constants

## Data Flow

### Webhook-triggered task

```
1. DeployBot API sends POST /streamWebhook
   Payload: { pipelineId, taskId, arguments }

2. Agent fetches full task: GET {API_BASE_URL}/task?pid=...&id=...

3. Agent marks task InProgress: PUT {API_BASE_URL}/taskStatus

4. HTTP 200 returned to caller immediately

5. Background goroutine executes:
   - BuildTask:  clone repo → tar files → docker build → docker push
   - DeployTask: write files → start container (async)

6. Agent marks task Done or Failed: PUT {API_BASE_URL}/taskStatus

7. If timeout configured: timer fires TimedOut before step 6 if exceeded
```

### Direct API call

```
1. Client sends request (e.g. POST /service)
2. Handler validates input
3. ContainerHelper performs Docker operation
4. Handler returns JSON response
```

## Deployment Model

The agent is distributed as a statically linked binary for Linux (`x86_64` and `aarch64`). It can also run as a Docker container (see `Dockerfile`). The binary connects to the Docker daemon on the same host, either via the default Unix socket or a TCP address configured through `DOCKER_HOST`.

TLS is optional. When `SERVICE_CRT` and `SERVICE_KEY` are set, the server listens with HTTPS; otherwise it uses plain HTTP.
