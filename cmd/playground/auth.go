package main

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/goccy/go-json"
)

const sessionCookie = "playground_session"

type authenticator struct {
	enabled  bool
	username string
	password string
	ttl      time.Duration
	mu       sync.RWMutex
	sessions map[string]time.Time
}

func newAuthenticator() *authenticator {
	user := strings.TrimSpace(os.Getenv("PLAYGROUND_AUTH_USER"))
	pass := os.Getenv("PLAYGROUND_AUTH_PASSWORD")
	a := &authenticator{
		enabled:  user != "" && pass != "",
		username: user,
		password: pass,
		ttl:      7 * 24 * time.Hour,
		sessions: map[string]time.Time{},
	}
	if a.enabled {
		log.Printf("Playground 登录验证已启用，用户: %s", user)
	}
	return a
}

func (a *authenticator) isPublic(path, method string) bool {
	switch {
	case path == "/health":
		return true
	case path == "/login.html" && method == http.MethodGet:
		return true
	case path == "/style.css" && method == http.MethodGet:
		return true
	case path == "/api/auth/login" && method == http.MethodPost:
		return true
	case path == "/api/auth/me" && method == http.MethodGet:
		return true
	default:
		return false
	}
}

func (a *authenticator) sessionToken(r *http.Request) string {
	c, err := r.Cookie(sessionCookie)
	if err != nil || c.Value == "" {
		return ""
	}
	return c.Value
}

func (a *authenticator) authenticated(r *http.Request) bool {
	if !a.enabled {
		return true
	}
	token := a.sessionToken(r)
	if token == "" {
		return false
	}
	a.mu.RLock()
	expires, ok := a.sessions[token]
	a.mu.RUnlock()
	if !ok || time.Now().After(expires) {
		a.revoke(token)
		return false
	}
	return true
}

func (a *authenticator) createSession() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)
	a.mu.Lock()
	a.sessions[token] = time.Now().Add(a.ttl)
	a.mu.Unlock()
	return token, nil
}

func (a *authenticator) revoke(token string) {
	if token == "" {
		return
	}
	a.mu.Lock()
	delete(a.sessions, token)
	a.mu.Unlock()
}

func (a *authenticator) setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(a.ttl.Seconds()),
	})
}

func (a *authenticator) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
}

func (a *authenticator) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.enabled || a.isPublic(r.URL.Path, r.Method) {
			next.ServeHTTP(w, r)
			return
		}
		if a.authenticated(r) {
			next.ServeHTTP(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/") {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "未登录"})
			return
		}
		http.Redirect(w, r, "/login.html", http.StatusFound)
	})
}

func (a *authenticator) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !a.enabled {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "auth": false})
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求体格式错误"})
		return
	}
	if strings.TrimSpace(req.Username) != a.username || req.Password != a.password {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "账号或密码错误"})
		return
	}
	token, err := a.createSession()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "创建会话失败"})
		return
	}
	a.setSessionCookie(w, token)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (a *authenticator) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	a.revoke(a.sessionToken(r))
	a.clearSessionCookie(w)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (a *authenticator) handleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"authenticated": a.authenticated(r),
		"authEnabled":   a.enabled,
		"username":      a.username,
	})
}
