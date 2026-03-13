package httpapi

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tianmuwu/short.ly/internal/storage"
)

type ServerConfig struct {
	HTTPAddr string
	Logger   *log.Logger
	Repo     storage.Repo
}

func NewServer(cfg ServerConfig) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})

	// short_url
	//   POST: /short_url
	//     body: {original_url: string}
	mux.HandleFunc("/short_url", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if cfg.Repo == nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "repo_not_configured"})
			return
		}

		var req struct {
			OriginalURL string `json:"original_url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_json"})
			return
		}

		original := strings.TrimSpace(req.OriginalURL)
		if original == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing_original_url"})
			return
		}

		parsed, err := url.Parse(original)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_url"})
			return
		}
		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_url_scheme"})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		code, err := cfg.Repo.CreateShortURL(ctx, original)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "create_failed"})
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"short_url_code": code})
	})

	// redirect
	//   GET: /{short_url_code}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if cfg.Repo == nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "repo_not_configured"})
			return
		}

		code := strings.Trim(strings.TrimPrefix(r.URL.Path, "/"), " ")
		if code == "" || strings.Contains(code, "/") {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_short_url_code"})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		original, found, err := cfg.Repo.ResolveOriginalURL(ctx, code)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "resolve_failed"})
			return
		}
		if !found {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "not_found"})
			return
		}

		http.Redirect(w, r, original, http.StatusMovedPermanently)
	})

	return &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           loggingMiddleware(cfg.Logger, mux),
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func loggingMiddleware(l *log.Logger, next http.Handler) http.Handler {
	if l == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		l.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start).String())
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
