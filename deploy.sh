#!/usr/bin/env bash
set -euo pipefail

LOCAL_DIR="/Users/arseny/Documents/VsCode/vk-forwarder"
REMOTE_HOST="root@192.168.1.167"
REMOTE_DIR="/root/vk-forwarder"
IMAGE_NAME="vk-forwarder:latest"
CONTAINER_NAME="vk-forwarder"
PORT="14888"

rsync -av --delete --exclude ".git" --exclude ".vscode" --exclude "tmp" --exclude "__debug_bin*" "$LOCAL_DIR/" "$REMOTE_HOST:$REMOTE_DIR"

ssh "$REMOTE_HOST" "cd '$REMOTE_DIR' && docker build -t '$IMAGE_NAME' ."

ssh "$REMOTE_HOST" "docker rm -f '$CONTAINER_NAME' 2>/dev/null || true"
ssh "$REMOTE_HOST" "docker run -d --name '$CONTAINER_NAME' --env-file '$REMOTE_DIR/.env' -p $PORT:$PORT --restart unless-stopped '$IMAGE_NAME'"

echo "Deploy complete: $CONTAINER_NAME on $REMOTE_HOST:$PORT"
