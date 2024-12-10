#!/bin/bash

VERSION=$1

if [ -z "$VERSION" ]; then
  VERSION="latest"
fi  

# Define paths
INSTALLER_USER=$(logname)  # Dynamically identify the user who initiated the script
USER_HOME=$(getent passwd "$INSTALLER_USER" | cut -d: -f6)
PROGRAM_PATH="/usr/local/bin/bot_agent"
SERVICE_FILE="/etc/systemd/system/bot_agent.service"
ENV_FILE="$USER_HOME/.bot_agent/env.conf"

# URL to download the file
FILE_URL="https://github.com/more-than-code/deploybot-service-agent/releases/download/$VERSION/bot_agent-linux-$(uname -m)"

# Ensure the user's home-based folder structure exists
BOT_AGENT_DIR="$USER_HOME/.bot_agent"
mkdir -p "$BOT_AGENT_DIR"
chmod 700 "$BOT_AGENT_DIR"

# Create environment configuration if it doesn't exist
if [ ! -f "$ENV_FILE" ]; then
  cat << EOF > "$ENV_FILE"
SERVICE_PORT=8080
SERVICE_CRT=$BOT_AGENT_DIR/service.crt
SERVICE_KEY=$BOT_AGENT_DIR/service.key
API_KEY=your_api_key
API_BASE_URL=https://api.example.com
DOCKER_HOST=unix:///var/run/docker.sock
DH_USERNAME=dockerhub_user
DH_PASSWORD=dockerhub_password
REPO_USERNAME=repo_user
REPO_PASSWORD=repo_password
EOF
  chmod 600 "$ENV_FILE"
fi

# Download the program
curl -fsSL --retry 3 --retry-delay 5 -o "$PROGRAM_PATH" "$FILE_URL"

if [ $? -ne 0 ]; then
  echo "Failed to download $FILE_URL"
  exit 1
fi

# Make the downloaded file executable and secure
chmod 750 "$PROGRAM_PATH"
chown root:root "$PROGRAM_PATH"

# Secure the service file creation
cat << EOF > "$SERVICE_FILE"
[Unit]
Description=BotAgent@$VERSION
After=network.target docker.service

[Service]
Type=simple
ExecStart=$PROGRAM_PATH start
Restart=on-failure
RestartSec=30s
User=$INSTALLER_USER
EnvironmentFile=$ENV_FILE
WorkingDirectory=$BOT_AGENT_DIR
PrivateTmp=true
ProtectSystem=full
ProtectHome=true
ReadWritePaths=$BOT_AGENT_DIR
SupplementaryGroups=docker  # Allow access to Docker
NoNewPrivileges=true
CapabilityBoundingSet=CAP_NET_BIND_SERVICE

[Install]
WantedBy=multi-user.target
EOF

# Secure the service file
chmod 644 "$SERVICE_FILE"

# Reload systemd to load the new service file
systemctl daemon-reload

# Start and enable the service
systemctl start bot_agent
systemctl enable bot_agent

# Display the status of the service
systemctl status bot_agent
