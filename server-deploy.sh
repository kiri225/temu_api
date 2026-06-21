#!/usr/bin/env bash
# Temu API Playground 服务器一键部署
# 用法: curl -fsSL https://raw.githubusercontent.com/kiri225/temu_api/master/server-deploy.sh | bash
#   或: bash server-deploy.sh
set -euo pipefail

REPO="${TEMU_REPO:-https://github.com/kiri225/temu_api.git}"
INSTALL_DIR="${INSTALL_DIR:-/opt/temu-api}"
DATA_DIR="${DATA_DIR:-/opt/temu-api-data}"
BRANCH="${BRANCH:-master}"
PLAYGROUND_PORT="${PLAYGROUND_PORT:-8080}"
GIT_CLONE_TIMEOUT="${GIT_CLONE_TIMEOUT:-120}"

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || { echo "缺少命令: $1" >&2; exit 1; }
}

clone_repo() {
  export GIT_TERMINAL_PROMPT=0
  export GIT_HTTP_LOW_SPEED_LIMIT=1000
  export GIT_HTTP_LOW_SPEED_TIME=30
  if [[ "$REPO" == git@* ]]; then
    export GIT_SSH_COMMAND="ssh -o ConnectTimeout=15 -o BatchMode=yes -o StrictHostKeyChecking=accept-new"
  fi
  echo ">> 克隆仓库: $REPO (分支: $BRANCH，超时: ${GIT_CLONE_TIMEOUT}s)"
  mkdir -p "$(dirname "$INSTALL_DIR")"
  if command -v timeout >/dev/null 2>&1; then
    timeout "$GIT_CLONE_TIMEOUT" git clone --progress -b "$BRANCH" "$REPO" "$INSTALL_DIR"
  else
    git clone --progress -b "$BRANCH" "$REPO" "$INSTALL_DIR"
  fi
}

clone_failed_hint() {
  echo ""
  echo "克隆失败。腾讯云/国内服务器常无法直连 GitHub，可尝试：" >&2
  echo "  1) 本机执行 remote-deploy.ps1（从本机同步代码，无需服务器 clone）" >&2
  echo "  2) 使用镜像: TEMU_REPO=https://ghproxy.com/https://github.com/kiri225/temu_api.git bash server-deploy.sh" >&2
  echo "  3) 在服务器配置 HTTP 代理后再 clone" >&2
}

need_cmd git
need_cmd docker
docker compose version >/dev/null 2>&1 || { echo "需要 Docker Compose v2 (docker compose)" >&2; exit 1; }

mkdir -p "$DATA_DIR/config"

if [[ -d "$INSTALL_DIR/.git" ]]; then
  echo ">> 更新代码: $INSTALL_DIR"
  git -C "$INSTALL_DIR" fetch origin "$BRANCH"
  git -C "$INSTALL_DIR" checkout "$BRANCH"
  git -C "$INSTALL_DIR" pull --ff-only origin "$BRANCH"
elif [[ -f "$INSTALL_DIR/docker-compose.yml" ]]; then
  echo ">> 使用已有代码: $INSTALL_DIR (跳过克隆)"
else
  clone_repo || { clone_failed_hint; exit 1; }
fi

if [[ ! -f "$DATA_DIR/config/config.json" ]]; then
  echo ""
  echo "首次部署需要上传配置文件（只需一次）："
  echo "  scp config/config.json user@服务器:$DATA_DIR/config/"
  echo ""
  echo "上传完成后重新运行: bash server-deploy.sh"
  exit 1
fi

if [[ ! -f "$DATA_DIR/unavailable.json" ]]; then
  if [[ -f "$INSTALL_DIR/cmd/playground/unavailable.json" ]]; then
    cp "$INSTALL_DIR/cmd/playground/unavailable.json" "$DATA_DIR/unavailable.json"
  else
    echo '{"byId":{},"byType":{}}' > "$DATA_DIR/unavailable.json"
  fi
fi

export TEMU_CONFIG_PATH="$DATA_DIR/config/config.json"
export TEMU_UNAVAILABLE_PATH="$DATA_DIR/unavailable.json"
export PLAYGROUND_PORT

cd "$INSTALL_DIR"
echo ">> 构建并启动容器..."
docker compose up -d --build

host_ip="$(hostname -I 2>/dev/null | awk '{print $1}')"
host_ip="${host_ip:-localhost}"

echo ""
echo "部署完成"
echo "  本地: http://localhost:${PLAYGROUND_PORT}"
echo "  外网: http://${host_ip}:${PLAYGROUND_PORT}"
echo "  配置: ${TEMU_CONFIG_PATH}"
echo "  代码: ${INSTALL_DIR}"
