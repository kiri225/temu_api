#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"

export TEMU_CONFIG_PATH="${TEMU_CONFIG_PATH:-./config/config.json}"
export TEMU_UNAVAILABLE_PATH="${TEMU_UNAVAILABLE_PATH:-./cmd/playground/unavailable.json}"

if [[ ! -f "$TEMU_CONFIG_PATH" ]]; then
  echo "缺少配置文件: $TEMU_CONFIG_PATH" >&2
  exit 1
fi

if [[ ! -f "$TEMU_UNAVAILABLE_PATH" ]]; then
  mkdir -p "$(dirname "$TEMU_UNAVAILABLE_PATH")"
  echo '{"byId":{},"byType":{}}' > "$TEMU_UNAVAILABLE_PATH"
fi

docker compose up -d --build

port="${PLAYGROUND_PORT:-8080}"
echo ""
echo "Temu API Playground 已启动: http://localhost:${port}"
