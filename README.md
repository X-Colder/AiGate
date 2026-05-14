# AiGate

AiGate 是一个多租户 AI 网关服务，统一代理 OpenAI、Anthropic、DeepSeek、豆包、通义千问、Kimi 等主流 AI 大模型 API。提供网关管理、策略控制（熔断/限流/降级）、RBAC 权限体系、监控统计等企业级功能，并附带 Vue3 管理前端。

## 架构概览

```
┌──────────────────────────────────────────────────────────┐
│                    Vue3 前端 (Element Plus)                │
│       登录 · 租户管理 · 用户/角色 · 网关管理 · 监控面板       │
└────────────────────────┬─────────────────────────────────┘
                         │ HTTP / REST
┌────────────────────────▼─────────────────────────────────┐
│                   Gin HTTP Server                        │
│  Middleware: Recovery → Logger → CORS → Auth → RBAC      │
├──────────────────────────────────────────────────────────┤
│  Handler    → Service    → Provider Registry             │
│  (路由处理)   (业务逻辑)    (AI 服务路由分发)                │
├──────────────────────────────────────────────────────────┤
│                Provider 层 (可插拔)                        │
│  OpenAI │ Anthropic │ DeepSeek │ 豆包 │ Qwen │ Kimi      │
├──────────────────────────────────────────────────────────┤
│  Store: MySQL + GORM 自动迁移                             │
└──────────────────────────────────────────────────────────┘
```

## 功能特性

- **多 AI 提供者** — 统一接口代理 6 种 AI 服务，支持通过配置动态启停
- **多租户隔离** — 数据按租户隔离，租户级网关与用量管理
- **RBAC 权限** — 角色定义模块级访问权限（租户管理/网关管理/监控面板/API开发），系统内置超级管理员、普通用户、开发者角色
- **网关策略** — 每个网关可独立配置限流（QPS/突发）、熔断（错误率阈值/恢复超时）、降级（备用 Provider + Model）
- **监控统计** — 按时间范围查询请求量、Token 消耗、平均延迟、错误率、独立用户数，支持趋势折线图
- **JWT 认证** — Bearer Token 认证，24 小时有效期
- **Vue3 前端** — Element Plus 蓝白主题管理界面，按权限动态渲染侧边栏菜单
- **C 端开发者能力** — API Key 管理、模型目录浏览、OpenAI 兼容推理接口、Token 计费、用量监控
- **模型管理** — 管理员配置可用模型、定价（按 Token/按次/月配额/免费额度）、接口文档
- **计费系统** — 预充值余额、实时扣费、交易流水、配额管理

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端框架 | Go + Gin |
| 数据库 | MySQL + GORM |
| 认证 | JWT (golang-jwt/v5) |
| 前端 | Vue 3 + Vite + Element Plus + ECharts |
| HTTP 客户端 | net/http (按提供者独立超时) |

## 快速开始

### 环境要求

- Go 1.21+
- MySQL 8.0+
- Node.js 18+（构建前端）

### 1. 克隆项目

```bash
git clone https://github.com/your-org/AiGate.git
cd AiGate
```

### 2. 配置

复制并编辑配置文件：

```bash
cp config.yaml.example config.yaml
```

在 `config.yaml` 中配置 MySQL 连接信息和 AI 提供者的 API Key：

```yaml
server:
  port: "8081"
  mode: "debug"

database:
  driver: "mysql"
  host: "127.0.0.1"
  port: 3306
  username: "aigate"
  password: "aigate_pass"
  database: "aigate"

jwt_secret: "your-secret-key"

providers:
  openai:
    enabled: true
    api_key: "sk-..."
    base_url: "https://api.openai.com/v1"
    model: "gpt-3.5-turbo"
    timeout: 30
```

也可通过环境变量覆盖数据库配置：`AIGATE_DB_HOST`、`AIGATE_DB_PORT`、`AIGATE_DB_USERNAME`、`AIGATE_DB_PASSWORD`、`AIGATE_DB_DATABASE`。

### 3. 构建前端

```bash
cd frontend
npm install
npm run build
cd ..
```

### 4. 准备数据库

确保 MySQL 服务已启动，并创建数据库：

```sql
CREATE DATABASE IF NOT EXISTS aigate DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS 'aigate'@'%' IDENTIFIED BY 'aigate_pass';
GRANT ALL PRIVILEGES ON aigate.* TO 'aigate'@'%';
FLUSH PRIVILEGES;
```

也可直接使用 Docker Compose 启动 MySQL（见下方 Docker 部署）。

### 5. 启动服务

```bash
go run main.go
```

服务默认监听 `http://localhost:8081`，首次启动自动创建数据表并初始化默认租户和管理员账户。

### 默认管理员

| 用户名 | 密码 |
|--------|------|
| admin | admin123 |

## 项目结构

