#!/usr/bin/env bash
# 不依赖 git 的完整重部署（适合腾讯云 HTTPS 不稳定场景）
# 用法: curl -fsSL .../server-redeploy.sh | bash
#   或: PLAYGROUND_PORT=27789 bash server-redeploy.sh
set -euo pipefail

cd /

DATA_DIR="${DATA_DIR:-/opt/temu-api-data}"
INSTALL_DIR="${INSTALL_DIR:-/opt/temu-api}"
PLAYGROUND_PORT="${PLAYGROUND_PORT:-27789}"
TARBALL_URL="${TARBALL_URL:-https://github.com/kiri225/temu_api/archive/refs/heads/master.tar.gz}"
TARBALL_MIRROR="${TARBALL_MIRROR:-https://ghproxy.com/https://github.com/kiri225/temu_api/archive/refs/heads/master.tar.gz}"

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || { echo "缺少命令: $1" >&2; exit 1; }
}

need_cmd docker
need_cmd curl
docker compose version >/dev/null 2>&1 || { echo "需要 Docker Compose v2" >&2; exit 1; }

echo ">> [1/5] 检查配置"
mkdir -p "$DATA_DIR/config"
if [[ ! -f "$DATA_DIR/config/config.json" ]]; then
  echo "缺少 $DATA_DIR/config/config.json" >&2
  echo "本机执行: scp config/config.json root@服务器:$DATA_DIR/config/" >&2
  exit 1
fi

if [[ ! -f "$DATA_DIR/playground.env" ]]; then
  echo "缺少 $DATA_DIR/playground.env（登录账号密码）" >&2
  echo "示例见 config/playground.env.example" >&2
  exit 1
fi

if [[ ! -f "$DATA_DIR/unavailable.json" ]]; then
  echo '{"byId":{},"byType":{}}' > "$DATA_DIR/unavailable.json"
fi

chmod 644 "$DATA_DIR/config/config.json"
chown 65532:65532 "$DATA_DIR/unavailable.json"
chmod 644 "$DATA_DIR/unavailable.json"

echo ">> [2/5] 下载代码（tarball，不用 git）"
rm -rf "$INSTALL_DIR"
mkdir -p "$INSTALL_DIR"

fetch_tarball() {
  local url="$1"
  echo "    尝试: $url"
  curl -fsSL --connect-timeout 30 --max-time 600 "$url" | tar -xz -C "$INSTALL_DIR" --strip-components=1
}

if ! fetch_tarball "$TARBALL_URL"; then
  echo ">> 直连失败，改用镜像..."
  fetch_tarball "$TARBALL_MIRROR"
fi

if [[ ! -f "$INSTALL_DIR/docker-compose.yml" ]]; then
  echo "代码下载不完整，缺少 docker-compose.yml" >&2
  exit 1
fi

echo ">> [3/5] 加载环境"
set -a
# shellcheck disable=SC1090
source "$DATA_DIR/playground.env"
set +a
export TEMU_CONFIG_PATH="$DATA_DIR/config/config.json"
export TEMU_UNAVAILABLE_PATH="$DATA_DIR/unavailable.json"
export PLAYGROUND_PORT

echo ">> [4/5] 构建并启动 Docker"
cd "$INSTALL_DIR"
docker compose down 2>/dev/null || true
docker compose up -d --build

echo ">> [5/5] 检查状态"
sleep 5
docker ps --filter name=temu-playground --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}'
echo ""
docker compose logs --tail 20 temu-playground 2>&1 || true

if curl -sf "http://127.0.0.1:${PLAYGROUND_PORT}/health" >/dev/null; then
  echo "health: OK"
else
  echo "health: 失败，请查看上方日志" >&2
  exit 1
fi

host_ip="$(hostname -I 2>/dev/null | awk '{print $1}')"
host_ip="${host_ip:-localhost}"
echo ""
echo "部署完成"
echo "  访问: http://${host_ip}:${PLAYGROUND_PORT}"
echo "  配置: ${TEMU_CONFIG_PATH}"
