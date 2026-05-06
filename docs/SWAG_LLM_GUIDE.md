# SWAG Service Deployment Guide for LLMs

## Overview
This guide helps LLMs construct proper CURL requests to deploy linuxserver/swag services with dynamic NGINX configurations using the deploybot-service-agent API.

## API Endpoint
```
POST /service
Content-Type: application/json
```

## Key Concepts

### Files Parameter
The `files` parameter in `DeployConfig` creates files inside mounted volumes **before** the container starts. This is perfect for NGINX configuration files.

**Format**: `"files": { "host_path": "file_content_as_string" }`

### Volume Mounts
The `volumeMounts` parameter maps host directories to container directories.

**Format**: `"volumeMounts": { "host_path": "container_path" }`

## NGINX Configuration Templates

### Basic Server Block Template
```nginx
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;

    server_name DOMAIN_NAME;

    root /config/www/STATIC_PATH;
    index index.html;

    include /config/nginx/ssl.conf;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location API_PATH {
        resolver 127.0.0.11;
        set $upstream_service SERVICE_NAME;
        proxy_pass http://$upstream_service:SERVICE_PORT;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### WebSocket Support Template
```nginx
location /socket.io {
    resolver 127.0.0.11;
    set $upstream_service SERVICE_NAME;
    proxy_pass http://$upstream_service:WEBSOCKET_PORT;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

## Complete CURL Request Example

### Single Domain Configuration
```bash
curl -X POST https://deploybot-host:8002/service \
  -H "Content-Type: application/json" \
  -d '{
    "imageName": "lscr.io/linuxserver/swag",
    "imageTag": "latest",
    "serviceName": "swag",
    "ports": {
      "80": "80",
      "443": "443"
    },
    "env": [
      "PUID=1000",
      "PGID=1000",
      "TZ=America/New_York",
      "URL=yourdomain.com",
      "SUBDOMAINS=admin,app",
      "VALIDATION=http",
      "EMAIL=admin@yourdomain.com"
    ],
    "volumeMounts": {
      "/opt/swag/config": "/config"
    },
    "files": {
      "/opt/swag/config/nginx/site-confs/admin.conf": "server {\n    listen 443 ssl http2;\n    listen [::]:443 ssl http2;\n\n    server_name admin.yourdomain.com;\n\n    root /config/www/admin-webapp;\n    index index.html;\n\n    include /config/nginx/ssl.conf;\n\n    location / {\n        try_files $uri $uri/ /index.html;\n    }\n\n    location /admin {\n        resolver 127.0.0.11;\n        set $upstream_service services-admin;\n        proxy_pass http://$upstream_service:8888;\n        proxy_set_header Host $host;\n        proxy_set_header X-Real-IP $remote_addr;\n        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\n        proxy_set_header X-Forwarded-Proto $scheme;\n    }\n}"
    },
    "networks": {
      "web-network": "network-id"
    },
    "restartPolicy": {
      "name": "unless-stopped"
    }
  }'
```

### Multi-Domain Configuration
```bash
curl -X POST https://deploybot-host:8002/service \
  -H "Content-Type: application/json" \
  -d '{
    "imageName": "lscr.io/linuxserver/swag",
    "imageTag": "latest", 
    "serviceName": "swag",
    "ports": {
      "80": "80",
      "443": "443"
    },
    "env": [
      "PUID=1000",
      "PGID=1000",
      "TZ=America/New_York",
      "URL=yourdomain.com",
      "SUBDOMAINS=admin,app,api",
      "VALIDATION=http",
      "EMAIL=admin@yourdomain.com"
    ],
    "volumeMounts": {
      "/opt/swag/config": "/config"
    },
    "files": {
      "/opt/swag/config/nginx/site-confs/admin.conf": "server {\n    listen 443 ssl http2;\n    listen [::]:443 ssl http2;\n\n    server_name admin.yourdomain.com;\n\n    root /config/www/admin-webapp;\n    index index.html;\n\n    include /config/nginx/ssl.conf;\n\n    location / {\n        try_files $uri $uri/ /index.html;\n    }\n\n    location /admin {\n        resolver 127.0.0.11;\n        set $upstream_service services-admin;\n        proxy_pass http://$upstream_service:8888;\n        proxy_set_header Host $host;\n        proxy_set_header X-Real-IP $remote_addr;\n        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\n        proxy_set_header X-Forwarded-Proto $scheme;\n    }\n\n    location /adminer {\n        resolver 127.0.0.11;\n        set $upstream_service adminer;\n        proxy_pass http://$upstream_service:8080;\n        proxy_set_header Host $host;\n        proxy_set_header X-Real-IP $remote_addr;\n        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\n        proxy_set_header X-Forwarded-Proto $scheme;\n    }\n}",
      "/opt/swag/config/nginx/site-confs/app.conf": "server {\n    listen 443 ssl http2;\n    listen [::]:443 ssl http2;\n\n    server_name app.yourdomain.com;\n\n    root /config/www/app-webapp;\n    index index.html;\n\n    include /config/nginx/ssl.conf;\n\n    location / {\n        try_files $uri $uri/ /index.html;\n    }\n\n    location /app {\n        resolver 127.0.0.11;\n        set $upstream_service services-app;\n        proxy_pass http://$upstream_service:8889;\n        proxy_set_header Host $host;\n        proxy_set_header X-Real-IP $remote_addr;\n        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\n        proxy_set_header X-Forwarded-Proto $scheme;\n    }\n\n    location /socket.io {\n        resolver 127.0.0.11;\n        set $upstream_service services-app;\n        proxy_pass http://$upstream_service:4999;\n        proxy_http_version 1.1;\n        proxy_set_header Upgrade $http_upgrade;\n        proxy_set_header Connection \"upgrade\";\n        proxy_set_header Host $host;\n        proxy_set_header X-Real-IP $remote_addr;\n        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\n        proxy_set_header X-Forwarded-Proto $scheme;\n    }\n}",
      "/opt/swag/config/nginx/site-confs/api.conf": "server {\n    listen 443 ssl http2;\n    listen [::]:443 ssl http2;\n\n    server_name api.yourdomain.com;\n\n    include /config/nginx/ssl.conf;\n\n    location / {\n        resolver 127.0.0.11;\n        set $upstream_service api-gateway;\n        proxy_pass http://$upstream_service:5000;\n        proxy_set_header Host $host;\n        proxy_set_header X-Real-IP $remote_addr;\n        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\n        proxy_set_header X-Forwarded-Proto $scheme;\n        \n        # CORS headers\n        add_header Access-Control-Allow-Origin \"*\" always;\n        add_header Access-Control-Allow-Methods \"GET, POST, PUT, DELETE, OPTIONS\" always;\n        add_header Access-Control-Allow-Headers \"Origin, Content-Type, Accept, Authorization\" always;\n        \n        if ($request_method = OPTIONS) {\n            return 204;\n        }\n    }\n}"
    },
    "networks": {
      "web-network": "network-id"
    },
    "restartPolicy": {
      "name": "unless-stopped"
    }
  }'
```

## Important Notes for LLMs

### 1. String Escaping
- **All newlines** in NGINX config must be `\\n`
- **All quotes** inside the config must be escaped as `\"`
- **Backslashes** must be escaped as `\\\\`

### 2. File Path Convention
- Host path: `/opt/swag/config/nginx/site-confs/DOMAIN.conf`
- Container path: `/config/nginx/site-confs/DOMAIN.conf` (via volumeMount)

### 3. Required Environment Variables
```json
"env": [
  "PUID=1000",
  "PGID=1000", 
  "TZ=America/New_York",
  "URL=yourdomain.com",
  "SUBDOMAINS=admin,app,api",
  "VALIDATION=http",
  "EMAIL=admin@yourdomain.com"
]
```

### 4. Network Requirements
- All services (SWAG + backend services) must be in the same Docker network
- Use actual network ID from the `networks` API: `GET /networks`

### 4.1. Permission Issues (UID 911) - RECOMMENDED SOLUTION
**Problem**: SWAG container runs as UID 911 internally, creating files owned by UID 911
**Best Solution**: Set PUID/PGID to match your host user - this prevents permission issues entirely

```bash
# Get your user and group IDs
id -u  # Returns your user ID (e.g., 1000)
id -g  # Returns your group ID (e.g., 1000)

# Use these values in SWAG environment variables
"env": [
  "PUID=1000",    # Replace with your actual UID
  "PGID=1000",    # Replace with your actual GID
  "URL=yourdomain.com",
  "SUBDOMAINS=api,app",
  "VALIDATION=http",
  "EMAIL=admin@yourdomain.com"
]
```

**Alternative (not recommended)**: Fix permissions after SWAG creates directories
```bash
# Only use this if you can't set PUID/PGID
sudo chown -R admin:admin /home/admin/config
sudo chmod -R 755 /home/admin/config
```

### 5. Service Dependencies
Deploy backend services **first**, then SWAG with proper PUID/PGID:
```bash
# 1. Create network
curl -X POST https://deploybot-host:8002/network -d '{"name": "web-network"}'

# 2. Deploy backend services
curl -X POST https://deploybot-host:8002/service -d '{"serviceName": "services-admin", ...}'
curl -X POST https://deploybot-host:8002/service -d '{"serviceName": "services-app", ...}'

# 3. Get your user/group IDs
USER_ID=$(id -u)    # e.g., 1000
GROUP_ID=$(id -g)   # e.g., 1000

# 4. Deploy SWAG with custom configs and correct PUID/PGID
curl -X POST https://deploybot-host:8002/service -d '{
  "serviceName": "swag",
  "imageName": "lscr.io/linuxserver/swag",
  "env": [
    "PUID='$USER_ID'",
    "PGID='$GROUP_ID'",
    "URL=yourdomain.com", 
    "SUBDOMAINS=api,app", 
    "VALIDATION=http",
    "EMAIL=admin@yourdomain.com"
  ],
  "volumeMounts": {"/home/admin/config": "/config"},
  "files": {
    "/home/admin/config/nginx/site-confs/api.conf": "server { ... }"
  }
}'
```

**Why this works better**:
- SWAG runs as your user from the start
- All files created have correct ownership  
- No permission denied errors
- No need for manual `chown` commands
- SWAG can read/write its own config files without issues

## Common Configuration Patterns

### Database Admin Tools
```json
{
  "/opt/swag/config/nginx/site-confs/admin.conf": "server {\n    listen 443 ssl http2;\n    listen [::]:443 ssl http2;\n\n    server_name admin.yourdomain.com;\n\n    include /config/nginx/ssl.conf;\n\n    location /adminer {\n        resolver 127.0.0.11;\n        set $upstream_service adminer;\n        proxy_pass http://$upstream_service:8080;\n        proxy_set_header Host $host;\n        proxy_set_header X-Real-IP $remote_addr;\n        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\n        proxy_set_header X-Forwarded-Proto $scheme;\n    }\n\n    location /pgadmin {\n        resolver 127.0.0.11;\n        set $upstream_service pgadmin;\n        proxy_pass http://$upstream_service:80;\n        proxy_set_header Host $host;\n        proxy_set_header X-Script-Name /pgadmin;\n        proxy_set_header X-Forwarded-Host $host;\n        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\n        proxy_set_header X-Forwarded-Proto $scheme;\n    }\n}"
}
```

### Static File + API Pattern
```json
{
  "/opt/swag/config/nginx/site-confs/app.conf": "server {\n    listen 443 ssl http2;\n    listen [::]:443 ssl http2;\n\n    server_name app.yourdomain.com;\n\n    root /config/www/app-frontend;\n    index index.html;\n\n    include /config/nginx/ssl.conf;\n\n    # Serve static files\n    location / {\n        try_files $uri $uri/ /index.html;\n    }\n\n    # Proxy API requests\n    location /api {\n        resolver 127.0.0.11;\n        set $upstream_service app-backend;\n        proxy_pass http://$upstream_service:3000;\n        proxy_set_header Host $host;\n        proxy_set_header X-Real-IP $remote_addr;\n        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\n        proxy_set_header X-Forwarded-Proto $scheme;\n    }\n}"
}
```

This guide provides LLMs with all the necessary information to construct proper CURL requests for SWAG deployment with dynamic NGINX configurations.