```
AiGate/
├── main.go                 # 入口：加载配置 → 初始化数据库 → 启动 HTTP
├── config/                 # 配置加载（YAML）
├── router/                 # 路由注册 + 中间件挂载
├── handler/                # HTTP Handler（请求解析/响应格式化）
├── service/                # 业务逻辑层
├── provider/               # AI 提供者抽象接口 + 各实现
│   ├── provider.go         #   Provider 接口 + Registry
│   ├── openai.go           #   OpenAI 实现
│   ├── anthropic.go        #   Anthropic 实现
│   └── openai_compatible.go#   OpenAI 兼容协议基类（DeepSeek/豆包/Qwen/Kimi）
├── model/                  # 数据模型（Entity + DTO + Chat 协议）
├── middleware/             # 中间件（Auth/RBAC/APIKeyAuth/CORS/Logger/Recovery）
├── store/                  # 数据库初始化 + 默认数据 + Redis缓存
├── pkg/                    # 内部工具包
│   ├── auth/               #   JWT 生成/验证
│   ├── logger/             #   日志封装
│   ├── resilience/         #   熔断器
│   └── response/           #   统一 JSON 响应
├── frontend/               # Vue3 前端
│   └── src/
│       ├── views/          #   B端页面（Login/Gateways/Monitor/Tenants/Users/Roles/Models/Billing）
│       ├── views/developer/#   C端开发者控制台（Dashboard/APIKeys/ModelList/Usage/Balance）
│       ├── router/         #   前端路由 + RBAC 守卫
│       └── api/            #   Axios 请求封装
├── docs/                   # API 文档 + 测试报告
└── config.yaml             # 配置文件（已 gitignore）
```

## API 概览

所有接口前缀 `/api/v1`，认证接口除外均需 `Authorization: Bearer <token>` 头。

### 公开接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/auth/login` | 用户登录 |
| POST | `/api/v1/auth/register` | 用户注册 |
| GET | `/health` | 健康检查 |

### 需认证接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/chat` | 聊天（指定 provider 转发） |
| GET | `/api/v1/providers` | 已启用的提供者列表 |
| GET/POST/PUT/DELETE | `/api/v1/gateways[/:id]` | 网关 CRUD |
| PUT | `/api/v1/gateways/:id/policy` | 更新网关策略 |
| GET | `/api/v1/metrics/summary` | 监控汇总 |
| GET | `/api/v1/metrics/trend` | 监控趋势 |

### 管理员接口（需 tenant_access 权限）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET/POST/PUT/DELETE | `/api/v1/admin/tenants[/:id]` | 租户 CRUD |
| GET | `/api/v1/admin/tenants/:id/usage` | 租户使用详情 |
| GET/POST/PUT/DELETE | `/api/v1/admin/roles[/:id]` | 角色 CRUD |
| GET/POST/PUT/DELETE | `/api/v1/admin/users[/:id]` | 用户 CRUD |
| GET/POST/PUT/DELETE | `/api/v1/admin/models[/:id]` | 模型目录 CRUD |
| PUT | `/api/v1/admin/models/:id/doc` | 编辑模型接口文档 |
| GET | `/api/v1/admin/billing/users` | 用户余额列表 |
| POST | `/api/v1/admin/billing/recharge` | 用户充值 |

### 开发者接口（需认证）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST/GET/DELETE | `/api/v1/developer/apikeys[/:id]` | API Key 管理 |
| GET | `/api/v1/developer/models` | 可用模型列表 |
| GET | `/api/v1/developer/models/:id/doc` | 模型接口文档 |
| GET | `/api/v1/developer/balance` | 余额查询 |
| GET | `/api/v1/developer/transactions` | 交易流水 |
| GET | `/api/v1/developer/usage/summary` | 用量汇总 |
| GET | `/api/v1/developer/usage/trend` | 用量趋势 |
| GET | `/api/v1/developer/usage/records` | 调用记录 |

### OpenAI 兼容推理接口（API Key 认证）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/v1/chat/completions` | 聊天推理（OpenAI 兼容格式） |
| GET | `/v1/models` | 可用模型列表 |

> 推理接口使用 `Authorization: Bearer sk-xxx` 认证（API Key），与管理接口的 JWT Token 认证独立。

详细接口文档见 [docs/api_doc.md](docs/api_doc.md)。

## 支持的 AI 提供者

| 提供者 | 标识 | 协议 | 默认模型 |
|--------|------|------|----------|
| OpenAI | `openai` | OpenAI API | gpt-3.5-turbo |
| Anthropic | `anthropic` | Anthropic Messages API | claude-3-sonnet |
| DeepSeek | `deepseek` | OpenAI 兼容 | deepseek-chat |
| 豆包 (字节跳动) | `doubao` | OpenAI 兼容 | doubao-pro-32k |
| 通义千问 (阿里云) | `qwen` | OpenAI 兼容 | qwen-turbo |
| Kimi (月之暗面) | `kimi` | OpenAI 兼容 | moonshot-v1-8k |

新增提供者只需实现 `Provider` 接口并注册到 `Registry`。兼容 OpenAI 协议的服务可直接复用 `OpenAICompatibleProvider`。

## Docker 部署

项目提供 `docker-compose.yml`，一键启动 MySQL + Redis + AiGate：

```bash
docker-compose up -d
```

服务启动后访问 `http://localhost:8081`。

## C 端开发者接入

### 1. 管理员配置模型

登录管理后台 → 模型管理 → 新增模型，设置提供者、定价、计费模式和接口文档。

### 2. 开发者注册并创建 API Key

开发者使用 developer 角色登录后，进入开发者控制台 → API Keys → 创建 Key，获取 `sk-xxx` 格式的密钥。

### 3. 调用推理接口

使用 OpenAI 兼容格式调用，可直接用 OpenAI SDK：

```bash
curl http://localhost:8081/v1/chat/completions \
  -H "Authorization: Bearer sk-your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "deepseek-chat",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

```python
from openai import OpenAI

client = OpenAI(base_url="http://localhost:8081/v1", api_key="sk-your-api-key")
response = client.chat.completions.create(
    model="deepseek-chat",
    messages=[{"role": "user", "content": "Hello!"}]
)
print(response.choices[0].message.content)
```

## 运行测试

```bash
go test ./... -v
```

## License

MIT
