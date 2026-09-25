# CostGuard · 建筑工程项目成本管控 API 服务

面向建筑公司财务与项目管理的纯后端 RESTful API：项目预算编制与审批、成本归集与差异核对、变更单管理、成本分析报表生成（Redis 缓存）。

## 快速启动（Docker Compose 一键部署）

```bash
cp .env.example .env
docker compose up -d --build
```

- 后端 API：http://localhost:19203
- Swagger / API 文档：http://localhost:19203/docs
- 健康检查：http://localhost:19203/healthz
- 就绪检查：http://localhost:19203/readyz

测试账号：
- `admin / admin123`（Admin）
- `finance / finance123`（FinanceManager，可审批预算）
- `accountant / accountant123`（Accountant，可录入成本）
- `pm / pm123456`（ProjectManager）
- `viewer / viewer123`（Viewer，只读）

## API 功能列表

- 项目预算：编制、编辑、提交审批、审批通过/驳回（仅 FinanceManager/Admin）
- 成本项：录入（自动计算差异金额）、核对、标记异常（仅 Accountant/FinanceManager/Admin）
- 变更单：发起（自动计算变更后金额）、提交审批、审批/驳回/作废
- 成本分析报表：按月/季/年度生成，Redis 缓存 10 分钟
- 横切：JWT + RBAC、审计日志、Redis 限流（每 IP 60 次/分钟）、全局异常处理、请求日志

## 本地开发

```bash
cd backend
go mod tidy
go run ./cmd/server
```

构建检查：`go build ./...`；测试：`go test ./...`

## 技术栈

| 层级 | 技术 |
| --- | --- |
| 后端 | Go 1.22 + Gin + GORM + PostgreSQL + Redis |
| 数据库 | PostgreSQL 15 |
| 缓存/限流 | Redis 7（go-redis/v9） |
| 认证 | JWT（golang-jwt/v5）+ RBAC + bcrypt |
| API 文档 | /docs（静态 OpenAPI + 页面） |
| 部署 | Docker Compose |

## 目录结构

```
├── backend/
│   ├── cmd/server/main.go
│   ├── database/migrations/    # 参考 DDL（建表由 GORM AutoMigrate）
│   ├── database/seeds/         # 种子数据说明
│   └── internal/
│       ├── config/  ├── model/  ├── repository/  ├── service/
│       ├── handler/ ├── router/ ├── middleware/  ├── dto/
│       ├── constants/ ├── util/ ├── logger/      └── docs/
├── docker-compose.yml
└── .env.example
```

## 环境变量

| 变量 | 说明 | 默认 |
| --- | --- | --- |
| COMPOSE_PROJECT_NAME | Compose 项目名 | costguard |
| BACKEND_PORT / DB_PORT / REDIS_PORT | 端口 | 19203 / 33334 / 36334 |
| DB_NAME / DB_USER / DB_PASSWORD | PostgreSQL 配置 | costguard / costguard / costguard_pwd |
| REDIS_HOST / REDIS_PORT / REDIS_PASSWORD / REDIS_DB | Redis 配置 | redis / 6379 / 空 / 0 |
| JWT_SECRET | JWT 签名密钥，生产必须替换为 32 位以上随机值 | 见 .env.example |
| CORS_ALLOWED_ORIGINS | 允许跨域来源，逗号分隔；生产禁止 `*` | http://localhost:19203 |
| AUTH_RATE_LIMIT / API_RATE_LIMIT | 登录/业务接口限流（次/分钟/IP） | 10 / 60 |
| DB_MAX_OPEN_CONNS / DB_MAX_IDLE_CONNS | PostgreSQL 连接池 | 25 / 5 |
| DB_CONN_MAX_LIFETIME_MIN | 连接最大存活时间（分钟） | 5 |
| DB_CONNECT_RETRIES / DB_CONNECT_RETRY_INTERVAL_SEC | PostgreSQL 启动重试次数/间隔 | 10 / 3 |
| REDIS_CONNECT_RETRIES / REDIS_CONNECT_RETRY_INTERVAL_SEC | Redis 启动重试次数/间隔 | 10 / 3 |

## 主要 API 列表

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | /api/v1/auth/login | JWT 登录 |
| GET/POST | /api/v1/budgets | 预算列表/编制 |
| GET/PUT | /api/v1/budgets/:id | 详情/编辑草稿 |
| POST | /api/v1/budgets/:id/submit · /approve · /reject | 预算审批流 |
| GET/POST | /api/v1/cost-items | 成本列表/录入 |
| GET/PUT | /api/v1/cost-items/:id | 详情/核对 |
| POST | /api/v1/cost-items/:id/mark-abnormal | 标记异常 |
| GET/POST | /api/v1/change-orders | 变更列表/发起 |
| GET/PUT | /api/v1/change-orders/:id | 详情/编辑草稿 |
| POST | /api/v1/change-orders/:id/submit · /approve · /reject · /cancel | 变更审批流 |
| GET | /api/v1/reports | 报表列表 |
| POST | /api/v1/reports/generate | 生成报表（Redis 缓存） |
| GET | /api/v1/reports/:id | 读取报表（优先缓存） |
| GET | /api/v1/audit-logs | 审计日志 |
| GET | /healthz、/readyz | 健康检查 |

统一响应格式：`{ "code": 0, "message": "ok", "data": ... }`

## 枚举定义位置

| 枚举 | 文件 |
| --- | --- |
| BudgetStatus（Draft/Submitted/Approved/Rejected） | `backend/internal/constants/budget_status.go` |
| CostCategory（Material/Labor/Equipment/Subcontract/Overhead/Other） | `backend/internal/constants/cost_category.go` |
| ChangeType（ScopeChange/DesignChange/PriceAdjustment/UnforeseenCondition） | `backend/internal/constants/change_type.go` |
| ChangeOrderStatus（Draft/Submitted/Approved/Rejected/Cancelled） | `backend/internal/constants/change_order_status.go` |
| UserRole（Admin/FinanceManager/ProjectManager/Accountant/Viewer） | `backend/internal/constants/roles.go` |

## License

MIT
