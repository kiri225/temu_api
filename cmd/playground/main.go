package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/goccy/go-json"
	temu "github.com/hiscaler/temu-go"
	"github.com/hiscaler/temu-go/config"
)

//go:embed web/*
var webFS embed.FS

var (
	client      *temu.Client
	cfg         config.Config
	unavailable *unavailableStore
)

func main() {
	port := flag.Int("port", 8080, "监听端口")
	configPath := flag.String("config", "./config/config.json", "配置文件路径")
	unavailablePath := flag.String("unavailable", "./cmd/playground/unavailable.json", "不可用接口标记文件路径")
	flag.Parse()

	if err := loadClient(*configPath); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	var err error
	unavailable, err = newUnavailableStore(*unavailablePath)
	if err != nil {
		log.Fatalf("加载不可用标记失败: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/config", handleConfig)
	mux.HandleFunc("/api/catalog", handleCatalog)
	mux.HandleFunc("/api/unavailable", handleUnavailable)
	mux.HandleFunc("/api/invoke", handleInvoke)

	webRoot, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatalf("加载静态文件失败: %v", err)
	}
	mux.Handle("/", http.FileServer(http.FS(webRoot)))

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Temu API Playground 已启动: http://localhost%s", addr)
	log.Printf("环境: %s | 区域: %s | 调试: %v", cfg.Env, cfg.Region, cfg.Debug)
	log.Printf("不可用标记: %s", *unavailablePath)
	if err := http.ListenAndServe(addr, withCORS(mux)); err != nil {
		log.Fatal(err)
	}
}

func loadClient(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(b, &cfg); err != nil {
		return err
	}
	client = temu.NewClient(cfg)
	return nil
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func maskSecret(s string) string {
	if len(s) <= 8 {
		return strings.Repeat("*", len(s))
	}
	return s[:4] + strings.Repeat("*", len(s)-8) + s[len(s)-4:]
}

func handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"env":          cfg.Env,
		"region":       cfg.Region,
		"debug":        cfg.Debug,
		"timeout":      cfg.Timeout,
		"verify_ssl":   cfg.VerifySSL,
		"app_key":      maskSecret(cfg.AppKey),
		"app_secret":   maskSecret(cfg.AppSecret),
		"access_token": maskSecret(cfg.AccessToken),
	})
}

func handleCatalog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"categories":  categoryOrder,
		"apis":        applyUnavailableFlags(apiCatalog, unavailable),
		"unavailable": unavailable.snapshot(),
	})
}

func handleUnavailable(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, unavailable.snapshot())
	case http.MethodPost:
		var req unavailableUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求体格式错误: " + err.Error()})
			return
		}
		req.ID = strings.TrimSpace(req.ID)
		req.Type = strings.TrimSpace(req.Type)
		if req.ID == "" && req.Type == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id 或 type 至少填一个"})
			return
		}
		if err := unavailable.apply(req); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":          true,
			"unavailable": unavailable.snapshot(),
		})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

type invokeRequest struct {
	Type string          `json:"type"`
	Body json.RawMessage `json:"body"`
}

func handleInvoke(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req invokeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求体格式错误: " + err.Error()})
		return
	}
	req.Type = strings.TrimSpace(req.Type)
	if req.Type == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "type 不能为空"})
		return
	}

	body := map[string]any{}
	if len(req.Body) > 0 && string(req.Body) != "null" {
		if err := json.Unmarshal(req.Body, &body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "body JSON 格式错误: " + err.Error()})
			return
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(cfg.Timeout)*time.Second)
	defer cancel()

	result, err := client.Invoke(ctx, req.Type, body)
	resp := map[string]any{
		"type": req.Type,
		"ok":   err == nil,
	}
	if result != nil {
		resp["statusCode"] = result.StatusCode
		resp["durationMs"] = result.DurationMs
		var parsed any
		if json.Unmarshal(result.Body, &parsed) == nil {
			resp["data"] = parsed
		} else {
			resp["data"] = string(result.Body)
		}
	}
	if err != nil {
		resp["error"] = err.Error()
		writeJSON(w, http.StatusOK, resp)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}
