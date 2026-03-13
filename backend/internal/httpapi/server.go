package httpapi

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type ServerConfig struct {
	HTTPAddr string
	Logger   *log.Logger
}

func NewServer(cfg ServerConfig) *http.Server {
	mux := http.NewServeMux()
	store := newInMemoryStore()

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

		var req struct {
			OriginalURL string `json:"original_url"`
			TypoURL     string `json:"oritinal_url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_json"})
			return
		}

		original := strings.TrimSpace(req.OriginalURL)
		if original == "" {
			original = strings.TrimSpace(req.TypoURL)
		}
		if original == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing_original_url"})
			return
		}

		code, err := store.Create(original)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "create_failed"})
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"short_url_code": code})
	})

	// original_url
	//   GET: /original_url/{short_url_code}
	mux.HandleFunc("/original_url/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		code := strings.TrimPrefix(r.URL.Path, "/original_url/")
		code = strings.TrimSpace(code)
		if code == "" || strings.Contains(code, "/") {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_short_url_code"})
			return
		}

		original, ok := store.Get(code)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "not_found"})
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"original_url": original})
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

type inMemoryStore struct {
	mu     sync.RWMutex
	byCode map[string]string
}

func newInMemoryStore() *inMemoryStore {
	return &inMemoryStore{byCode: make(map[string]string)}
}

func (s *inMemoryStore) Create(originalURL string) (string, error) {
	for i := 0; i < 5; i++ {
		code, err := newCode(8)
		if err != nil {
			return "", err
		}

		s.mu.Lock()
		if _, exists := s.byCode[code]; !exists {
			s.byCode[code] = originalURL
			s.mu.Unlock()
			return code, nil
		}
		s.mu.Unlock()
	}
	return "", fmt.Errorf("could not allocate unique code")
}

func (s *inMemoryStore) Get(code string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.byCode[code]
	return v, ok
}

const codeAlphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func newCode(n int) (string, error) {
	if n <= 0 {
		return "", fmt.Errorf("invalid code length")
	}

	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("rand read: %w", err)
	}

	for i := range b {
		b[i] = codeAlphabet[int(b[i])%len(codeAlphabet)]
	}
	return string(b), nil
}
