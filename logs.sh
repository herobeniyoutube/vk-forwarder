#!/usr/bin/env bash
set -euo pipefail

REMOTE_HOST="root@192.168.1.167"
CONTAINER_NAME="vk-forwarder"

ssh "$REMOTE_HOST" "docker logs -f --tail=200 $CONTAINER_NAME"
