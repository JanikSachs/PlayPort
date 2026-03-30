#!/usr/bin/env bash
set -euo pipefail

SERVICE=playport
BINARY=./playport
BUILD_CMD="go build -o playport ./cmd/playport"

echo "==> Building $SERVICE..."
$BUILD_CMD

echo "==> Restarting $SERVICE.service..."
sudo systemctl restart "$SERVICE"

echo "==> Waiting for service to come up..."
sleep 2

if systemctl is-active --quiet "$SERVICE"; then
    echo "==> $SERVICE is running."
    systemctl status "$SERVICE" --no-pager -l
else
    echo "ERROR: $SERVICE failed to start." >&2
    journalctl -u "$SERVICE" -n 30 --no-pager >&2
    exit 1
fi
