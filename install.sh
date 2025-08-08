#!/bin/bash

VERSION=$1

if [ -z "$VERSION" ]; then
  VERSION="latest"
fi  

# Pre-installation checks
echo "Performing pre-installation checks..."

# Check if running as root
if [ "$EUID" -ne 0 ]; then
  echo "This script must be run as root (use sudo)"
  exit 1
fi

# Check if required commands exist
REQUIRED_COMMANDS="curl systemctl usermod groups getent logname"
for cmd in $REQUIRED_COMMANDS; do
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "Error: Required command '$cmd' not found"
    exit 1
  fi
done

# Check if Docker is installed
if ! command -v docker >/dev/null 2>&1; then
  echo "Warning: Docker is not installed. Please install Docker first."
  echo "Visit: https://docs.docker.com/engine/install/"
  exit 1
fi

# Check if systemd is available
if ! systemctl --version >/dev/null 2>&1; then
  echo "Error: systemd is required but not available"
  exit 1
fi

echo "✓ Pre-installation checks passed"

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
chmod 755 "$BOT_AGENT_DIR"

# Create or update environment configuration
if [ ! -f "$ENV_FILE" ]; then
  # Create new environment file
  cat << EOF > "$ENV_FILE"
SERVICE_PORT=8002
SERVICE_CRT=/etc/letsencrypt/live/your-domain.com/fullchain.pem
SERVICE_KEY=/etc/letsencrypt/live/your-domain.com/privkey.pem
API_KEY=your_api_key_here
API_BASE_URL=https://api.example.com
DOCKER_HOST=unix:///var/run/docker.sock
DH_USERNAME=your_dockerhub_username
DH_PASSWORD=your_dockerhub_password
REPO_USERNAME=your_repo_username
REPO_PASSWORD=your_repo_password
EOF
  chmod 600 "$ENV_FILE"
  echo "Created environment file at $ENV_FILE"
  echo "Please edit $ENV_FILE and update the configuration values before starting the service."
  echo "NOTE: Update SERVICE_CRT and SERVICE_KEY paths to match your actual domain certbot certificates."
else
  # Merge with existing environment file
  echo "Existing environment file found at $ENV_FILE"
  TEMP_ENV_FILE="$ENV_FILE.tmp"
  
  # Define default values
  declare -A defaults=(
    ["SERVICE_PORT"]="8002"
    ["SERVICE_CRT"]="/etc/letsencrypt/live/your-domain.com/fullchain.pem"
    ["SERVICE_KEY"]="/etc/letsencrypt/live/your-domain.com/privkey.pem"
    ["API_KEY"]="your_api_key_here"
    ["API_BASE_URL"]="https://api.example.com"
    ["DOCKER_HOST"]="unix:///var/run/docker.sock"
    ["DH_USERNAME"]="your_dockerhub_username"
    ["DH_PASSWORD"]="your_dockerhub_password"
    ["REPO_USERNAME"]="your_repo_username"
    ["REPO_PASSWORD"]="your_repo_password"
  )
  
  # Copy existing file to temp
  cp "$ENV_FILE" "$TEMP_ENV_FILE"
  
  # Add missing keys from defaults
  for key in "${!defaults[@]}"; do
    if ! grep -q "^$key=" "$ENV_FILE"; then
      echo "$key=${defaults[$key]}" >> "$TEMP_ENV_FILE"
      echo "Added new configuration: $key"
    fi
  done
  
  # Replace original with merged version
  mv "$TEMP_ENV_FILE" "$ENV_FILE"
  chmod 600 "$ENV_FILE"
  echo "Environment file updated with new configuration options"
fi

# Download the program
echo "Downloading bot_agent from $FILE_URL..."

# Backup existing binary if it exists
if [ -f "$PROGRAM_PATH" ]; then
  BACKUP_PATH="$PROGRAM_PATH.backup.$(date +%Y%m%d_%H%M%S)"
  cp "$PROGRAM_PATH" "$BACKUP_PATH"
  echo "Backed up existing binary to $BACKUP_PATH"
fi

# Download binary
curl -fsSL --retry 3 --retry-delay 5 -o "$PROGRAM_PATH" "$FILE_URL"

if [ $? -ne 0 ]; then
  echo "Failed to download $FILE_URL"
  echo "Please check:"
  echo "1. Internet connectivity"
  echo "2. Release version exists: https://github.com/more-than-code/deploybot-service-agent/releases"
  echo "3. Architecture compatibility: $(uname -m)"
  
  # Restore backup if download failed and backup exists
  if [ -f "$BACKUP_PATH" ]; then
    echo "Restoring previous version from backup..."
    mv "$BACKUP_PATH" "$PROGRAM_PATH"
  fi
  
  exit 1
fi

# Download and verify checksum if available
CHECKSUM_URL="${FILE_URL}.sha256"
CHECKSUM_FILE="/tmp/bot_agent.sha256"

