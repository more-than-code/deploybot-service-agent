# Service Deployment Guide for deploybot-service-agent

## Overview
The deploybot-service-agent is a Go-based REST API service that provides programmatic Docker container management capabilities. It allows you to deploy, manage, and monitor containerized services through HTTP endpoints, offering an alternative to direct Docker Compose or Docker CLI usage.

## Core Concepts

### Service vs Container
- **Service**: A logical unit representing an application (e.g., nginx, postgres, redis)
- **Container**: The actual Docker container instance running the service
- Services can be identified by either `name` or `container_id` in API calls

### Networks
- Services can be connected to Docker networks for inter-container communication
- Network names in Docker Compose are prefixed with the project directory name
- Example: `synerthink-net` becomes `deploybot-service-agent_synerthink-net`

## API Endpoints

### Base URL
```
# Development
http://localhost:8002

# Production (HTTPS)
https://{HOST}:{PORT}
```

**Note:** In production environments, the service should be accessed via HTTPS. Replace `{HOST}` with your server hostname/IP and `{PORT}` with the configured service port.

### Production Deployment
For production environments, ensure:
- **HTTPS Configuration**: Configure TLS certificates using `SERVICE_CRT` and `SERVICE_KEY` environment variables
- **Security**: Implement proper authentication and authorization mechanisms
- **Network Security**: Use firewalls and network policies to restrict access
- **Monitoring**: Set up logging and monitoring for the service agent itself

### Health Check
```http
GET /healthCheck
```
Verifies the service is running and responsive.

### Service Management

#### Deploy a Service
```http
POST /service
Content-Type: application/json

{
  "imageName": "nginx",
  "imageTag": "latest",
  "serviceName": "my-nginx",
  "env": [
    "ENV_VAR1=value1",
    "ENV_VAR2=value2"
  ],
  "ports": {"80": "8080"},
  "networks": {"network-name": "network-id"},
  "restartPolicy": {
    "Name": "unless-stopped",
    "MaximumRetryCount": 0
  }
}
```

**Field Specifications:**
- `imageName`: Docker image name (required)
- `imageTag`: Image tag, defaults to "latest"
- `serviceName`: Container name for identification
- `env`: Array of environment variables in "KEY=value" format
- `ports`: Map of container_port:host_port (both as strings)
- `networks`: Map of network_name:network_id
- `restartPolicy`: Docker restart policy configuration

#### Get Service Information
```http
GET /service/{name_or_id}
```
Returns detailed information about a service including container ID, status, and configuration.

#### Update Service
```http
PUT /service
Content-Type: application/json
```
Updates an existing service configuration.

#### Delete Service
```http
DELETE /service/{name_or_id}
```
Stops and removes a service/container.

#### List All Services
```http
GET /services
```
Returns a list of all managed services.

#### Get Service Logs
```http
GET /serviceLogs?name={service_name}&showStdout=true&showStderr=true&timestamps=true&tail=10&follow=false&details=false&since={timestamp}&until={timestamp}
```
Retrieves logs from a service container with comprehensive filtering options.

**Query Parameters:**
- `name` (required): Container/service name or ID
- `showStdout` (boolean): Include stdout logs (default: false)
- `showStderr` (boolean): Include stderr logs (default: false)
- `timestamps` (boolean): Include timestamps in log output (default: false)
- `tail` (string): Number of lines to show from the end of logs (e.g., "10", "100")
- `follow` (boolean): Stream logs continuously (default: false)
- `details` (boolean): Show additional log details (default: false)
- `since` (string): Show logs since timestamp (RFC3339 format, e.g., "2025-01-01T00:00:00Z")
- `until` (string): Show logs until timestamp (RFC3339 format, e.g., "2025-01-01T23:59:59Z")

**Examples:**
```bash
# Basic log retrieval with timestamps (development)
curl "http://localhost:8002/serviceLogs?name=nginx&showStdout=true&timestamps=true&tail=10"

# Basic log retrieval with timestamps (production)
curl "https://{HOST}:{PORT}/serviceLogs?name=nginx&showStdout=true&timestamps=true&tail=10"

# Get both stdout and stderr logs
curl "https://{HOST}:{PORT}/serviceLogs?name=redis-service&showStdout=true&showStderr=true&tail=50"

# Get logs without timestamps (clean output)
curl "https://{HOST}:{PORT}/serviceLogs?name=mysql&showStdout=true&timestamps=false&tail=20"

# Get logs from a specific time range
curl "https://{HOST}:{PORT}/serviceLogs?name=nginx&showStdout=true&since=2025-08-01T00:00:00Z&until=2025-08-01T23:59:59Z"
```

