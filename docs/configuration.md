# Configuration

All configuration is read from environment variables at startup using the `envconfig` library. No configuration files are required.

## Environment Variables

| Variable | Required | Description |
|---|---|---|
| `SERVICE_PORT` | Yes | Port the HTTP server listens on (e.g. `8002`) |
| `SERVICE_CRT` | No | Path to a TLS certificate file. When set together with `SERVICE_KEY`, the server uses HTTPS. |
| `SERVICE_KEY` | No | Path to a TLS private key file. |
| `API_BASE_URL` | Yes | Base URL of the DeployBot orchestration API (e.g. `https://api.example.com`) |
| `API_KEY` | Yes | API key sent as `X-Api-Key` header on all outbound calls to the DeployBot API |
| `DOCKER_HOST` | Yes | Docker daemon address. Use `unix:///var/run/docker.sock` for local socket or `tcp://host:2376` for remote. |
| `DH_USERNAME` | No | Docker Hub username for pushing built images. Required for build tasks. |
| `DH_PASSWORD` | No | Docker Hub password or access token for pushing built images. Required for build tasks. |
| `REPO_USERNAME` | No | Git username for cloning private repositories. Required for build tasks against private repos. |
| `REPO_PASSWORD` | No | Git password or personal access token for private repositories. Required for build tasks against private repos. |

## TLS Configuration

When `SERVICE_CRT` and `SERVICE_KEY` are both set, the service starts an HTTPS server using those credentials. When either is absent, the service falls back to plain HTTP.

Self-signed certificates can be generated with the included `certgen.sh` script:

```bash
./certgen.sh
```

## Verifying Configuration

Run the binary with the `env` command to print the resolved configuration without starting the service:

```bash
./bot_agent-linux-x86_64 env
```

This prints the `Config` struct with all values populated from the environment, which is useful for verifying variable names and values at deployment time.

## Example — Minimal Development Setup

```bash
export SERVICE_PORT=8002
export API_BASE_URL=http://localhost:3000
export API_KEY=dev-secret
export DOCKER_HOST=unix:///var/run/docker.sock
```

## Example — Production Setup with TLS and Registry

```bash
export SERVICE_PORT=8443
export SERVICE_CRT=/etc/ssl/certs/agent.crt
export SERVICE_KEY=/etc/ssl/private/agent.key
export API_BASE_URL=https://deploybot-api.example.com
export API_KEY=<strong-random-secret>
export DOCKER_HOST=unix:///var/run/docker.sock
export DH_USERNAME=myorg
export DH_PASSWORD=<docker-hub-access-token>
export REPO_USERNAME=git
export REPO_PASSWORD=<github-personal-access-token>
```

## Security Notes

- Never commit `API_KEY`, `DH_PASSWORD`, or `REPO_PASSWORD` to source control.
- Use a secret manager (e.g. AWS Secrets Manager, HashiCorp Vault, systemd `EnvironmentFile` with restricted permissions) to inject credentials at runtime.
- Restrict access to the agent port using firewall rules. The API has no built-in authentication.
