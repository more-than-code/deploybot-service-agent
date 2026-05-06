# API Reference

## Base URL

| Environment | URL |
|---|---|
| Development | `http://localhost:{SERVICE_PORT}` |
| Production (TLS) | `https://{host}:{SERVICE_PORT}` |

## Authentication

Direct API endpoints have no built-in authentication. Secure access using network-level controls (firewall rules, reverse proxy with auth). The agent authenticates **outbound** calls to the DeployBot API using the `X-Api-Key` header.

## Response Envelope

Most endpoints return a JSON envelope:

```json
{
  "code": 0,
  "msg": "",
  "payload": { ... }
}
```

`code` is `0` on success. Non-zero codes map to error categories defined in the constants package (e.g. `1000` = client error, `2000` = server error).

---

## Endpoints

### Health Check

```
GET /healthCheck
```

Returns `200 OK` with an empty body. Use as a liveness probe.

---

### Webhook

#### Receive Task Trigger

```
POST /streamWebhook
Content-Type: application/json
```

Called by the DeployBot API to trigger a pipeline task. The agent fetches the full task definition, marks it `InProgress`, responds immediately, then executes the task in the background.

**Request body:**

```json
{
  "payload": {
    "pipelineId": "<ObjectId hex>",
    "taskId": "<ObjectId hex>"
  },
  "arguments": ["arg1", "arg2"]
}
```

**Response:** `200 OK` with an empty `WebhookResponse` envelope on acceptance, or `400 Bad Request` on error.

---

### Services

#### Create / Deploy a Service

```
POST /service
Content-Type: application/json
```

Pulls the specified image and starts a container. If a container with the same `serviceName` already exists it is stopped and removed before the new one is created.

**Request body (`DeployConfig`):**

```json
{
  "imageName": "nginx",
  "imageTag": "latest",
  "serviceName": "my-nginx",
  "env": ["ENV_VAR=value"],
  "ports": { "80": "8080" },
  "volumeMounts": { "/host/path": "/container/path" },
  "files": { "/host/path/nginx.conf": "<file content>" },
  "networks": { "my-network": "<network-id>" },
  "restartPolicy": {
    "name": "unless-stopped",
    "maximumRetryCount": 0
  },
  "logConfig": {
    "type": "json-file",
    "config": { "max-size": "10m" }
  },
  "command": "nginx -g 'daemon off;'",
  "links": ["other-container"],
  "autoRemove": false,
  "shmSize": 0
}
```

**Field reference:**

| Field | Type | Required | Description |
|---|---|---|---|
| `imageName` | string | Yes | Docker image name |
| `imageTag` | string | No | Image tag (default: `latest`) |
| `serviceName` | string | Yes | Container name |
| `env` | []string | No | Environment variables in `KEY=value` format |
| `ports` | map[string]string | No | `containerPort` → `hostPort` |
| `volumeMounts` | map[string]string | No | `hostPath` (or named volume) → `containerPath`. Host paths (starting with `/` or `.`) create bind mounts; other keys create named volumes. |
| `files` | map[string]string | No | Files written to the host before the container starts. Key = host path, value = file content. |
| `networks` | map[string]string | No | `networkName` → `networkId` |
| `restartPolicy` | object | No | Docker restart policy (`name`, `maximumRetryCount`) |
| `logConfig` | object | No | Docker log driver config (`type`, `config` map) |
| `command` | string | No | Override the container entrypoint command (space-delimited) |
| `links` | []string | No | Container links (legacy Docker feature) |
| `autoRemove` | bool | No | Automatically remove the container on exit |
| `shmSize` | int64 | No | Shared memory size in bytes |

**Response:** `200 OK`

```json
{ "status": "deployment started" }
```

---

#### List Services

```
GET /services
```

Returns all containers visible to the Docker daemon.

**Response:** `200 OK`

```json
{
  "code": 0,
  "msg": "",
  "payload": [ { ...docker container summary... } ]
}
```

---

#### Get Service

```
GET /service/:name
```

Returns detailed inspection output for a container.

| Parameter | Location | Description |
|---|---|---|
| `name` | path | Container name or ID |

**Response:** `200 OK`

```json
{
  "code": 0,
  "msg": "",
  "payload": { ...docker inspect output... }
}
```