### Network Management

#### Get Network Information
```http
GET /network/{network_name}
```
Returns network details including network ID.

#### Create Network
```http
POST /network
Content-Type: application/json

{
  "name": "my-network",
  "driver": "bridge"
}
```

#### Delete Network
```http
DELETE /network/{network_name}
```

#### List Networks
```http
GET /networks
```

## Common Deployment Patterns

### 1. Simple Web Service
```json
{
  "imageName": "nginx",
  "imageTag": "latest",
  "serviceName": "web-server",
  "ports": {"80": "8080"}
}
```

### 2. Database Service
```json
{
  "imageName": "postgres",
  "imageTag": "16",
  "serviceName": "database",
  "env": [
    "POSTGRES_DB=myapp",
    "POSTGRES_USER=admin",
    "POSTGRES_PASSWORD=secret"
  ],
  "networks": {"app-network": "network-id"},
  "restartPolicy": {
    "Name": "unless-stopped"
  }
}
```

### 3. Application Service with Dependencies
```json
{
  "imageName": "trowsoft/synerthink-app-service",
  "imageTag": "latest",
  "serviceName": "synerthink-app-service",
  "env": [
    "SERVER_PORT=8888",
    "PT_HOST=postgres",
    "PT_PORT=5432",
    "PT_DATABASE=synerthink",
    "PT_USERNAME=dev",
    "PT_PASSWORD=smart",
    "REDIS_HOST=127.0.0.1",
    "REDIS_PORT=6379",
    "REDIS_PASSWORD=smartredis"
  ],
  "ports": {"8888": "8888"},
  "networks": {"deploybot-service-agent_synerthink-net": "network-id"},
  "restartPolicy": {
    "Name": "unless-stopped"
  }
}
```

## Workflow Examples

### Complete Service Stack Deployment

1. **Create Network (if needed)**
```bash
curl -X POST https://{HOST}:{PORT}/network \
  -H "Content-Type: application/json" \
  -d '{"name": "app-network", "driver": "bridge"}'
```

2. **Get Network ID**
```bash
curl https://{HOST}:{PORT}/network/app-network
```

3. **Deploy Database**
```bash
curl -X POST https://{HOST}:{PORT}/service \
  -H "Content-Type: application/json" \
  -d '{
    "imageName": "postgres",
    "imageTag": "16",
    "serviceName": "postgres",
    "env": ["POSTGRES_DB=myapp", "POSTGRES_USER=admin", "POSTGRES_PASSWORD=secret"],
    "networks": {"app-network": "network-id"}
  }'
```

4. **Deploy Application**
```bash
curl -X POST https://{HOST}:{PORT}/service \
  -H "Content-Type: application/json" \
  -d '{
    "imageName": "myapp",
    "imageTag": "latest",
    "serviceName": "myapp",
    "env": ["DB_HOST=postgres", "DB_PORT=5432"],
    "ports": {"8000": "8000"},
    "networks": {"app-network": "network-id"}
  }'
```

5. **Verify Deployment**
```bash
curl https://{HOST}:{PORT}/services
curl https://{HOST}:{PORT}/service/myapp
```

## Error Handling

### Common Errors and Solutions

1. **"invalid reference format"**
   - Check image name format
   - Verify environment variables don't contain invalid characters
   - Ensure all required fields are present

2. **"json: cannot unmarshal array into Go struct field"**
   - `ports` must be a map: `{"80": "8080"}` not `["80:8080"]`
   - `networks` must be a map: `{"net": "id"}` not `["net"]`

3. **"network not found"**
   - Use full network name including project prefix
   - Get network ID first: `GET /network/{name}`
   - Create network if it doesn't exist

4. **Port conflicts**
   - Check if host port is already in use
   - Use different host ports for multiple services

## Best Practices

