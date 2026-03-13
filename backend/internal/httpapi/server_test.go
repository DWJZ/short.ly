package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
)

func baseURL(t *testing.T) string {
	t.Helper()
	url := os.Getenv("TEST_BASE_URL")
	if url == "" {
		url = "http://localhost:8080"
	}
	return strings.TrimRight(url, "/")
}

func TestHealth(t *testing.T) {
	resp, err := http.Get(baseURL(t) + "/health")
	if err != nil {
		t.Skipf("GET /health failed (service not running?): %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestCreateAndVisit(t *testing.T) {
	base := baseURL(t)

	// Create short URL.
	createBody := []byte(`{"original_url":"https://example.com/path?q=1"}`)
	resp, err := http.Post(base+"/short_url", "application/json", bytes.NewReader(createBody))
	if err != nil {
		t.Skipf("POST /short_url failed (service not running?): %v", err)
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
	getResp, err := http.Get(base + "/original_url/" + createResp.Code)
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
	resp, err := http.Post(baseURL(t)+"/short_url", "application/json", bytes.NewReader([]byte(`{}`)))
	if err != nil {
		t.Skipf("POST /short_url failed (service not running?): %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestResolveNotFound(t *testing.T) {
	resp, err := http.Get(baseURL(t) + "/original_url/doesnotexist")
	if err != nil {
		t.Skipf("GET /original_url/{code} failed (service not running?): %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

