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

	// Visit / resolve via redirect.
	getResp, err := http.Get(base + "/" + createResp.Code)
	if err != nil {
		t.Fatalf("GET /original_url/{code}: %v", err)
	}
	defer getResp.Body.Close()

	if getResp.StatusCode != http.StatusMovedPermanently {
		t.Fatalf("GET status = %d, want %d", getResp.StatusCode, http.StatusMovedPermanently)
	}

	loc, err := getResp.Location()
	if err != nil {
		t.Fatalf("Location header missing or invalid: %v", err)
	}
	if loc.String() != "https://example.com/path?q=1" {
		t.Fatalf("Location = %q, want %q", loc.String(), "https://example.com/path?q=1")
	}
}

func TestCreateSameOriginalURLReturnsSameCode(t *testing.T) {
	base := baseURL(t)

	body := []byte(`{"original_url":"https://example.com/same"}`)

	resp1, err := http.Post(base+"/short_url", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Skipf("POST /short_url failed (service not running?): %v", err)
	}
	defer resp1.Body.Close()

	if resp1.StatusCode != http.StatusOK {
		t.Fatalf("first POST status = %d, want %d", resp1.StatusCode, http.StatusOK)
	}

	var r1 struct {
		Code string `json:"short_url_code"`
	}
	if err := json.NewDecoder(resp1.Body).Decode(&r1); err != nil {
		t.Fatalf("decode first create response: %v", err)
	}
	if r1.Code == "" {
		t.Fatalf("first short_url_code is empty")
	}

	resp2, err := http.Post(base+"/short_url", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Skipf("second POST /short_url failed (service not running?): %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("second POST status = %d, want %d", resp2.StatusCode, http.StatusOK)
	}

	var r2 struct {
		Code string `json:"short_url_code"`
	}
	if err := json.NewDecoder(resp2.Body).Decode(&r2); err != nil {
		t.Fatalf("decode second create response: %v", err)
	}
	if r2.Code != r1.Code {
		t.Fatalf("second code = %q, want same as first %q", r2.Code, r1.Code)
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

func TestCreateInvalidURL(t *testing.T) {
	resp, err := http.Post(baseURL(t)+"/short_url", "application/json", bytes.NewReader([]byte(`{"original_url":"ftp://example.com"}`)))
	if err != nil {
		t.Skipf("POST /short_url failed (service not running?): %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestResolveNotFound(t *testing.T) {
	resp, err := http.Get(baseURL(t) + "/doesnotexist")
	if err != nil {
		t.Skipf("GET /original_url/{code} failed (service not running?): %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

