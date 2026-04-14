# AiGate API 文档

> Base URL: `http://localhost:8080`  
> API 版本: v1  
> Content-Type: `application/json`

---

## 统一响应格式

所有接口均采用统一的 JSON 响应结构：

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| code | int | 业务状态码，`0` 表示成功，`-1` 表示通用错误 |
| message | string | 状态描述 |
| data | object | 响应数据，错误时可能为空 |

---

## 接口列表

### 1. 健康检查

检查服务是否正常运行。

**请求**

```
GET /health
```

**响应示例**

```json
HTTP/1.1 200 OK

{
  "status": "ok",
  "service": "AiGate"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| status | string | 服务状态，`ok` 表示正常 |
| service | string | 服务名称 |

---

### 2. 获取提供者列表

获取所有已启用的 AI 服务提供者。

**请求**

```
GET /api/v1/providers
```

**响应示例**

```json
HTTP/1.1 200 OK

{
  "code": 0,
  "message": "success",
  "data": [
    {
      "name": "openai",
      "enabled": true
    },
    {
      "name": "anthropic",
      "enabled": true
    }
  ]
}
```

**响应字段 - data 数组元素**

| 字段 | 类型 | 说明 |
|------|------|------|
| name | string | 提供者名称 |
| enabled | bool | 是否启用 |
| models | string[] | 可用模型列表（可选） |

---

### 3. 聊天对话

向指定的 AI 提供者发送聊天请求。

**请求**

```
POST /api/v1/chat
Content-Type: application/json
```

**请求体**

```json
{
  "provider": "openai",
  "model": "gpt-4",
  "messages": [
    {
      "role": "system",
      "content": "You are a helpful assistant."
    },
    {
      "role": "user",
      "content": "Hello!"
    }
  ],
  "stream": false
}
```

**请求字段**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| provider | string | 是 | AI 提供者名称（`openai`、`anthropic`） |
| model | string | 否 | 模型名称，不填则使用配置中的默认模型 |
| messages | Message[] | 是 | 消息列表 |
| stream | bool | 否 | 是否使用流式响应，默认 `false` |

**Message 结构**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| role | string | 是 | 角色：`system`、`user`、`assistant` |
| content | string | 是 | 消息内容 |

#### 3.1 普通响应（stream=false）

**响应示例**

```json
HTTP/1.1 200 OK

{
  "code": 0,
  "message": "success",
  "data": {
    "id": "chatcmpl-abc123",
    "provider": "openai",
    "model": "gpt-4",
    "content": "Hello! How can I help you today?",
    "usage": {
      "prompt_tokens": 20,
      "completion_tokens": 10,
      "total_tokens": 30
    }
  }
}
```

**响应字段 - data**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 响应唯一标识 |
| provider | string | 提供者名称 |
| model | string | 实际使用的模型 |
| content | string | AI 回复内容 |
| usage | Usage | Token 用量统计 |

**Usage 结构**

| 字段 | 类型 | 说明 |
|------|------|------|
| prompt_tokens | int | 提示词消耗 Token 数 |
| completion_tokens | int | 生成内容消耗 Token 数 |
| total_tokens | int | 总消耗 Token 数 |

#### 3.2 流式响应（stream=true）

响应格式为 Server-Sent Events (SSE)。

**响应 Headers**

```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
```

**事件格式**

```
event: message
data: {"id":"chunk-1","provider":"openai","model":"gpt-4","delta":"Hello","done":false}

event: message
data: {"id":"chunk-2","provider":"openai","model":"gpt-4","delta":" World","done":true}
```

**StreamChunk 结构**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 块唯一标识 |
| provider | string | 提供者名称 |
| model | string | 模型名称 |
| delta | string | 本次增量内容 |
| done | bool | 是否为最后一个块 |

---

## 错误响应

### 400 Bad Request - 请求参数错误

```json
{
  "code": -1,
  "message": "Invalid request: Key: 'ChatRequest.Provider' Error:Field validation for 'Provider' failed on the 'required' tag"
}
```

### 500 Internal Server Error - 服务端错误

```json
{
  "code": -1,
  "message": "Chat error: get provider error: provider xxx not found"
}
```

---

## 支持的提供者

| 提供者 | 标识 | 默认模型 | 说明 |
|--------|------|----------|------|
| OpenAI | `openai` | gpt-3.5-turbo | 支持 GPT 系列模型 |
| Anthropic | `anthropic` | claude-3-sonnet-20240229 | 支持 Claude 系列模型 |