---

#### Update Service

```
PUT /service/:name
Content-Type: application/json
```

Stop or restart a running container.

**Request body:**

```json
{
  "name": "my-nginx",
  "running": true,
  "restarting": false
}
```

| Field | Type | Description |
|---|---|---|
| `restarting` | bool | If `true`, restart the container |
| `running` | bool | If `false` (and `restarting` is `false`), stop the container |

**Response:** `200 OK`

```json
{ "code": 0, "msg": "", "payload": null }
```

---

#### Delete Service

```
DELETE /service/:name
```

Force-removes a container (equivalent to `docker rm -f`).

| Parameter | Location | Description |
|---|---|---|
| `name` | path | Container name or ID |

**Response:** `200 OK`

```json
{ "code": 0, "msg": "", "payload": null }
```

---

### Service Logs

```
GET /serviceLogs
GET /serviceLogs/:name
```

Retrieves logs from a container. Both routes accept the same query parameters. The path parameter form (`/serviceLogs/:name`) is preferred; the query parameter form (`/serviceLogs?name=...`) is supported for backward compatibility.

**Query parameters:**

| Parameter | Type | Default | Description |
|---|---|---|---|
| `name` | string | — | Container name or ID (required when using query form) |
| `showStdout` | bool | `false` | Include stdout |
| `showStderr` | bool | `false` | Include stderr |
| `timestamps` | bool | `false` | Prepend timestamps to each line |
| `tail` | string | `""` (all) | Number of lines from end of logs |
| `follow` | bool | `false` | Stream logs continuously |
| `details` | bool | `false` | Include extra log metadata |
| `since` | string | `""` | Show logs since RFC3339 timestamp |
| `until` | string | `""` | Show logs until RFC3339 timestamp |

**Response:** `200 OK` with plain-text log output.

**Examples:**

```bash
# Last 50 lines of stdout with timestamps
curl "http://localhost:8002/serviceLogs/my-nginx?showStdout=true&timestamps=true&tail=50"

# Both stdout and stderr in a time range
curl "http://localhost:8002/serviceLogs/my-nginx?showStdout=true&showStderr=true&since=2026-01-01T00:00:00Z&until=2026-01-01T23:59:59Z"
```

---

### Networks

#### Create Network

```
POST /network
Content-Type: application/json
```

Creates a new bridge network.

**Request body:**

```json
{ "name": "my-network" }
```

**Response:** `200 OK`

```json
{
  "code": 0,
  "msg": "",
  "payload": { "name": "my-network", "id": "<network-id>" }
}
```

---

#### Get Network

```
GET /network/:name
```

Look up a network by name and return its ID.

**Response:** `200 OK`

```json
{
  "code": 0,
  "msg": "",
  "payload": { "name": "my-network", "id": "<network-id>" }
}
```

---

#### List Networks

```
GET /networks
```

**Response:** `200 OK`

```json
{
  "code": 0,
  "msg": "",
  "payload": [ { "name": "...", "id": "..." } ]
}
```

---

#### Delete Network

```
DELETE /network/:name
```

**Response:** `200 OK`

```json
{ "code": 0, "msg": "" }
```

---

### Disk Information

```
GET /diskInfo/:path
```

Returns filesystem usage statistics for the specified mount point.

| Parameter | Location | Description |
|---|---|---|
| `path` | path | Mount point to inspect (e.g. `/`) |

**Response:** `200 OK`

```json
{
  "code": 0,
  "msg": "",
  "payload": {
    "totalSize": 107374182400,
    "availSize": 53687091200,
    "path": "/"
  }
}
```

Sizes are in bytes.

---

### Image Management

#### Delete All Images

```
DELETE /images
```

Removes all locally cached Docker images (force + prune children). Use with caution — this will remove images needed by stopped or future containers.

**Response:** `200 OK` with `"OK"` body.

---

#### Delete Builder Cache

```
DELETE /builderCache
```

Prunes the Docker BuildKit build cache.

**Response:** `200 OK` with `"OK"` body.

---

## Error Responses

All handlers return a JSON body on error:

```json
{
  "code": 2000,
  "msg": "error description"
}
```

HTTP status codes used: `400 Bad Request`, `500 Internal Server Error`.