echo "Verifying binary integrity..."
if curl -fsSL --retry 2 --retry-delay 2 -o "$CHECKSUM_FILE" "$CHECKSUM_URL" 2>/dev/null; then
  # Extract expected checksum and filename
  EXPECTED_CHECKSUM=$(cat "$CHECKSUM_FILE" | cut -d' ' -f1)
  
  # Calculate actual checksum
  ACTUAL_CHECKSUM=$(sha256sum "$PROGRAM_PATH" | cut -d' ' -f1)
  
  if [ "$EXPECTED_CHECKSUM" = "$ACTUAL_CHECKSUM" ]; then
    echo "✓ Binary integrity verification passed"
    rm -f "$CHECKSUM_FILE"
  else
    echo "✗ Binary integrity verification failed!"
    echo "Expected: $EXPECTED_CHECKSUM"
    echo "Actual:   $ACTUAL_CHECKSUM"
    
    # Restore backup if verification failed
    if [ -f "$BACKUP_PATH" ]; then
      echo "Restoring previous version from backup..."
      mv "$BACKUP_PATH" "$PROGRAM_PATH"
    else
      rm -f "$PROGRAM_PATH"
    fi
    
    rm -f "$CHECKSUM_FILE"
    exit 1
  fi
else
  echo "⚠ Checksum file not available, skipping integrity verification"
  echo "  (This is normal for development builds)"
fi

echo "Download completed successfully"

# Make the downloaded file executable and secure
chmod 755 "$PROGRAM_PATH"
chown root:root "$PROGRAM_PATH"

# Check if service is currently running
SERVICE_RUNNING=false
if systemctl is-active --quiet bot_agent; then
  SERVICE_RUNNING=true
  echo "Service is currently running, will restart after installation"
fi

# Ensure the installer user is in the docker group
if ! groups "$INSTALLER_USER" | grep -q docker; then
  echo "Adding user $INSTALLER_USER to docker group..."
  usermod -aG docker "$INSTALLER_USER"
  echo "Note: $INSTALLER_USER needs to log out and back in for docker group membership to take effect"
else
  echo "✓ User $INSTALLER_USER is already in docker group"
fi

# Setup certificate access for certbot certificates
if [ -d "/etc/letsencrypt" ]; then
  echo "Setting up certbot certificate access..."
  
  # Add user to ssl-cert group (if it exists) or create it
  if ! getent group ssl-cert >/dev/null 2>&1; then
    groupadd ssl-cert
    echo "Created ssl-cert group"
  fi
  
  # Add installer user to ssl-cert group
  usermod -aG ssl-cert "$INSTALLER_USER"
  
  # Set proper permissions for letsencrypt directories
  chgrp -R ssl-cert /etc/letsencrypt/live/ /etc/letsencrypt/archive/ 2>/dev/null || true
  chmod -R g+rx /etc/letsencrypt/live/ /etc/letsencrypt/archive/ 2>/dev/null || true
  
  echo "✓ Certificate access configured for user $INSTALLER_USER"
  echo "Note: If using certbot certificates, ensure they are readable by the ssl-cert group"
else
  echo "⚠ /etc/letsencrypt not found - certbot may not be installed"
  echo "  Install certbot and generate certificates before starting the service"
fi

# Validate Docker socket access
DOCKER_SOCK="/var/run/docker.sock"
if [ -S "$DOCKER_SOCK" ]; then
  if [ -r "$DOCKER_SOCK" ] && [ -w "$DOCKER_SOCK" ]; then
    echo "✓ Docker socket is accessible"
  else
    echo "⚠ Docker socket permissions may need adjustment"
    echo "  Socket: $DOCKER_SOCK"
    echo "  Current permissions: $(ls -la $DOCKER_SOCK)"
  fi
else
  echo "⚠ Docker socket not found at $DOCKER_SOCK"
  echo "  Make sure Docker is installed and running"
fi

# Set proper ownership and permissions for the bot agent directory
chown -R "$INSTALLER_USER:docker" "$BOT_AGENT_DIR"
chmod -R 755 "$BOT_AGENT_DIR"
chmod 600 "$ENV_FILE"  # Keep config file secure

# Ensure user home directory has proper permissions for service operations
echo "Setting up home directory permissions for service operations..."
USER_HOME_PERMS=$(stat -c "%a" "$USER_HOME")
if [ "$USER_HOME_PERMS" -lt 755 ]; then
  echo "Adjusting home directory permissions to allow service access..."
  chmod 755 "$USER_HOME"
  echo "✓ Home directory permissions updated: $(stat -c "%a" "$USER_HOME")"
else
  echo "✓ Home directory permissions are adequate: $USER_HOME_PERMS"
fi

# Test directory write permissions for bot agent directory
if ! sudo -u "$INSTALLER_USER" test -w "$BOT_AGENT_DIR"; then
  echo "⚠ Warning: $INSTALLER_USER may not have write access to $BOT_AGENT_DIR"
  echo "Attempting to fix permissions..."
  chmod 775 "$BOT_AGENT_DIR"
  if ! sudo -u "$INSTALLER_USER" test -w "$BOT_AGENT_DIR"; then
    echo "✗ Failed to set write permissions for $BOT_AGENT_DIR"
    echo "The service may not be able to write logs or temporary files"
  else
    echo "✓ Bot agent directory write permissions verified"
  fi
