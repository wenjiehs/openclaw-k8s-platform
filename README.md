# OpenClaw SaaS 平台

OpenClaw 企业级 SaaS 管理平台，基于 TKE（腾讯云容器服务）提供 OpenClaw AI 编程助手的多租户管理能力。

## 项目简介

该平台允许企业内员工自助申请 OpenClaw 实例，管理员在线审批，系统自动化部署和管理 K8s 资源，实现 OpenClaw 的企业级 SaaS 化运营。

### 核心功能

- **员工自助申请**：员工填写申请表单，管理员在线审批
- **自动化部署**：审批通过后，5分钟内自动创建 K8s 资源（Namespace、PVC、Deployment、Service、Ingress）
- **多租户隔离**：每个用户独立 Namespace + NetworkPolicy 网络隔离
- **监控大盘**：管理员可查看所有实例的运行状态、资源使用情况
- **成本管理**：按部门统计资源成本，支持配额管理
- **审计日志**：记录所有操作，满足合规要求

## 技术架构

```
┌─────────────────────────────────────────────────────┐
│                     前端层                            │
│         React 18 + TypeScript + Ant Design 5.x       │
└───────────────────────┬─────────────────────────────┘
                        │ HTTP/REST API
┌───────────────────────▼─────────────────────────────┐
│                     后端层                            │
│            Go 1.21 + Gin + GORM                      │
│      （认证/鉴权/审批/实例编排/监控聚合）               │
└───────────────────────┬─────────────────────────────┘
                        │
          ┌─────────────┴─────────────┐
          │                           │
┌─────────▼──────────┐  ┌────────────▼────────────┐
│    PostgreSQL       │  │         Redis           │
│  （业务数据存储）    │  │    （缓存/任务队列）      │
└────────────────────┘  └────────────────────────┘
                        │
┌───────────────────────▼─────────────────────────────┐
│                    TKE 集群                           │
│      每用户独立 Namespace + OpenClaw 富容器实例        │
└────────────────────────────────────────────────────┘
```

## 技术栈

| 组件 | 技术选型 | 版本 |
|-----|---------|------|
| 后端语言 | Go | 1.21 |
| 后端框架 | Gin | v1.9.1 |
| ORM | GORM | v1.25.5 |
| 数据库 | PostgreSQL | 12+ |
| 缓存 | Redis | 7.x |
| 前端框架 | React | 18.x |
| UI 组件库 | Ant Design | 5.x |
| 前端构建工具 | Vite | 4.x |
| 前端语言 | TypeScript | 5.x |
| 容器编排 | Kubernetes / TKE | 1.20+ |

## 目录结构

```
openclaw-k8s-platform/
├── backend/                    # Go 后端
│   ├── cmd/server/             # 启动入口
│   ├── internal/
│   │   ├── api/
│   │   │   ├── handler/        # HTTP 处理器
│   │   │   ├── middleware/     # 中间件（JWT/RBAC/审计）
│   │   │   └── router.go       # 路由配置
│   │   ├── model/              # 数据模型（GORM）
│   │   ├── service/            # 业务逻辑层
│   │   ├── repository/         # 数据访问层
│   │   └── k8s/                # K8s 客户端封装
│   ├── pkg/
│   │   ├── config/             # 配置加载
│   │   └── logger/             # 日志封装
│   ├── migrations/             # 数据库迁移 SQL
│   ├── go.mod
│   └── Dockerfile
├── frontend/                   # React 前端
│   ├── src/
│   │   ├── api/                # API 接口封装
│   │   ├── components/         # 公共组件
│   │   ├── pages/              # 页面组件
│   │   ├── store/              # 状态管理
│   │   ├── types/              # TypeScript 类型定义
│   │   └── utils/              # 工具函数
│   ├── package.json
│   ├── vite.config.ts
│   └── Dockerfile
├── k8s/                        # K8s 配置文件
│   ├── operator/               # OpenClaw Operator 安装
│   ├── base/                   # 基础资源（Namespace/RBAC）
│   └── saas-platform/          # SaaS 平台部署配置
├── docker-compose.yml          # 本地开发环境
├── Makefile                    # 常用命令
└── README.md
```

