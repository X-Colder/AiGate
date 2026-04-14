package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/aigate/model"
	"github.com/aigate/pkg/logger"
	"github.com/aigate/pkg/response"
	"github.com/aigate/service"
)

// ChatHandler 聊天处理器
type ChatHandler struct {
	chatService *service.ChatService
}

// NewChatHandler 创建聊天处理器
func NewChatHandler(chatService *service.ChatService) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
	}
}

// Chat 处理聊天请求
func (h *ChatHandler) Chat(c *gin.Context) {
	var req model.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// 非流式请求
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

	// 流式请求 (SSE)
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

// ListProviders 获取可用的 AI 提供者列表
func (h *ChatHandler) ListProviders(c *gin.Context) {
	providers := h.chatService.ListProviders()
	response.Success(c, providers)
}