### 1. Network Management
- Always get network ID before deploying services
- Use consistent network naming across related services
- Consider network isolation for different application stacks

### 2. Service Configuration
- Use meaningful service names for easy identification
- Set appropriate restart policies based on service criticality
- Configure environment variables for service connectivity

### 3. Monitoring and Logs
- Use service logs endpoint for debugging with appropriate parameters
- Enable both stdout and stderr logging for comprehensive troubleshooting
- Use timestamps for log correlation and time-based analysis
- Limit log output with `tail` parameter to avoid overwhelming responses
- Monitor service status after deployment with regular log checks
- Implement health checks for critical services

**Log Monitoring Best Practices:**
- Start with recent logs: `?showStdout=true&timestamps=true&tail=20`
- Check errors: `?showStderr=true&timestamps=true&tail=50`
- For debugging: `?showStdout=true&showStderr=true&timestamps=true&tail=100`
- Time-based analysis: Use `since` and `until` parameters for specific periods

### 4. Resource Management
- Clean up unused services and networks
- Monitor system resources (disk, memory)
- Use appropriate image tags (avoid 'latest' in production)

## Integration with Docker Compose

The API can complement Docker Compose workflows:

1. **Use Compose for initial setup** (networks, base services)
2. **Use API for dynamic service management** (scaling, updates, temporary services)
3. **Maintain consistency** between Compose and API deployments

Example Docker Compose to API migration:
```yaml
# docker-compose.yaml
services:
  app:
    image: myapp:latest
    ports:
      - "8000:8000"
    environment:
      - DB_HOST=postgres
    networks:
      - app-net
```

Equivalent API call:
```json
{
  "imageName": "myapp",
  "imageTag": "latest",
  "serviceName": "app",
  "ports": {"8000": "8000"},
  "env": ["DB_HOST=postgres"],
  "networks": {"project_app-net": "network-id"}
}
```

## Troubleshooting

### Debug Workflow
1. Check service status: `GET /service/{name}`
2. Review service logs: `GET /serviceLogs?name={name}&showStdout=true&showStderr=true&timestamps=true&tail=50`
3. Verify network connectivity: `GET /network/{name}`
4. List all services: `GET /services`
5. Check system resources: `GET /diskInfo`

### Log Analysis Examples
```bash
# Quick status check with recent logs
curl https://{HOST}:{PORT}/service/nginx
curl "https://{HOST}:{PORT}/serviceLogs?name=nginx&showStdout=true&timestamps=true&tail=10"

# Debug service startup issues
curl "https://{HOST}:{PORT}/serviceLogs?name=failing-service&showStderr=true&timestamps=true&tail=50"

# Monitor service in real-time (follow mode)
curl "https://{HOST}:{PORT}/serviceLogs?name=web-app&showStdout=true&follow=true&timestamps=true"

# Historical log analysis
curl "https://{HOST}:{PORT}/serviceLogs?name=database&showStdout=true&since=2025-08-07T00:00:00Z&until=2025-08-07T23:59:59Z"
```

### Common Issues
- **Service startup failures**: Check logs with `?showStderr=true&timestamps=true&tail=50` and verify environment variables
- **Network connectivity**: Verify network configuration, service names, and check application logs for connection errors
- **Port accessibility**: Ensure host ports are not blocked, check service logs for binding errors
- **Image pull failures**: Verify image name and registry access, check stderr logs for detailed error messages
- **Performance issues**: Use time-based log analysis with `since` and `until` parameters to identify patterns
- **Missing logs**: Ensure both `showStdout=true` and `showStderr=true` are set when troubleshooting

**Log-Based Troubleshooting:**
```bash
# Check for startup errors
curl "https://{HOST}:{PORT}/serviceLogs?name=service&showStderr=true&timestamps=true&tail=100"

# Monitor application behavior
curl "https://{HOST}:{PORT}/serviceLogs?name=service&showStdout=true&timestamps=true&since=2025-08-08T10:00:00Z"

# Full diagnostic output
curl "https://{HOST}:{PORT}/serviceLogs?name=service&showStdout=true&showStderr=true&timestamps=true&details=true&tail=200"
```

This guide provides a comprehensive foundation for understanding and using the deploybot-service-agent API for container management and service deployment.
