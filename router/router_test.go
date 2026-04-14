package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/aigate/config"
	"github.com/aigate/pkg/logger"
)

func init() {
	gin.SetMode(gin.TestMode)
	logger.Init("error")
}

func testConfig() *config.Config {
	return &config.Config{
		Server: config.ServerConfig{
			Port: "8080",
			Mode: "test",
		},
		LogLevel: "error",
		Providers: map[string]config.ProviderConfig{
			"openai": {
				Enabled: false,
			},
		},
	}
}

func TestSetup(t *testing.T) {
	r := Setup(testConfig())
	if r == nil {
		t.Fatal("expected non-nil engine")
	}
}

func TestHealthEndpoint(t *testing.T) {
	r := Setup(testConfig())

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", resp["status"])
	}
	if resp["service"] != "AiGate" {
		t.Errorf("expected service 'AiGate', got %v", resp["service"])
	}
}

func TestProvidersEndpoint(t *testing.T) {
	r := Setup(testConfig())

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/providers", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestChatEndpoint_NoBody(t *testing.T) {
	r := Setup(testConfig())

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/chat", nil)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestNotFoundRoute(t *testing.T) {
	r := Setup(testConfig())

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/nonexistent", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}
