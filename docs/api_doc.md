# AiGate API 文档

> Base URL: `http://localhost:8081`  
> API 版本: v1  
> Content-Type: `application/json`  
> 认证方式: Bearer Token (JWT)

---

## 统一响应格式

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| code | int | `0` 成功，`-1` 错误 |
| message | string | 状态描述 |
| data | object | 响应数据 |

---

## 一、认证接口（公开）

### 1.1 用户登录

```
POST /api/v1/auth/login
```

**请求体**

```json
{
  "username": "admin",
  "password": "admin123"
}
```

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user_id": "uuid",
    "username": "admin",
    "tenant_id": "uuid",
    "role": "admin"
  }
}
```

### 1.2 用户注册

```
POST /api/v1/auth/register
```

**请求体**

```json
{
  "username": "newuser",
  "password": "abc123",
  "tenant_id": "00000000-0000-0000-0000-000000000001"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 3-50 字符 |
| password | string | 是 | 6-50 字符 |
| tenant_id | string | 是 | 归属租户 ID |

---

## 二、网关管理接口（需认证）

所有网关接口需要在请求头添加：`Authorization: Bearer <token>`

### 2.1 获取网关列表

```
GET /api/v1/gateways
```

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": "uuid",
      "tenant_id": "uuid",
      "name": "DeepSeek 网关",
      "provider": "deepseek",
      "base_url": "https://api.deepseek.com/v1",
      "model": "deepseek-chat",
      "timeout": 60,
      "status": 1,
      "policy": {
        "id": "uuid",
        "gateway_id": "uuid",
        "rate_limit_enabled": true,
        "rate_limit_qps": 100,
        "rate_limit_burst": 200,
        "circuit_breaker_enabled": false,
        "circuit_breaker_threshold": 0.5,
        "circuit_breaker_timeout": 30,
        "circuit_breaker_min_reqs": 10,
        "fallback_enabled": false,
        "fallback_provider": "",
        "fallback_model": ""
      },
      "created_at": "2026-04-15T10:00:00Z",
      "updated_at": "2026-04-15T10:00:00Z"
    }
  ]
}
```

### 2.2 创建网关

```
POST /api/v1/gateways
```

**请求体**

```json
{
  "name": "DeepSeek 网关",
  "provider": "deepseek",
  "base_url": "https://api.deepseek.com/v1",
  "api_key": "sk-xxx",
  "model": "deepseek-chat",
  "timeout": 60
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 网关名称 |
| provider | string | 是 | AI 服务标识 |
| base_url | string | 否 | API 基础地址 |
| api_key | string | 否 | API 密钥 |
| model | string | 否 | 默认模型 |
| timeout | int | 否 | 超时（秒），默认 60 |

### 2.3 获取单个网关

```
GET /api/v1/gateways/:id
```

### 2.4 更新网关

```
PUT /api/v1/gateways/:id
```

**请求体**（所有字段可选）

```json
{
  "name": "新名称",
  "provider": "qwen",
  "status": 0
}
```

### 2.5 删除网关

```
DELETE /api/v1/gateways/:id
```

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": { "message": "gateway deleted" }
}
```

### 2.6 更新网关策略

```
PUT /api/v1/gateways/:id/policy
```

**请求体**（所有字段可选）

```json
{
  "rate_limit_enabled": true,
  "rate_limit_qps": 50,
  "rate_limit_burst": 100,
  "circuit_breaker_enabled": true,
  "circuit_breaker_threshold": 0.3,
  "circuit_breaker_timeout": 30,
  "circuit_breaker_min_reqs": 10,
  "fallback_enabled": true,
  "fallback_provider": "qwen",
  "fallback_model": "qwen-turbo"
}
```

**策略字段说明**

| 字段 | 类型 | 说明 |
|------|------|------|
| rate_limit_enabled | bool | 启用限流 |
| rate_limit_qps | int | 每秒最大请求数 |
| rate_limit_burst | int | 突发最大请求数 |
| circuit_breaker_enabled | bool | 启用熔断 |
| circuit_breaker_threshold | float | 错误率阈值 (0-1) |
| circuit_breaker_timeout | int | 熔断恢复时间（秒） |
| circuit_breaker_min_reqs | int | 触发熔断最小请求数 |
| fallback_enabled | bool | 启用降级 |
| fallback_provider | string | 降级备用服务 |
| fallback_model | string | 降级使用模型 |

---

## 三、监控接口（需认证）

### 3.1 获取监控汇总

```
GET /api/v1/metrics/summary?start_date=2026-04-01&end_date=2026-04-15&gateway_id=xxx
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| start_date | string | 是 | 开始日期 YYYY-MM-DD |
| end_date | string | 是 | 结束日期 YYYY-MM-DD |
| gateway_id | string | 否 | 指定网关，不填则查全部 |

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total_requests": 10000,
    "total_tokens": 500000,
    "avg_latency_ms": 150.5,
    "total_errors": 50,
    "unique_users": 120,
    "error_rate": 0.005
  }
}
```

### 3.2 获取监控趋势（按日聚合）

```
GET /api/v1/metrics/trend?start_date=2026-04-01&end_date=2026-04-15&gateway_id=xxx
```

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "date": "2026-04-01",
      "requests": 1500,
      "tokens": 75000,
      "avg_latency": 140.2,
      "errors": 5,
      "users": 30
    }
  ]
}
```

