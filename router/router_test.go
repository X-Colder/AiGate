package router

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/aigate/config"
	"github.com/aigate/pkg/logger"
	"github.com/aigate/pkg/response"
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

// =============================================================
// API 集成测试：GET /health
// =============================================================

func TestAPI_Health_OK(t *testing.T) {
	r := Setup(testConfig())

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", resp["status"])
	}
	if resp["service"] != "AiGate" {
		t.Errorf("expected service 'AiGate', got %v", resp["service"])
	}
}

func TestAPI_Health_MethodNotAllowed(t *testing.T) {
	r := Setup(testConfig())

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/health", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound && w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 404 or 405, got %d", w.Code)
	}
}

// =============================================================
// API 集成测试：GET /api/v1/providers
// =============================================================

func TestAPI_Providers_EmptyList(t *testing.T) {
	r := Setup(testConfig()) // openai disabled

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/providers", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp response.R
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 0 {
		t.Errorf("expected code 0, got %d", resp.Code)
	}
	if resp.Message != "success" {
		t.Errorf("expected message 'success', got %s", resp.Message)
	}
}

func TestAPI_Providers_WithEnabled(t *testing.T) {
	cfg := testConfig()
	cfg.Providers["openai"] = config.ProviderConfig{
		Enabled: true,
		APIKey:  "test-key",
		BaseURL: "https://api.openai.com/v1",
		Model:   "gpt-4",
		Timeout: 30,
	}
	r := Setup(cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/providers", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp response.R
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 0 {
		t.Errorf("expected code 0, got %d", resp.Code)
	}
	// data 应该是非空数组
	dataList, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatalf("expected data to be array, got %T", resp.Data)
	}
	if len(dataList) != 1 {
		t.Errorf("expected 1 provider, got %d", len(dataList))
	}
}

func TestAPI_Providers_CORS_Headers(t *testing.T) {
	r := Setup(testConfig())

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/providers", nil)
	r.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("expected CORS Allow-Origin header")
	}
}

// =============================================================
// API 集成测试：POST /api/v1/chat
// =============================================================

func TestAPI_Chat_EmptyBody(t *testing.T) {
	r := Setup(testConfig())

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/chat", nil)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	var resp response.R
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != -1 {
		t.Errorf("expected code -1, got %d", resp.Code)
	}
}

func TestAPI_Chat_InvalidJSON(t *testing.T) {
	r := Setup(testConfig())

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/chat", bytes.NewReader([]byte("{invalid")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestAPI_Chat_MissingProvider(t *testing.T) {
	r := Setup(testConfig())

	body := map[string]interface{}{
		"messages": []map[string]string{{"role": "user", "content": "hi"}},
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/chat", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	var resp response.R
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != -1 {
		t.Errorf("expected code -1, got %d", resp.Code)
	}
}

func TestAPI_Chat_MissingMessages(t *testing.T) {
	r := Setup(testConfig())

	body := map[string]interface{}{
		"provider": "openai",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/chat", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestAPI_Chat_ProviderNotFound(t *testing.T) {
	r := Setup(testConfig()) // openai disabled, not registered

	body := map[string]interface{}{
		"provider": "openai",
		"messages": []map[string]string{{"role": "user", "content": "hi"}},
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/chat", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}

	var resp response.R
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != -1 {
		t.Errorf("expected code -1, got %d", resp.Code)
	}
}

func TestAPI_Chat_MethodNotAllowed(t *testing.T) {
	r := Setup(testConfig())

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/chat", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound && w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 404 or 405, got %d", w.Code)
	}
}

// =============================================================
// API 集成测试：404 路由
// =============================================================

func TestAPI_NotFound(t *testing.T) {
	r := Setup(testConfig())

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/nonexistent", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestAPI_NotFound_V1(t *testing.T) {
	r := Setup(testConfig())

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/nonexistent", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

// =============================================================
// API 集成测试：OPTIONS 预检请求 (CORS)
// =============================================================

func TestAPI_Providers_AllDomestic(t *testing.T) {
	cfg := testConfig()
	// 启用四个国内 provider
	cfg.Providers["deepseek"] = config.ProviderConfig{Enabled: true, Timeout: 60}
	cfg.Providers["doubao"] = config.ProviderConfig{Enabled: true, Timeout: 60}
	cfg.Providers["qwen"] = config.ProviderConfig{Enabled: true, Timeout: 60}
	cfg.Providers["kimi"] = config.ProviderConfig{Enabled: true, Timeout: 60}
	r := Setup(cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/providers", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp response.R
	json.Unmarshal(w.Body.Bytes(), &resp)
	dataList, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatalf("expected data to be array, got %T", resp.Data)
	}
	if len(dataList) != 4 {
		t.Errorf("expected 4 providers, got %d", len(dataList))
	}
}

func TestAPI_Chat_DeepSeekNotRegistered(t *testing.T) {
	r := Setup(testConfig()) // deepseek not in config

	body := map[string]interface{}{
		"provider": "deepseek",
		"messages": []map[string]string{{"role": "user", "content": "hi"}},
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/chat", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestAPI_CORS_Preflight(t *testing.T) {
	r := Setup(testConfig())

	w := httptest.NewRecorder()
	req := httptest.NewRequest("OPTIONS", "/api/v1/chat", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	r.ServeHTTP(w, req)

	if w.Code != 204 {
		t.Errorf("expected 204, got %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("expected CORS Allow-Origin *")
	}
	if w.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("expected CORS Allow-Methods header")
	}
	if w.Header().Get("Access-Control-Allow-Headers") == "" {
		t.Error("expected CORS Allow-Headers header")
	}
}
