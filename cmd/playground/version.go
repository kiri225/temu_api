package main

// 由 Dockerfile / go build -ldflags 注入，用于确认服务器是否已部署新版本。
var (
	buildVersion = "dev"
	buildTime    = "unknown"
)
