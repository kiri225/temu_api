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
  git_http_env
  echo ">> 克隆仓库: $REPO (分支: $BRANCH，超时: ${GIT_CLONE_TIMEOUT}s)"
  mkdir -p "$(dirname "$INSTALL_DIR")"
  if command -v timeout >/dev/null 2>&1; then
    timeout "$GIT_CLONE_TIMEOUT" git clone --depth 1 --progress -b "$BRANCH" "$REPO" "$INSTALL_DIR"
  else
    git clone --depth 1 --progress -b "$BRANCH" "$REPO" "$INSTALL_DIR"
  fi
}

clone_failed_hint() {
  echo ""
  echo "克隆失败。腾讯云/国内服务器常无法直连 GitHub，可尝试：" >&2
  echo "  1) 本机执行 remote-deploy.ps1（从本机同步代码，无需服务器 clone）" >&2
  echo "  2) 使用镜像: TEMU_REPO=https://ghproxy.com/https://github.com/kiri225/temu_api.git bash server-deploy.sh" >&2
  echo "  3) 在服务器配置 HTTP 代理后再 clone" >&2
  echo "  4) 已手动 clone 后: SKIP_GIT_UPDATE=1 bash server-deploy.sh" >&2
}

git_http_env() {
  export GIT_TERMINAL_PROMPT=0
  export GIT_HTTP_LOW_SPEED_LIMIT=1000
  export GIT_HTTP_LOW_SPEED_TIME=30
  if [[ "$REPO" == git@* ]]; then
    export GIT_SSH_COMMAND="ssh -o ConnectTimeout=15 -o BatchMode=yes -o StrictHostKeyChecking=accept-new"
  fi
}

git_with_timeout() {
  git_http_env
  if command -v timeout >/dev/null 2>&1; then
    timeout "$GIT_CLONE_TIMEOUT" git "$@"
  else
    git "$@"
  fi
}

update_repo() {
  echo ">> 更新代码: $INSTALL_DIR (超时: ${GIT_CLONE_TIMEOUT}s)"
  git_with_timeout -C "$INSTALL_DIR" fetch origin "$BRANCH"
  git_with_timeout -C "$INSTALL_DIR" checkout "$BRANCH"
  git_with_timeout -C "$INSTALL_DIR" pull --ff-only origin "$BRANCH"
}

need_cmd git
need_cmd docker
docker compose version >/dev/null 2>&1 || { echo "需要 Docker Compose v2 (docker compose)" >&2; exit 1; }

mkdir -p "$DATA_DIR/config"

if [[ -d "$INSTALL_DIR/.git" ]]; then
  if [[ "${SKIP_GIT_UPDATE:-}" == "1" ]]; then
    echo ">> 跳过 git 更新 (SKIP_GIT_UPDATE=1)"
  elif ! update_repo; then
    echo ">> git 更新失败或超时，使用当前代码继续部署" >&2
  fi
elif [[ -f "$INSTALL_DIR/docker-compose.yml" ]]; then
  echo ">> 使用已有代码: $INSTALL_DIR (跳过克隆)"
else
  if clone_repo; then
    :
  else
    echo ">> git 克隆失败，尝试 tarball 下载..."
    mkdir -p "$INSTALL_DIR"
    TARBALL="https://github.com/kiri225/temu_api/archive/refs/heads/master.tar.gz"
    MIRROR="https://ghproxy.com/https://github.com/kiri225/temu_api/archive/refs/heads/master.tar.gz"
    if ! curl -fsSL --connect-timeout 30 --max-time 600 "$TARBALL" | tar -xz -C "$INSTALL_DIR" --strip-components=1; then
      curl -fsSL --connect-timeout 30 --max-time 600 "$MIRROR" | tar -xz -C "$INSTALL_DIR" --strip-components=1
    fi
    if [[ ! -f "$INSTALL_DIR/docker-compose.yml" ]]; then
      clone_failed_hint
      exit 1
    fi
  fi
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

# 容器内以非 root 用户 (uid 65532) 运行
chmod 644 "$DATA_DIR/config/config.json"
chown 65532:65532 "$DATA_DIR/unavailable.json"
chmod 644 "$DATA_DIR/unavailable.json"

if [[ -f "$DATA_DIR/playground.env" ]]; then
  echo ">> 加载登录配置: $DATA_DIR/playground.env"
  set -a
  # shellcheck disable=SC1090
  source "$DATA_DIR/playground.env"
  set +a
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
