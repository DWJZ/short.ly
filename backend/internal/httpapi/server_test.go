package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
	srv := NewServer(ServerConfig{HTTPAddr: "127.0.0.1:0"})
	ts := httptest.NewServer(srv.Handler)
	t.Cleanup(ts.Close)

	resp, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("GET /health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestCreateAndVisit(t *testing.T) {
	srv := NewServer(ServerConfig{HTTPAddr: "127.0.0.1:0"})
	ts := httptest.NewServer(srv.Handler)
	t.Cleanup(ts.Close)

	// Create short URL.
	createBody := []byte(`{"original_url":"https://example.com/path?q=1"}`)
	resp, err := http.Post(ts.URL+"/short_url", "application/json", bytes.NewReader(createBody))
	if err != nil {
		t.Fatalf("POST /short_url: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var createResp struct {
		Code string `json:"short_url_code"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if createResp.Code == "" {
		t.Fatalf("short_url_code is empty")
	}

	// Visit / resolve.
	getResp, err := http.Get(ts.URL + "/original_url/" + createResp.Code)
	if err != nil {
		t.Fatalf("GET /original_url/{code}: %v", err)
	}
	defer getResp.Body.Close()

	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("GET status = %d, want %d", getResp.StatusCode, http.StatusOK)
	}

	var resolved struct {
		OriginalURL string `json:"original_url"`
	}
	if err := json.NewDecoder(getResp.Body).Decode(&resolved); err != nil {
		t.Fatalf("decode resolve response: %v", err)
	}
	if resolved.OriginalURL != "https://example.com/path?q=1" {
		t.Fatalf("original_url = %q, want %q", resolved.OriginalURL, "https://example.com/path?q=1")
	}
}

func TestCreateMissingOriginalURL(t *testing.T) {
	srv := NewServer(ServerConfig{HTTPAddr: "127.0.0.1:0"})
	ts := httptest.NewServer(srv.Handler)
	t.Cleanup(ts.Close)

	resp, err := http.Post(ts.URL+"/short_url", "application/json", bytes.NewReader([]byte(`{}`)))
	if err != nil {
		t.Fatalf("POST /short_url: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestResolveNotFound(t *testing.T) {
	srv := NewServer(ServerConfig{HTTPAddr: "127.0.0.1:0"})
	ts := httptest.NewServer(srv.Handler)
	t.Cleanup(ts.Close)

	resp, err := http.Get(ts.URL + "/original_url/doesnotexist")
	if err != nil {
		t.Fatalf("GET /original_url/{code}: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