else
  echo "✓ Bot agent directory write permissions verified"
fi

# Test home directory write permissions for creating config directories
if ! sudo -u "$INSTALLER_USER" test -w "$USER_HOME"; then
  echo "⚠ Warning: Service may not be able to create config directories in $USER_HOME"
  echo "This is needed for creating SWAG config directories and mounting them"
  echo "Current home directory permissions: $(ls -ld $USER_HOME)"
else
  echo "✓ Home directory write permissions verified for config creation"
fi

# Test creating a config directory structure (simulate SWAG config creation)
TEST_CONFIG_DIR="$USER_HOME/test_config"
if sudo -u "$INSTALLER_USER" mkdir -p "$TEST_CONFIG_DIR" 2>/dev/null; then
  sudo -u "$INSTALLER_USER" rmdir "$TEST_CONFIG_DIR"
  echo "✓ Service can create config directories in home directory"
else
  echo "⚠ Warning: Service cannot create config directories in $USER_HOME"
  echo "This may prevent SWAG configuration from working properly"
  echo "Manual fix: sudo chmod 755 $USER_HOME"
fi

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
Group=docker
EnvironmentFile=$ENV_FILE
WorkingDirectory=$BOT_AGENT_DIR
PrivateTmp=true
ProtectSystem=strict
ProtectHome=false
ReadWritePaths=$BOT_AGENT_DIR $USER_HOME
ReadOnlyPaths=/etc/letsencrypt
NoNewPrivileges=true
# Allow binding to privileged ports if needed
CapabilityBoundingSet=CAP_NET_BIND_SERVICE
# Allow access to docker socket and SSL certificates
SupplementaryGroups=docker ssl-cert

[Install]
WantedBy=multi-user.target
EOF

# Secure the service file
chmod 644 "$SERVICE_FILE"

# Reload systemd and restart service if it was running
systemctl daemon-reload

if [ "$SERVICE_RUNNING" = true ]; then
  echo "Restarting bot_agent service..."
  systemctl restart bot_agent
  if [ $? -eq 0 ]; then
    echo "Service restarted successfully"
  else
    echo "Warning: Service restart failed. Check with: systemctl status bot_agent"
  fi
fi

echo "=============================================="
echo "Bot Agent Installation Complete"
echo "=============================================="
echo "Version: $VERSION"
echo "Binary: $PROGRAM_PATH"
echo "Config: $ENV_FILE"
echo "Service: $SERVICE_FILE"
echo ""

if [ "$SERVICE_RUNNING" = true ]; then
  echo "Service Status: Restarted with new version"
  echo ""
  echo "Check service status: sudo systemctl status bot_agent"
  echo "View logs: sudo journalctl -u bot_agent -f"
else
  echo "Next steps:"
  echo "1. Edit configuration: sudo nano $ENV_FILE"
  echo "2. Update certificate paths in config to match your domain:"
  echo "   SERVICE_CRT=/etc/letsencrypt/live/your-domain.com/fullchain.pem"
  echo "   SERVICE_KEY=/etc/letsencrypt/live/your-domain.com/privkey.pem"
  echo "3. Ensure certbot certificates exist and are accessible"
  echo "4. Enable service: sudo systemctl enable bot_agent"
  echo "5. Start service: sudo systemctl start bot_agent"
  echo "6. Check status: sudo systemctl status bot_agent"
  echo "7. View logs: sudo journalctl -u bot_agent -f"
fi

echo ""
echo "Security Information:"
echo "• Service runs as user: $INSTALLER_USER"
echo "• Service group: docker"
echo "• Working directory: $BOT_AGENT_DIR"
echo "• Directory permissions: $(ls -ld $BOT_AGENT_DIR)"
echo "• Home directory permissions: $(ls -ld $USER_HOME)"
echo "• Config file: $ENV_FILE (readable only by owner)"
echo "• Config file permissions: $(ls -l $ENV_FILE)"
echo "• Binary integrity: $([ -f "/tmp/bot_agent.sha256" ] && echo "Verified" || echo "Skipped")"
echo "• Systemd hardening: Enabled (PrivateTmp, ProtectSystem, etc.)"
echo "• Home directory access: Enabled for config creation (SWAG, etc.)"
echo ""
echo "If the service fails to start, check:"
echo "1. Directory permissions: ls -la $BOT_AGENT_DIR"
echo "2. Home directory permissions: ls -ld $USER_HOME"
echo "3. Service logs: sudo journalctl -u bot_agent -f"
echo "4. User can write to home directory: sudo -u $INSTALLER_USER touch $USER_HOME/test && rm $USER_HOME/test"
echo "5. User can create config directories: sudo -u $INSTALLER_USER mkdir $USER_HOME/test_config && rmdir $USER_HOME/test_config"
echo "=============================================="
