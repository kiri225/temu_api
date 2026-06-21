param(
    [Parameter(Mandatory = $true)]
    [string]$Server,

    [string]$InstallDir = "/opt/temu-api",
    [string]$DataDir = "/opt/temu-api-data",
    [string]$Port = "8080"
)

$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot

if (-not (Test-Path "config\config.json")) {
    Write-Error "本地缺少 config\config.json"
}

Write-Host ">> 上传配置到服务器..." -ForegroundColor Cyan
ssh $Server "mkdir -p $DataDir/config"
scp config\config.json "${Server}:${DataDir}/config/config.json"

if (Test-Path "cmd\playground\unavailable.json") {
    scp cmd\playground\unavailable.json "${Server}:${DataDir}/unavailable.json"
}

Write-Host ">> 同步代码到服务器（绕过 GitHub 直连）..." -ForegroundColor Cyan
ssh $Server "mkdir -p $InstallDir"
git archive --format=tar HEAD | ssh $Server "tar -x -C $InstallDir"

Write-Host ">> 在服务器执行部署..." -ForegroundColor Cyan
ssh $Server @"
set -e
export INSTALL_DIR='$InstallDir'
export DATA_DIR='$DataDir'
export PLAYGROUND_PORT='$Port'
bash `$INSTALL_DIR/server-deploy.sh
"@

Write-Host ""
Write-Host "部署完成，访问: http://${Server.Split('@')[-1]}:${Port}" -ForegroundColor Green
