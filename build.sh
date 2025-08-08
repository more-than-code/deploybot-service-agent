#!/bin/bash

VERSION=$1

if [ -z "$VERSION" ]; then
  printf "Version not provided.\n"
  exit 1
fi

# Clean previous builds
echo "Cleaning previous builds..."
rm -f bot_agent-linux-* *.sha256

# Build for different architectures
echo "Building bot_agent version $VERSION..."

# Build for x86_64 (amd64)
echo "Building for linux/x86_64..."
CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags "-X main.Version=$VERSION -w -s" -o bot_agent-linux-x86_64
if [ $? -ne 0 ]; then
  echo "Failed to build x86_64 binary"
  exit 1
fi

# Build for aarch64 (arm64)  
echo "Building for linux/aarch64..."
CGO_ENABLED=0 GOARCH=arm64 GOOS=linux go build -ldflags "-X main.Version=$VERSION -w -s" -o bot_agent-linux-aarch64
if [ $? -ne 0 ]; then
  echo "Failed to build aarch64 binary"
  exit 1
fi

# Make binaries executable
chmod +x bot_agent-linux-*

# Verify binaries were created
if [ ! -f "bot_agent-linux-x86_64" ] || [ ! -f "bot_agent-linux-aarch64" ]; then
  echo "Error: One or more binaries were not created successfully"
  exit 1
fi

# Generate SHA256 checksums
echo "Generating SHA256 checksums..."
sha256sum bot_agent-linux-x86_64 > bot_agent-linux-x86_64.sha256
sha256sum bot_agent-linux-aarch64 > bot_agent-linux-aarch64.sha256

echo "Build complete!"
echo "Files generated:"
echo "  - bot_agent-linux-x86_64"
echo "  - bot_agent-linux-x86_64.sha256"
echo "  - bot_agent-linux-aarch64"
echo "  - bot_agent-linux-aarch64.sha256"

# Verify checksums
echo ""
echo "Verifying checksums..."
sha256sum -c bot_agent-linux-x86_64.sha256
sha256sum -c bot_agent-linux-aarch64.sha256

# Optional: Sign binaries if signing key is available
if [ -n "$SIGNING_KEY" ] && [ -f "$SIGNING_KEY" ]; then
  echo ""
  echo "Signing binaries with GPG..."
  gpg --detach-sign --armor --local-user "$SIGNING_KEY" bot_agent-linux-x86_64
  gpg --detach-sign --armor --local-user "$SIGNING_KEY" bot_agent-linux-aarch64
  echo "Generated signatures:"
  echo "  - bot_agent-linux-x86_64.asc"
  echo "  - bot_agent-linux-aarch64.asc"
fi

echo ""
echo "Build summary for version $VERSION:"
ls -la bot_agent-linux-* | awk '{print "  " $9 " (" $5 " bytes)"}'
