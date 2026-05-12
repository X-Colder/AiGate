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
│  Store: SQLite + GORM 自动迁移                            │
└──────────────────────────────────────────────────────────┘
```

## 功能特性

- **多 AI 提供者** — 统一接口代理 6 种 AI 服务，支持通过配置动态启停
- **多租户隔离** — 数据按租户隔离，租户级网关与用量管理
- **RBAC 权限** — 角色定义模块级访问权限（租户管理/网关管理/监控面板），系统内置超级管理员和普通用户角色
- **网关策略** — 每个网关可独立配置限流（QPS/突发）、熔断（错误率阈值/恢复超时）、降级（备用 Provider + Model）
- **监控统计** — 按时间范围查询请求量、Token 消耗、平均延迟、错误率、独立用户数，支持趋势折线图
- **JWT 认证** — Bearer Token 认证，24 小时有效期
- **Vue3 前端** — Element Plus 蓝白主题管理界面，按权限动态渲染侧边栏菜单

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端框架 | Go + Gin |
| 数据库 | SQLite + GORM |
| 认证 | JWT (golang-jwt/v5) |
| 前端 | Vue 3 + Vite + Element Plus + ECharts |
| HTTP 客户端 | net/http (按提供者独立超时) |

## 快速开始

### 环境要求

- Go 1.21+
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

在 `config.yaml` 中填入 AI 提供者的 API Key 并设置 `enabled: true`：

```yaml
server:
  port: "8081"
  mode: "debug"

jwt_secret: "your-secret-key"

providers:
  openai:
    enabled: true
    api_key: "sk-..."
    base_url: "https://api.openai.com/v1"
    model: "gpt-3.5-turbo"
    timeout: 30
```

也可通过环境变量 `AIGATE_CONFIG` 指定配置文件路径。

### 3. 构建前端

```bash
cd frontend
npm install
npm run build
cd ..
```

### 4. 启动服务

```bash
go run main.go
```

服务默认监听 `http://localhost:8081`，首次启动自动创建 SQLite 数据库并初始化默认租户和管理员账户。

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
├── middleware/             # 中间件（Auth/RBAC/CORS/Logger/Recovery）
├── store/                  # 数据库初始化 + 默认数据
├── pkg/                    # 内部工具包
│   ├── auth/               #   JWT 生成/验证
│   ├── logger/             #   日志封装
│   └── response/           #   统一 JSON 响应
├── frontend/               # Vue3 前端
│   └── src/
│       ├── views/          #   页面组件（Login/Gateways/Monitor/Tenants/Users/Roles）
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

## 运行测试

```bash
go test ./... -v
```

## License

MIT