---

## 四、聊天接口（需认证）

### 4.1 聊天对话

```
POST /api/v1/chat
```

**请求体**

```json
{
  "provider": "deepseek",
  "model": "deepseek-chat",
  "messages": [
    { "role": "user", "content": "Hello!" }
  ],
  "stream": false
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| provider | string | 是 | `openai`/`anthropic`/`deepseek`/`doubao`/`qwen`/`kimi` |
| model | string | 否 | 模型名称 |
| messages | Message[] | 是 | 消息列表 |
| stream | bool | 否 | 流式响应 |

### 4.2 获取提供者列表

```
GET /api/v1/providers
```

---

## 五、租户管理接口（管理员专用）

所有租户管理接口需 admin 角色 Token：`Authorization: Bearer <admin_token>`

### 5.1 获取租户列表

```
GET /api/v1/admin/tenants
```

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": "uuid",
      "name": "Default",
      "status": 1,
      "created_at": "2026-04-15T10:00:00Z",
      "updated_at": "2026-04-15T10:00:00Z",
      "user_count": 5,
      "gateway_count": 3
    }
  ]
}
```

### 5.2 创建租户

```
POST /api/v1/admin/tenants
```

**请求体**

```json
{
  "name": "新租户"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 2-100 字符，不可重名 |

### 5.3 获取单个租户

```
GET /api/v1/admin/tenants/:id
```

### 5.4 更新租户

```
PUT /api/v1/admin/tenants/:id
```

**请求体**（字段可选）

```json
{
  "name": "新名称",
  "status": 0
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 否 | 租户名称 |
| status | int | 否 | 1=启用 0=禁用 |

### 5.5 删除租户

```
DELETE /api/v1/admin/tenants/:id
```

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": { "message": "tenant deleted" }
}
```

> 删除租户会级联删除其下所有用户、网关、策略和监控数据。

### 5.6 获取租户使用详情

```
GET /api/v1/admin/tenants/:id/usage
```

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "tenant_id": "uuid",
    "tenant_name": "Default",
    "gateways": [
      {
        "gateway_id": "uuid",
        "gateway_name": "DeepSeek 网关",
        "provider": "deepseek",
        "status": 1,
        "tokens_used": 50000,
        "request_count": 1200,
        "avg_latency_ms": 150.5,
        "error_rate": 0.005
      }
    ],
    "total_tokens": 50000,
    "total_requests": 1200
  }
}
```

---

## 六、公开接口

### 6.1 健康检查

```
GET /health
```

**响应**

```json
{
  "status": "ok",
  "service": "AiGate"
}
```

---

## 七、支持的 AI 服务

| 提供者 | 标识 | API 地址 | 默认模型 |
|--------|------|----------|----------|
| OpenAI | `openai` | api.openai.com/v1 | gpt-3.5-turbo |
| Anthropic | `anthropic` | api.anthropic.com/v1 | claude-3-sonnet-20240229 |
| DeepSeek | `deepseek` | api.deepseek.com/v1 | deepseek-chat |
| 豆包 | `doubao` | ark.cn-beijing.volces.com/api/v3 | doubao-pro-32k |
| 通义千问 | `qwen` | dashscope.aliyuncs.com/compatible-mode/v1 | qwen-turbo |
| Kimi | `kimi` | api.moonshot.cn/v1 | moonshot-v1-8k |

---

## 八、错误响应

### 400 Bad Request

```json
{ "code": -1, "message": "Invalid request: ..." }
```

### 401 Unauthorized

```json
{ "code": -1, "message": "missing authorization header" }
```

### 403 Forbidden

```json
{ "code": -1, "message": "admin access required" }
```

### 404 Not Found

```json
{ "code": -1, "message": "not found" }
```

### 500 Internal Server Error

```json
{ "code": -1, "message": "..." }
```

---

## 九、默认账户

| 用户名 | 密码 | 角色 | 租户 |
|--------|------|------|------|
| admin | admin123 | admin | Default |