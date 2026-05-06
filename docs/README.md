# deploybot-service-agent Documentation

**deploybot-service-agent** is a Go-based HTTP service that runs on a target host and acts as the execution layer for the DeployBot CI/CD orchestration system. It receives task triggers from the DeployBot API, executes build and deploy pipelines against the local Docker daemon, and exposes a REST API for direct container and network management.

## Documentation Index

| Document | Description |
|---|---|
| [Architecture](architecture.md) | System design, component overview, and data flow |
| [API Reference](api-reference.md) | Complete HTTP endpoint reference with request/response schemas |
| [Configuration](configuration.md) | Environment variables and runtime configuration |
| [Development Guide](development.md) | Building, running, and testing the service locally |

## Quick Start

### 1. Set environment variables

```bash
export SERVICE_PORT=8002
export API_BASE_URL=https://your-deploybot-api.example.com
export API_KEY=your-api-key
export DOCKER_HOST=unix:///var/run/docker.sock
export DH_USERNAME=your-dockerhub-username
export DH_PASSWORD=your-dockerhub-password
export REPO_USERNAME=your-git-username
export REPO_PASSWORD=your-git-token
```

### 2. Run the service

```bash
# Build
./build.sh 1.0.0

# Start
./bot_agent-linux-x86_64 start
```

### 3. Verify

```bash
curl http://localhost:8002/healthCheck
```
