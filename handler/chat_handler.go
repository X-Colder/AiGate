// Package handler 实现 HTTP 请求处理器层（Controller），
// 负责参数校验、调用 Service 层、格式化响应。
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/aigate/model"
	"github.com/aigate/pkg/logger"
	"github.com/aigate/pkg/response"
	"github.com/aigate/service"
)

// ChatHandler 聊天请求处理器，持有 ChatService 实例
type ChatHandler struct {
	chatService *service.ChatService
}

// NewChatHandler 创建聊天处理器
func NewChatHandler(chatService *service.ChatService) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
	}
}

// Chat 处理 POST /api/v1/chat 聊天请求。
// 根据 stream 字段决定返回普通 JSON 响应或 SSE 流式响应。
func (h *ChatHandler) Chat(c *gin.Context) {
	// 绑定并校验请求参数
	var req model.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// 注入租户 ID（从认证中间件获取）
	if tid, exists := c.Get("tenant_id"); exists {
		req.TenantID, _ = tid.(string)
	}

	// 非流式请求：调用 Service 获取完整响应后一次性返回
	if !req.Stream {
		resp, err := h.chatService.Chat(c.Request.Context(), &req)
		if err != nil {
			logger.Errorf("Chat error: %v", err)
			response.Error(c, http.StatusInternalServerError, "Chat error: "+err.Error())
			return
		}
		response.Success(c, resp)
		return
	}

	// 流式请求：设置 SSE 响应头，通过 callback 逐块推送数据
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	err := h.chatService.ChatStream(c.Request.Context(), &req, func(chunk *model.StreamChunk) error {
		c.SSEvent("message", chunk)
		c.Writer.Flush()
		return nil
	})

	if err != nil {
		logger.Errorf("ChatStream error: %v", err)
		c.SSEvent("error", gin.H{"error": err.Error()})
		c.Writer.Flush()
	}
}

// ListProviders 处理 GET /api/v1/providers，返回所有已启用的 AI 提供者列表
func (h *ChatHandler) ListProviders(c *gin.Context) {
	providers := h.chatService.ListProviders()
	response.Success(c, providers)
}
