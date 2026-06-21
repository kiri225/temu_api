$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot

$configPath = if ($env:TEMU_CONFIG_PATH) { $env:TEMU_CONFIG_PATH } else { "config\config.json" }
$unavailPath = if ($env:TEMU_UNAVAILABLE_PATH) { $env:TEMU_UNAVAILABLE_PATH } else { "cmd\playground\unavailable.json" }

if (-not (Test-Path $configPath)) {
    Write-Error "缺少配置文件: $configPath"
}

if (-not (Test-Path $unavailPath)) {
    New-Item -ItemType Directory -Force -Path (Split-Path $unavailPath) | Out-Null
    '{"byId":{},"byType":{}}' | Set-Content $unavailPath -Encoding UTF8
}

$env:TEMU_CONFIG_PATH = (Resolve-Path $configPath).Path
$env:TEMU_UNAVAILABLE_PATH = (Resolve-Path $unavailPath).Path

docker compose up -d --build
if ($LASTEXITCODE -ne 0) {
    Write-Error "Docker 部署失败，请确认 Docker Desktop 已启动"
}

$port = if ($env:PLAYGROUND_PORT) { $env:PLAYGROUND_PORT } else { "8080" }
Write-Host ""
Write-Host "Temu API Playground 已启动: http://localhost:$port" -ForegroundColor Green
