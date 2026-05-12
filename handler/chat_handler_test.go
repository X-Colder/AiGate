package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/aigate/model"
	"github.com/aigate/pkg/logger"
	"github.com/aigate/pkg/response"
	"github.com/aigate/provider"
	"github.com/aigate/service"
)

type mockProvider struct {
	name     string
	chatResp *model.ChatResponse
	chatErr  error
}

func (m *mockProvider) Name() string { return m.name }

func (m *mockProvider) Chat(ctx context.Context, req *model.ChatRequest) (*model.ChatResponse, error) {
	if m.chatErr != nil {
		return nil, m.chatErr
	}
	return m.chatResp, nil
}

func (m *mockProvider) ChatStream(ctx context.Context, req *model.ChatRequest, callback func(chunk *model.StreamChunk) error) error {
	return fmt.Errorf("stream not implemented")
}

func init() {
	gin.SetMode(gin.TestMode)
	logger.Init("error")
}

func setupHandler() (*ChatHandler, *gin.Engine) {
	registry := provider.NewRegistry()
	registry.Register(&mockProvider{
		name: "test",
		chatResp: &model.ChatResponse{
			ID:       "resp-1",
			Provider: "test",
			Model:    "test-model",
			Content:  "Hello!",
			Usage:    &model.Usage{PromptTokens: 5, CompletionTokens: 10, TotalTokens: 15},
		},
	})
	svc := service.NewChatService(registry, nil, nil)
	h := NewChatHandler(svc)

	r := gin.New()
	r.POST("/api/v1/chat", h.Chat)
	r.GET("/api/v1/providers", h.ListProviders)
	return h, r
}

func TestChat_Success(t *testing.T) {
	_, r := setupHandler()

	body := model.ChatRequest{
		Provider: "test",
		Messages: []model.Message{{Role: "user", Content: "hi"}},
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/chat", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp response.R
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.Code != 0 {
		t.Errorf("expected code 0, got %d", resp.Code)
	}
}

func TestChat_InvalidJSON(t *testing.T) {
	_, r := setupHandler()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/chat", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestChat_MissingProvider(t *testing.T) {
	_, r := setupHandler()

	body := map[string]interface{}{
		"messages": []map[string]string{{"role": "user", "content": "hi"}},
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/chat", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestChat_ProviderNotFound(t *testing.T) {
	_, r := setupHandler()

	body := model.ChatRequest{
		Provider: "nonexistent",
		Messages: []model.Message{{Role: "user", Content: "hi"}},
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/chat", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestListProviders(t *testing.T) {
	_, r := setupHandler()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/providers", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp response.R
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.Code != 0 {
		t.Errorf("expected code 0, got %d", resp.Code)
	}
}
