package httpapi

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type ServerConfig struct {
	HTTPAddr string
	Logger   *log.Logger
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
