package service

import (
	"context"
	"fmt"

	"github.com/aigate/model"
	"github.com/aigate/provider"
)

// ChatService 聊天服务
type ChatService struct {
	registry *provider.Registry
}

// NewChatService 创建聊天服务
func NewChatService(registry *provider.Registry) *ChatService {
	return &ChatService{
		registry: registry,
	}
}

// Chat 处理聊天请求
func (s *ChatService) Chat(ctx context.Context, req *model.ChatRequest) (*model.ChatResponse, error) {
	p, err := s.registry.Get(req.Provider)
	if err != nil {
		return nil, fmt.Errorf("get provider error: %w", err)
	}
	return p.Chat(ctx, req)
}

// ChatStream 处理流式聊天请求
func (s *ChatService) ChatStream(ctx context.Context, req *model.ChatRequest, callback func(chunk *model.StreamChunk) error) error {
	p, err := s.registry.Get(req.Provider)
	if err != nil {
		return fmt.Errorf("get provider error: %w", err)
	}
	return p.ChatStream(ctx, req, callback)
}

// ListProviders 列出所有可用的提供者
func (s *ChatService) ListProviders() []model.ProviderInfo {
	return s.registry.List()
}
