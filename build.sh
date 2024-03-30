#!/bin/bash

VERSION=$1

if [ -z "$VERSION" ]; then
  printf "Version not provided.\n"
  exit 1
fi

GOARCH=amd64 GOOS=linux go build -ldflags "-X main.Version=$VERSION" -o bot_agent-linux-x86_64; GOARCH=arm64 GOOS=linux go build -ldflags "-X main.Version=$VERSION" -o bot_agent-linux-aarch64
