#!/bin/bash

VERSION=$1

if [ -z "$VERSION" ]; then
  VERSION="latest"
fi  

# Define the path to your program
PROGRAM_PATH="/usr/local/bin/bot_agent"

# URL to download the file
FILE_URL="https://github.com/more-than-code/deploybot-service-agent/releases/download/$VERSION/bot_agent-linux-$(uname -m)"

# Download the file
wget -O "$PROGRAM_PATH" "$FILE_URL"

# Make the downloaded file executable
chmod +x "$PROGRAM_PATH"

# Define the service name and description
SERVICE_NAME="bot_agent"
SERVICE_DESCRIPTION="BotAgent@$VERSION"

# Source the environment file
source /etc/bot/bot.conf

# Create the systemd service file
cat << EOF > "/etc/systemd/system/$SERVICE_NAME.service"
[Unit]
Description=$SERVICE_DESCRIPTION
After=network.target

[Service]
Type=simple
ExecStart=$PROGRAM_PATH start
Restart=on-failure
RestartSec=30s
StartLimitBurst=10
Environment="SERVICE_PORT=$SERVICE_PORT"
Environment="SERVICE_CRT=$SERVICE_CRT"
Environment="SERVICE_KEY=$SERVICE_KEY"
Environment="API_KEY=$API_KEY"
Environment="API_BASE_URL=$API_BASE_URL"
Environment="DOCKER_HOST=$DOCKER_HOST"
Environment="DH_USERNAME=$DH_USERNAME"
Environment="DH_PASSWORD=$DH_PASSWORD"
Environment="REPO_USERNAME=$REPO_USERNAME"
Environment="REPO_PASSWORD=$REPO_PASSWORD"

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd to load the new service file
systemctl daemon-reload

# Start and enable the service
systemctl start $SERVICE_NAME
systemctl enable $SERVICE_NAME

# Display the status of the service
systemctl status $SERVICE_NAME
