# trade-hub

DEX 链下 API 后端服务，兼顾少量 CEX 概念（资产 / 订单 mock）。

**技术栈**：Go 1.22 · Gin · GORM · MySQL 8 · Redis 7 · JWT · Gorilla WebSocket

## 架构（DEX 侧重）

```text
React 前台 ──REST/WS──► Gin API
                          ├── Auth (JWT)
                          ├── /dex/swaps   ← Swap 入库（tx_hash 幂等）
                          ├── /dex/pools   ← 池子列表 / 行情缓存
                          ├── /ws/ticker   ← WebSocket 价格推送
                          └── /cex/*       ← 可选：资产 / 订单 mock
                          ▼
                    MySQL + Redis
```

链上签名在钱包；本服务只做 **链下数据、缓存、查询**（Indexer 可后续接 Solana RPC）。

## 快速启动

### 1. 依赖

- **Go 1.22+**（未安装：<https://go.dev/dl/> 下载 Windows 安装包，装完重开终端，`go version` 验证）
- **Docker Desktop**

### 2. 启动数据库

```bash
cd trade-hub
docker compose up -d
```

### 3. 配置环境变量

复制 `.env.example` 为 `.env`，本地 `go run` 时会自动读取（生产环境仍用平台注入的环境变量）：

```bash
cp .env.example .env
```

### 4. 运行 API

```bash
go mod tidy
go run ./cmd/api
```

访问：<http://localhost:8080/health>

### 5. 示例请求

```bash
# 注册
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"demo@example.com","password":"demo1234"}'

# 登录
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"demo@example.com","password":"demo1234"}'

# 记录 Swap（需 Bearer token）
curl -X POST http://localhost:8080/api/v1/dex/swaps \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "idempotency_key":"swap-001",
    "tx_hash":"5x...",
    "chain":"solana",
    "pool_id":"SOL-USDC",
    "amount_in":"1.5",
    "amount_out":"142.3",
    "status":"confirmed"
  }'
```

## 目录说明

```text
cmd/api/           入口 main
internal/config/   配置
internal/model/    GORM 模型
internal/repository/
internal/service/
internal/handler/
internal/middleware/
internal/ws/         WebSocket ticker
pkg/response/      统一 JSON 响应
```