## 快速开始

### 前置条件

- Go 1.21+
- Node.js 18+
- Docker & Docker Compose
- kubectl（已配置 TKE 集群访问权限）

### 本地开发环境启动

1. **启动基础服务**（PostgreSQL + Redis）

```bash
docker-compose up -d postgres redis
```

2. **初始化数据库**

```bash
make db-migrate
```

3. **启动后端服务**

```bash
make backend-dev
```

4. **启动前端开发服务器**

```bash
make frontend-dev
```

5. **访问服务**

- 前端：http://localhost:3000
- 后端 API：http://localhost:8080
- API 文档：http://localhost:8080/swagger

### 完整环境（Docker Compose）

```bash
docker-compose up -d
```

访问 http://localhost:3000

## 环境变量配置

复制环境变量模板：

```bash
cp backend/.env.example backend/.env
```

主要配置项：

| 环境变量 | 说明 | 示例 |
|---------|------|------|
| `DATABASE_URL` | PostgreSQL 连接串 | `postgres://user:pass@localhost:5432/openclaw` |
| `REDIS_URL` | Redis 连接串 | `redis://localhost:6379` |
| `JWT_SECRET` | JWT 签名密钥 | 随机字符串 |
| `TKE_CLUSTER_ID` | TKE 集群 ID | `cls-xxxxxx` |
| `TKE_REGION` | TKE 集群所在地域 | `ap-guangzhou` |
| `QCLOUD_SECRET_ID` | 腾讯云 API 密钥 ID | `AKIDxxxx` |
| `QCLOUD_SECRET_KEY` | 腾讯云 API 密钥 Key | `xxxxxxxx` |

## K8s 集群接入注意事项

- 后端创建实例时会自动创建 Namespace/PVC/Deployment/Service/Ingress。
- PVC 默认使用 `cbs` 存储类；部署前请确认集群存在该 StorageClass：

```bash
kubectl get storageclass
```

- Namespace 标签会自动规范化（例如部门中文名会转换为合法 label），原始值会保存在 Namespace annotations。
- 如果实例长期处于 `creating`，建议优先排查以下资源事件：

```bash
kubectl -n <namespace> get pvc,pods,ingress
kubectl -n <namespace> describe pvc openclaw-data
kubectl -n <namespace> describe pod <pod-name>
```

## API 文档

### 认证接口

| 方法 | 路径 | 说明 |
|-----|------|------|
| POST | `/api/v1/auth/login` | 用户登录 |
| GET | `/api/v1/auth/me` | 获取当前用户信息 |

### 申请管理

| 方法 | 路径 | 说明 |
|-----|------|------|
| POST | `/api/v1/applications` | 提交申请 |
| GET | `/api/v1/applications` | 我的申请列表 |
| GET | `/api/v1/applications/:id` | 申请详情 |
| DELETE | `/api/v1/applications/:id` | 撤销申请 |

### 实例管理

| 方法 | 路径 | 说明 |
|-----|------|------|
| GET | `/api/v1/instances` | 我的实例列表 |
| GET | `/api/v1/instances/:id` | 实例详情 |
| DELETE | `/api/v1/instances/:id` | 删除实例 |

### 管理员接口

| 方法 | 路径 | 说明 |
|-----|------|------|
| GET | `/api/v1/admin/applications` | 待审批列表 |
| POST | `/api/v1/admin/applications/:id/approve` | 批准申请 |
| POST | `/api/v1/admin/applications/:id/reject` | 拒绝申请 |
| GET | `/api/v1/admin/instances` | 所有实例列表 |
| GET | `/api/v1/admin/metrics/summary` | 监控汇总数据 |
| GET | `/api/v1/admin/audit-logs` | 审计日志查询 |

## 部署

### 生产环境部署

```bash
# 构建镜像
make docker-build

# 部署到 TKE
make k8s-deploy
```

详细部署文档请参考 [k8s/README.md](k8s/README.md)。

## 开发规范

- 所有代码必须有中文注释
- 提交前运行 `make lint` 检查代码风格
- 提交前运行 `make test` 确保测试通过
- 数据库变更通过 migrations 文件管理

## 许可证

内部使用，暂不开源。
