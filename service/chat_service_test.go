package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/aigate/model"
	"github.com/aigate/provider"
)

// mockProvider 模拟 provider
type mockProvider struct {
	name         string
	chatResp     *model.ChatResponse
	chatErr      error
	streamErr    error
	streamChunks []*model.StreamChunk
}

func (m *mockProvider) Name() string { return m.name }

func (m *mockProvider) Chat(ctx context.Context, req *model.ChatRequest) (*model.ChatResponse, error) {
	if m.chatErr != nil {
		return nil, m.chatErr
	}
	return m.chatResp, nil
}

func (m *mockProvider) ChatStream(ctx context.Context, req *model.ChatRequest, callback func(chunk *model.StreamChunk) error) error {
	if m.streamErr != nil {
		return m.streamErr
	}
	for _, chunk := range m.streamChunks {
		if err := callback(chunk); err != nil {
			return err
		}
	}
	return nil
}

func setupService() (*ChatService, *mockProvider) {
	registry := provider.NewRegistry()
	mock := &mockProvider{
		name: "test",
		chatResp: &model.ChatResponse{
			ID:       "resp-1",
			Provider: "test",
			Content:  "Hello from mock",
		},
	}
	registry.Register(mock)
	svc := NewChatService(registry, nil, nil)
	return svc, mock
}

func TestNewChatService(t *testing.T) {
	registry := provider.NewRegistry()
	svc := NewChatService(registry, nil, nil)
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestChat_Success(t *testing.T) {
	svc, _ := setupService()

	req := &model.ChatRequest{
		Provider: "test",
		Messages: []model.Message{{Role: "user", Content: "hi"}},
	}

	resp, err := svc.Chat(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "Hello from mock" {
		t.Errorf("expected 'Hello from mock', got %s", resp.Content)
	}
	if resp.Provider != "test" {
		t.Errorf("expected provider 'test', got %s", resp.Provider)
	}
}

func TestChat_ProviderNotFound(t *testing.T) {
	svc, _ := setupService()

	req := &model.ChatRequest{
		Provider: "nonexistent",
		Messages: []model.Message{{Role: "user", Content: "hi"}},
	}

	_, err := svc.Chat(context.Background(), req)
	if err == nil {
		t.Error("expected error for nonexistent provider")
	}
}

func TestChat_ProviderError(t *testing.T) {
	svc, mock := setupService()
	mock.chatErr = fmt.Errorf("api error")

	req := &model.ChatRequest{
		Provider: "test",
		Messages: []model.Message{{Role: "user", Content: "hi"}},
	}

	_, err := svc.Chat(context.Background(), req)
	if err == nil {
		t.Error("expected error from provider")
	}
}

func TestChatStream_Success(t *testing.T) {
	svc, mock := setupService()
	mock.streamChunks = []*model.StreamChunk{
		{ID: "chunk-1", Delta: "Hello", Done: false},
		{ID: "chunk-2", Delta: " World", Done: true},
	}

	req := &model.ChatRequest{
		Provider: "test",
		Stream:   true,
		Messages: []model.Message{{Role: "user", Content: "hi"}},
	}

	var chunks []*model.StreamChunk
	err := svc.ChatStream(context.Background(), req, func(chunk *model.StreamChunk) error {
		chunks = append(chunks, chunk)
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chunks) != 2 {
		t.Errorf("expected 2 chunks, got %d", len(chunks))
	}
}

func TestChatStream_ProviderNotFound(t *testing.T) {
	svc, _ := setupService()

	req := &model.ChatRequest{
		Provider: "nonexistent",
		Stream:   true,
		Messages: []model.Message{{Role: "user", Content: "hi"}},
	}

	err := svc.ChatStream(context.Background(), req, func(chunk *model.StreamChunk) error {
		return nil
	})
	if err == nil {
		t.Error("expected error for nonexistent provider")
	}
}

func TestListProviders(t *testing.T) {
	svc, _ := setupService()
	list := svc.ListProviders()
	if len(list) != 1 {
		t.Errorf("expected 1 provider, got %d", len(list))
	}
	if list[0].Name != "test" {
		t.Errorf("expected provider name 'test', got %s", list[0].Name)
	}
}
