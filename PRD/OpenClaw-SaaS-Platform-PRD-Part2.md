# OpenClaw SaaS 平台产品需求文档（Part 2）

**接续 Part 1**

---

## 五、技术方案与部署架构

### 5.1 整体架构设计

#### 5.1.1 系统架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                         用户层                                    │
├─────────────────────────────────────────────────────────────────┤
│  员工端                    │  管理员端                │  OpenClaw 实例  │
│  • Web 控制台              │  • 管理控制台             │  • Web 界面     │
│  • 企业微信 Bot            │  • 审批工作台             │  • API 接口     │
│  • Slack Bot              │  • 监控大盘               │                │
└─────────────────────────────────────────────────────────────────┘
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      OpenClaw SaaS 平台                          │
├─────────────────────────────────────────────────────────────────┤
│  前端服务                   │  后端服务                            │
│  ┌──────────────┐          │  ┌──────────────────────────────┐  │
│  │ React + Ant  │          │  │  API Gateway（认证/鉴权/限流）  │  │
│  │ Design       │◄─────────┤  └──────────────────────────────┘  │
│  └──────────────┘          │              ▼                      │
│                            │  ┌──────────────────────────────┐  │
│                            │  │  核心业务服务（Go + Gin）       │  │
│                            │  ├──────────────────────────────┤  │
│                            │  │ • 申请管理服务                  │  │
│                            │  │ • 审批管理服务                  │  │
│                            │  │ • 实例编排服务（部署/销毁）      │  │
│                            │  │ • 监控数据聚合服务               │  │
│                            │  │ • 成本计算服务                  │  │
│                            │  │ • 通知服务（企业微信/邮件）      │  │
│                            │  └──────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                       数据存储层                                  │
├─────────────────────────────────────────────────────────────────┤
│  PostgreSQL            │  Redis                │  Prometheus     │
│  • 用户表               │  • 任务队列            │  • 监控指标      │
│  • 实例表               │  • 缓存（审批状态）     │  • 告警规则      │
│  • 申请记录             │  • 分布式锁            │                │
│  • 成本记录             │                       │                │
│  • 审计日志             │                       │                │
└─────────────────────────────────────────────────────────────────┘
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      TKE 集群（容器编排层）                        │
├─────────────────────────────────────────────────────────────────┤
│  OpenClaw 实例（富容器模式，每用户独立）                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │ Namespace:   │  │ Namespace:   │  │ Namespace:   │         │
│  │ openclaw-    │  │ openclaw-    │  │ openclaw-    │         │
│  │ zhangsan     │  │ lisi         │  │ wangwu       │         │
│  ├──────────────┤  ├──────────────┤  ├──────────────┤         │
│  │ Pod:         │  │ Pod:         │  │ Pod:         │         │
│  │ • Gateway    │  │ • Gateway    │  │ • Gateway    │         │
│  │ • Agent      │  │ • Agent      │  │ • Agent      │         │
│  │ • Skills     │  │ • Skills     │  │ • Skills     │         │
│  │ • Memory     │  │ • Memory     │  │ • Memory     │         │
│  ├──────────────┤  ├──────────────┤  ├──────────────┤         │
│  │ PVC: 10GB    │  │ PVC: 10GB    │  │ PVC: 10GB    │         │
│  ├──────────────┤  ├──────────────┤  ├──────────────┤         │
│  │ Service +    │  │ Service +    │  │ Service +    │         │
│  │ Ingress      │  │ Ingress      │  │ Ingress      │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
│                                                                 │
│  共享基础设施：                                                   │
│  • TKE 监控服务（Prometheus）                                     │
│  • TKE 日志服务（CLS）                                            │
│  • TKE 镜像仓库（TCR）                                            │
│  • NetworkPolicy（网络隔离）                                      │
└─────────────────────────────────────────────────────────────────┘
```

---

### 5.2 部署架构详解

#### 5.2.1 OpenClaw 实例部署模式（富容器模式）

**选型理由**：

| 方案 | 优势 | 劣势 | 是否采用 |
|-----|------|------|---------|
| **方案 A：共享 Gateway + 按需 Agent** | 节省资源，共享 Gateway 降低成本 | • 多租户共享 Gateway，隔离复杂<br>• 按需启动 Agent 延迟高（冷启动 30s+） | ❌ 不采用 |
| **方案 B：每用户独立富容器** | • 完全隔离，安全性高<br>• 7×24 在线，响应快<br>• 配置独立，易于管理 | 资源占用较高（每实例 2-4GB 内存） | ✅ 采用 |

**部署清单（每个用户实例）**：

```yaml
# 1. Namespace（隔离边界）
apiVersion: v1
kind: Namespace
metadata:
  name: openclaw-zhangsan
  labels:
    app: openclaw
    user: zhangsan
    department: research-dev

---
# 2. PVC（持久化存储）
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: openclaw-data
  namespace: openclaw-zhangsan
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
  storageClassName: tke-cbs  # TKE 云硬盘

---
# 3. Deployment（OpenClaw 富容器）
apiVersion: apps/v1
kind: Deployment
metadata:
  name: openclaw
  namespace: openclaw-zhangsan
spec:
  replicas: 1
  selector:
    matchLabels:
      app: openclaw
  template:
    metadata:
      labels:
        app: openclaw
        user: zhangsan
    spec:
      containers:
      - name: openclaw-all-in-one
        image: tcr.tke-cloud.com/openclaw:v2.3
        ports:
        - containerPort: 8080  # Gateway 端口
        resources:
          requests:
            cpu: "2"          # 标准版：2 核
            memory: "4Gi"     # 标准版：4GB
          limits:
            cpu: "2"
            memory: "4Gi"
        env:
        - name: WECHAT_BOT_TOKEN
          valueFrom:
            secretKeyRef:
              name: wechat-secret
              key: token
        - name: TKE_CLUSTER_ENDPOINT
          value: "https://tke-api.tke-cloud.com"
        volumeMounts:
        - name: data
          mountPath: /app/data  # Skills、配置、对话历史
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: openclaw-data

---
# 4. Service（内部访问）
apiVersion: v1
kind: Service
metadata:
  name: openclaw-service
  namespace: openclaw-zhangsan
spec:
  selector:
    app: openclaw
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080

---
# 5. Ingress（外部访问）
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: openclaw-ingress
  namespace: openclaw-zhangsan
  annotations:
    kubernetes.io/ingress.class: "nginx"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"  # 自动 HTTPS 证书
spec:
  tls:
  - hosts:
    - openclaw-zhangsan.tke-cloud.com
    secretName: openclaw-zhangsan-tls
  rules:
  - host: openclaw-zhangsan.tke-cloud.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: openclaw-service
            port:
              number: 80

---
# 6. NetworkPolicy（网络隔离）
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: openclaw-netpol
  namespace: openclaw-zhangsan
spec:
  podSelector:
    matchLabels:
      app: openclaw
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx  # 仅允许 Ingress 访问
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - namespaceSelector: {}  # 允许访问 TKE API、企业微信 API
    ports:
    - protocol: TCP
      port: 443
```

---

#### 5.2.2 自动化部署流程（5 分钟内完成）

```
管理员批准申请
       ▼
┌──────────────────────────────────────────────────────────────┐
│  步骤 1：创建 Namespace（10 秒）                               │
│  ├─ 调用 TKE API：POST /api/v1/namespaces                    │
│  └─ 添加 Label：user=zhangsan, department=research-dev       │
└──────────────────────────────────────────────────────────────┘
       ▼
┌──────────────────────────────────────────────────────────────┐
│  步骤 2：创建 PVC（30 秒）                                     │
│  ├─ 调用 TKE API：POST /api/v1/namespaces/{ns}/pvcs          │
│  ├─ 存储类型：TKE CBS（云硬盘）                                │
│  └─ 等待 PVC Bound                                            │
└──────────────────────────────────────────────────────────────┘
       ▼
┌──────────────────────────────────────────────────────────────┐
│  步骤 3：创建 Secret（企业微信 Token）（5 秒）                  │
│  ├─ 调用企业微信 API：获取 Bot Token                           │
│  └─ 存储到 K8s Secret                                         │
└──────────────────────────────────────────────────────────────┘
       ▼
┌──────────────────────────────────────────────────────────────┐
│  步骤 4：创建 Deployment（2 分钟）                             │
│  ├─ 调用 TKE API：POST /apis/apps/v1/namespaces/{ns}/deployments │
│  ├─ 镜像：tcr.tke-cloud.com/openclaw:v2.3                     │
│  ├─ 资源配额：CPU 2核，内存 4GB（标准版）                       │
│  └─ 等待 Pod Running                                          │
└──────────────────────────────────────────────────────────────┘
       ▼
┌──────────────────────────────────────────────────────────────┐
│  步骤 5：创建 Service + Ingress（30 秒）                       │
│  ├─ 创建 Service：ClusterIP                                   │
│  ├─ 创建 Ingress：域名 openclaw-zhangsan.tke-cloud.com        │
│  └─ 自动申请 HTTPS 证书（Let's Encrypt）                       │
└──────────────────────────────────────────────────────────────┘
       ▼
┌──────────────────────────────────────────────────────────────┐
│  步骤 6：健康检查（1 分钟）                                    │
│  ├─ HTTP GET https://openclaw-zhangsan.tke-cloud.com/health  │
│  ├─ 检查返回：{"status": "healthy"}                           │
│  └─ 重试 3 次，间隔 20 秒                                      │
└──────────────────────────────────────────────────────────────┘
       ▼
┌──────────────────────────────────────────────────────────────┐
│  步骤 7：创建 NetworkPolicy（10 秒）                           │
│  ├─ 配置网络隔离规则                                           │
│  └─ 仅允许 Ingress 和必要的外部 API 访问                        │
└──────────────────────────────────────────────────────────────┘
       ▼
┌──────────────────────────────────────────────────────────────┐
│  步骤 8：初始化配置（30 秒）                                   │
│  ├─ 调用 OpenClaw API：配置 TKE 集群信息                       │
│  ├─ 安装默认 Skills（TKE 运维常用 Skills）                     │
│  └─ 创建欢迎消息                                               │
└──────────────────────────────────────────────────────────────┘
       ▼
┌──────────────────────────────────────────────────────────────┐
│  步骤 9：发送通知（5 秒）                                      │
│  ├─ 企业微信通知用户："你的 OpenClaw 已就绪！"                  │
│  ├─ 邮件通知：包含访问地址和快速入门指南                         │
│  └─ 更新平台状态：实例状态 = Running                            │
└──────────────────────────────────────────────────────────────┘

总耗时：约 5 分钟（P50），< 10 分钟（P95）
```

---

### 5.3 多租户安全与隔离

#### 5.3.1 四层隔离机制

| 隔离层级 | 隔离方式 | 作用 | 实现细节 |
|---------|---------|------|---------|
| **1. Namespace 隔离** | K8s Namespace | 逻辑隔离，不同用户的资源互不可见 | 每个实例独立 Namespace（openclaw-{username}） |
| **2. RBAC 权限隔离** | K8s RBAC | 员工只能操作自己的实例 | • 员工角色：仅能访问自己的 Namespace<br>• 管理员角色：可访问所有 Namespace |
| **3. 网络隔离** | NetworkPolicy | 不同实例间网络不互通 | • Ingress：仅允许 Nginx Ingress 访问<br>• Egress：仅允许访问 TKE API、企业微信 API |
| **4. 数据隔离** | 独立 PVC | 每个实例的数据独立存储 | 每个实例挂载独立的 TKE CBS 云硬盘 |

#### 5.3.2 RBAC 权限矩阵

| 角色 | 可操作资源 | 具体权限 |
|-----|-----------|---------|
| **员工（Employee）** | 自己的实例（Namespace: openclaw-{username}） | • 查看实例状态<br>• 查看日志<br>• 删除实例<br>❌ 无法修改配额、访问其他用户实例 |
| **管理员（Admin）** | 所有实例 + 平台管理 | • 审批申请<br>• 查看/操作所有实例<br>• 配置配额<br>• 强制删除实例<br>• 查看审计日志 |
| **超级管理员（Super Admin）** | 平台所有功能 | • 管理用户角色<br>• 修改全局配置<br>• 升级平台版本 |

#### 5.3.3 安全加固措施

| 安全措施 | 实现方式 | 目的 |
|---------|---------|------|
| **API 认证** | JWT Token（有效期 1 小时，自动刷新） | 防止未授权访问 |
| **API 鉴权** | 基于 RBAC 的细粒度权限控制 | 确保用户只能操作自己的资源 |
| **审计日志** | 记录所有 API 调用（谁、何时、做了什么、结果） | 安全审计、合规要求 |
| **敏感数据加密** | • PVC 数据加密存储<br>• Secret 加密（企业微信 Token） | 防止数据泄露 |
| **镜像安全扫描** | 定期扫描 OpenClaw 镜像的 CVE 漏洞 | 及时发现和修复安全漏洞 |
| **网络隔离** | NetworkPolicy + TKE 安全组 | 防止横向攻击 |
| **配额限制** | 部门配额 + 个人配额 | 防止资源滥用 |

---

### 5.4 监控与告警

#### 5.4.1 监控指标体系

| 指标类型 | 指标名称 | 采集方式 | 告警阈值 |
|---------|---------|---------|---------|
| **实例健康** | instance_status | Prometheus（每 1 分钟） | Pod 状态 != Running 持续 5 分钟 |
| **资源使用** | cpu_usage_percent | Prometheus | CPU > 90% 持续 10 分钟 |
| **资源使用** | memory_usage_percent | Prometheus | 内存 > 90% 持续 10 分钟 |
| **API 性能** | api_latency_seconds | Prometheus | P99 延迟 > 2 秒 |
| **部署成功率** | deployment_success_rate | 后端日志 | 成功率 < 95% |
| **成本** | monthly_cost_by_department | PostgreSQL | 超出预算 10% |

#### 5.4.2 告警规则示例

```yaml
# Prometheus AlertManager 规则
groups:
- name: openclaw_alerts
  rules:
  # 实例异常告警
  - alert: InstanceUnhealthy
    expr: kube_pod_status_phase{namespace=~"openclaw-.*", phase!="Running"} == 1
    for: 5m
    labels:
      severity: critical
    annotations:
      summary: "OpenClaw 实例 {{ $labels.namespace }} 异常"
      description: "Pod 状态：{{ $labels.phase }}，持续 5 分钟"

  # CPU 过高告警
  - alert: HighCPUUsage
    expr: |
      rate(container_cpu_usage_seconds_total{namespace=~"openclaw-.*"}[5m]) > 0.9
    for: 10m
    labels:
      severity: warning
    annotations:
      summary: "实例 {{ $labels.namespace }} CPU 使用率过高"
      description: "当前 CPU 使用率：{{ $value }}%"

  # 部署失败率告警
  - alert: HighDeploymentFailureRate
    expr: |
      (sum(rate(deployment_failures_total[1h])) / sum(rate(deployment_total[1h]))) > 0.05
    for: 30m
    labels:
      severity: critical
    annotations:
      summary: "部署失败率过高"
      description: "最近 1 小时部署失败率：{{ $value }}%"
```

---

## 六、非功能需求

### 6.1 性能要求

| 性能指标 | 目标值 | 验证方式 |
|---------|--------|---------|
| **实例创建时长** | P50 < 5 分钟，P95 < 10 分钟 | 压测：并发创建 20 个实例 |
| **审批响应时间** | 批准操作 < 2 秒（从点击到状态更新） | 前端性能监控 |
| **大盘加载时间** | 管理员控制台首屏加载 < 3 秒 | 前端性能监控 |
| **并发审批** | 支持 10 个管理员同时审批，无冲突 | 压测：模拟并发审批 |
| **实例并发数** | 单集群支持 100 个实例同时运行 | 压测：TKE 集群资源充足情况下 |
| **API 响应时间** | 99% 请求响应时间 < 500ms | 后端性能监控（Prometheus） |

### 6.2 安全要求

| 安全要求 | 实现方案 | 验证方式 |
|---------|---------|---------|
| **身份认证** | SSO 单点登录（支持 LDAP/企业微信/飞书） | 安全测试：暴力破解防护 |
| **权限控制** | RBAC：员工只能操作自己的实例，管理员可操作所有实例 | 单元测试：权限边界测试 |
| **数据隔离** | 每个实例独立 Namespace + NetworkPolicy 网络隔离 | 渗透测试：跨实例访问 |
| **审计日志** | 记录所有 API 调用（谁、何时、做了什么），保留 180 天 | 合规审计：等保三级要求 |
| **敏感数据加密** | PVC 数据加密存储，API Token 加密传输（HTTPS） | 安全扫描：TLS 配置检查 |
| **防止配额滥用** | 部门配额 + 个人配额双重限制，异常申请告警 | 功能测试：尝试超额申请 |
| **安全漏洞修复** | 定期扫描镜像漏洞，自动推送升级通知 | 安全扫描：镜像 CVE 检测 |

### 6.3 兼容性要求

| 兼容性维度 | 要求 | 备注 |
|-----------|------|------|
| **浏览器** | Chrome 90+、Edge 90+、Firefox 88+、Safari 14+ | 前端使用 React + Ant Design |
| **TKE 版本** | TKE 1.20+ | 依赖 K8s Ingress、NetworkPolicy 特性 |
| **企业通讯工具** | 企业微信、飞书、钉钉、Slack | MVP 阶段先支持企业微信 |
| **移动端** | 响应式设计，支持手机/平板访问管理控制台 | 员工主要通过企业微信使用 |
| **数据库** | PostgreSQL 12+ | 后续可支持 MySQL |

---

## 七、数据需求

### 7.1 数据库表设计（核心表）

#### 表 1：users（用户表）

| 字段名 | 类型 | 说明 | 示例 |
|-------|------|------|------|
| id | SERIAL PRIMARY KEY | 用户 ID | 1 |
| username | VARCHAR(50) UNIQUE | 用户名（唯一） | zhangsan |
| email | VARCHAR(100) | 邮箱 | zhangsan@company.com |
| department | VARCHAR(100) | 部门 | 研发部-运维组 |
| role | VARCHAR(20) | 角色（employee/admin/super_admin） | employee |
| created_at | TIMESTAMP | 创建时间 | 2026-03-04 10:00:00 |

#### 表 2：instances（实例表）

| 字段名 | 类型 | 说明 | 示例 |
|-------|------|------|------|
| id | SERIAL PRIMARY KEY | 实例 ID | 1001 |
| name | VARCHAR(50) UNIQUE | 实例名称（唯一） | openclaw-zhangsan |
| user_id | INT REFERENCES users(id) | 所属用户 | 1 |
| spec | VARCHAR(20) | 规格（basic/standard/enterprise） | standard |
| status | VARCHAR(20) | 状态（pending/running/stopped/failed） | running |
| namespace | VARCHAR(100) | K8s Namespace | openclaw-zhangsan |
| ingress_url | VARCHAR(200) | 访问地址 | https://openclaw-zhangsan.tke-cloud.com |
| duration_type | VARCHAR(20) | 使用时长（long/temporary） | long |
| expire_at | TIMESTAMP | 到期时间（临时实例） | NULL |
| created_at | TIMESTAMP | 创建时间 | 2026-03-04 10:15:00 |
| deleted_at | TIMESTAMP | 删除时间 | NULL |

#### 表 3：applications（申请记录表）

| 字段名 | 类型 | 说明 | 示例 |
|-------|------|------|------|
| id | SERIAL PRIMARY KEY | 申请 ID | 20260304001 |
| user_id | INT REFERENCES users(id) | 申请人 | 1 |
| instance_name | VARCHAR(50) | 申请的实例名称 | openclaw-zhangsan |
| spec | VARCHAR(20) | 申请规格 | standard |
| reason | TEXT | 申请理由 | 用于 TKE 集群日常运维 |
| status | VARCHAR(20) | 状态（pending/approved/rejected） | approved |
| approver_id | INT REFERENCES users(id) | 审批人 | 2 |
| approve_note | TEXT | 审批备注 | 已确认用途合理，批准 |
| approved_at | TIMESTAMP | 审批时间 | 2026-03-04 10:10:00 |
| created_at | TIMESTAMP | 申请时间 | 2026-03-04 10:00:00 |

#### 表 4：cost_records（成本记录表）

| 字段名 | 类型 | 说明 | 示例 |
|-------|------|------|------|
| id | SERIAL PRIMARY KEY | 记录 ID | 1 |
| instance_id | INT REFERENCES instances(id) | 实例 ID | 1001 |
| user_id | INT REFERENCES users(id) | 用户 ID | 1 |
| department | VARCHAR(100) | 部门 | 研发部-运维组 |
| month | VARCHAR(7) | 月份（YYYY-MM） | 2026-03 |
| spec_cost | DECIMAL(10,2) | 规格费用 | 500.00 |
| api_calls | INT | API 调用次数 | 8000 |
| api_cost | DECIMAL(10,2) | API 超额费用 | 30.00 |
| total_cost | DECIMAL(10,2) | 总费用 | 530.00 |
| created_at | TIMESTAMP | 记录时间 | 2026-04-01 00:00:00 |

#### 表 5：audit_logs（审计日志表）

| 字段名 | 类型 | 说明 | 示例 |
|-------|------|------|------|
| id | SERIAL PRIMARY KEY | 日志 ID | 1 |
| user_id | INT REFERENCES users(id) | 操作用户 | 1 |
| action | VARCHAR(50) | 操作类型（create/delete/approve） | approve |
| resource_type | VARCHAR(50) | 资源类型（instance/application） | application |
| resource_id | VARCHAR(100) | 资源 ID | 20260304001 |
| ip | VARCHAR(50) | 操作 IP | 192.168.1.100 |
| user_agent | TEXT | User-Agent | Mozilla/5.0... |
| result | VARCHAR(20) | 操作结果（success/failed） | success |
| created_at | TIMESTAMP | 操作时间 | 2026-03-04 10:10:00 |

---

### 7.2 数据埋点方案

#### 前端埋点（关键用户行为）

```javascript
// 1. 申请提交
track('instance_apply_submit', {
  user_id: 'zhangsan',
  department: '研发部',
  spec: 'standard',
  duration: 'long',
  wechat_bot_enabled: true,
  timestamp: '2026-03-04T10:00:00Z'
});

// 2. 审批操作
track('instance_approve', {
  admin_id: 'lijingli',
  applicant_id: 'zhangsan',
  instance_name: 'openclaw-zhangsan',
  action: 'approve',  // approve / reject
  timestamp: '2026-03-04T10:10:00Z'
});

// 3. 实例操作
track('instance_operation', {
  user_id: 'zhangsan',
  instance_id: 'inst-12345',
  operation: 'delete',  // start / stop / restart / delete
  timestamp: '2026-03-04T11:00:00Z'
});

// 4. 页面浏览
track('page_view', {
  user_id: 'zhangsan',
  page: '/my-instances',
  referrer: '/dashboard',
  duration_seconds: 120,
  timestamp: '2026-03-04T10:00:00Z'
});
```

---

## 八、竞品参考

### 8.1 竞品分析

| 竞品 | 产品定位 | 核心优势 | 劣势 | 可借鉴点 |
|-----|---------|---------|------|---------|
| **Jupyter Hub** | 企业级 Jupyter Notebook 管理平台 | • 成熟的多租户管理<br>• 灵活的资源配置<br>• 完善的用户权限体系 | • 仅支持 Jupyter Notebook<br>• UI 较为简陋 | • 自助申请流程<br>• 资源配额管理<br>• Spawner 机制（按需创建实例） |
| **GitLab SaaS** | 企业级 Git 仓库管理平台 | • 一键创建 GitLab 实例<br>• 自动备份和升级<br>• 丰富的监控和告警 | • 主要面向代码管理，不适用于 AI 场景 | • 统一升级和灰度发布<br>• 成本分摊和报表<br>• Admin Area 设计 |
| **Databricks Workspace** | 企业级数据科学平台 | • 优秀的用户体验<br>• 自动扩缩容<br>• 集成 Skill 市场 | • 价格昂贵（$$$）<br>• 绑定 Databricks 生态 | • Workspace 隔离模型<br>• Skill/Library 管理<br>• 成本透明化 |
| **AWS SageMaker Studio** | 企业级 ML 开发环境 | • 完善的资源管理<br>• 按需付费<br>• 与 AWS 服务深度集成 | • 学习曲线陡峭<br>• 锁定 AWS 生态 | • 实例规格选择界面<br>• 成本预估工具<br>• 自动化运维能力 |

---

## 九、风险与依赖

### 9.1 风险点

| 风险类型 | 风险描述 | 影响程度 | 应对措施 | 责任人 |
|---------|---------|---------|---------|--------|
| **技术风险** | TKE API 不稳定，导致部署失败率高 | 🔴 高 | • 增加重试机制（3 次）<br>• 部署失败时人工介入<br>• 与 TKE 团队建立快速响应通道 | 技术负责人 |
| **技术风险** | 单集群资源不足，无法创建新实例 | 🟡 中 | • 提前监控集群资源使用率<br>• 资源使用率 > 80% 时告警<br>• 准备多集群部署方案 | 技术负责人 |
| **业务风险** | 员工滥用配额，成本失控 | 🟡 中 | • 设置部门配额和个人配额<br>• 监控异常使用（如单实例 API 调用量异常）<br>• 自动休眠长期未使用的实例 | 产品经理 |
| **业务风险** | 管理员审批不及时，影响用户体验 | 🟡 中 | • 设置审批超时自动催办<br>• 配置自动审批规则（基础版 + 配额充足）<br>• 增加审批 SLA 指标 | 产品经理 |
| **安全风险** | 多租户隔离不彻底，数据泄露 | 🔴 高 | • Namespace + RBAC + NetworkPolicy 三重隔离<br>• 定期进行渗透测试<br>• 引入第三方安全审计 | 安全负责人 |
| **合规风险** | 审计日志不完整，无法通过等保审核 | 🟡 中 | • 记录所有 API 调用和操作日志<br>• 日志保留 180 天<br>• 提供日志导出和搜索功能 | 合规负责人 |
| **运营风险** | 用户留存率低，产品价值未被认可 | 🟡 中 | • MVP 阶段深度服务 10 家客户<br>• 每周收集用户反馈<br>• 快速迭代核心功能 | 产品经理 |

### 9.2 外部依赖

| 依赖项 | 依赖内容 | 关键程度 | 风险 | 应对措施 |
|-------|---------|---------|------|---------|
| **TKE API** | 创建 Namespace、Deployment、Service、Ingress | 🔴 关键 | TKE API 变更导致平台不可用 | • 与 TKE 团队建立 API 变更通知机制<br>• 适配 TKE 多个版本 |
| **TKE 监控服务** | Prometheus 集成，获取实例 CPU/内存数据 | 🟡 重要 | 监控数据不准确 | • 增加自定义监控采集<br>• 降级方案：仅展示 K8s 原生指标 |
| **TKE 日志服务（CLS）** | 集中收集所有实例日志 | 🟡 重要 | CLS 故障导致日志丢失 | • 本地缓存日志<br>• 降级方案：直接查询 Pod 日志 |
| **企业微信 API** | 发送通知、配置 Bot | 🟡 重要 | 企业微信 API 限流 | • 控制通知频率<br>• 降级方案：邮件通知 |
| **LDAP/SSO** | 用户登录和组织架构同步 | 🔴 关键 | LDAP 服务故障导致无法登录 | • 缓存用户信息<br>• 降级方案：本地账号登录 |
| **OpenClaw 官方镜像** | OpenClaw Docker 镜像（含 Gateway + Agent + Skills） | 🔴 关键 | 镜像漏洞或不稳定 | • 自建镜像仓库（TCR）<br>• 镜像安全扫描<br>• 固定镜像版本，避免自动升级 |

---

## 十、排期建议

### 10.1 迭代计划

#### 阶段 1：MVP 版本（M1 - M3，0-3 个月）

**目标**：验证核心流程，完成 10 家企业 POC

| 里程碑 | 时间 | 关键交付物 | 验收标准 |
|-------|------|-----------|---------|
| **M1** | 第 1 个月 | • 员工申请流程<br>• 管理员审批流程<br>• 基础部署能力 | • 员工可提交申请<br>• 管理员可审批<br>• 审批通过后自动创建实例（成功率 > 80%） |
| **M2** | 第 2 个月 | • 我的实例列表<br>• 全局监控大盘<br>• 健康检查和告警 | • 员工可查看自己的实例<br>• 管理员可查看所有实例<br>• 异常实例自动告警 |
| **M3** | 第 3 个月 | • 成本统计<br>• 配额管理<br>• 审计日志 | • 按部门统计成本<br>• 配额超限时禁止申请<br>• 所有操作记录审计日志 |

**POC 目标客户**：
- 腾讯内部 3 个业务线（PCG、IEG、CSIG）
- 外部客户 7 家（互联网公司，技术团队 50-200 人）

---

#### 阶段 2：产品化版本（M4 - M6，3-6 个月）

**目标**：完善产品功能，获得 50 家付费客户

| 里程碑 | 时间 | 关键交付物 | 验收标准 |
|-------|------|-----------|---------|
| **M4** | 第 4 个月 | • 配置管理（Skills、环境变量）<br>• 详细成本分析（导出报表） | • 员工可管理 Skills<br>• 管理员可导出成本报表 |
| **M5** | 第 5 个月 | • 版本管理（统一升级、灰度发布）<br>• 自动备份恢复 | • 管理员可一键升级所有实例<br>• 支持灰度发布（先升级 10% 实例） |
| **M6** | 第 6 个月 | • 数据加密<br>• 多集群支持<br>• 完善的监控和告警 | • PVC 数据加密存储<br>• 支持跨集群部署实例 |

---

#### 阶段 3：规模化版本（M7 - M12，6-12 个月）

**目标**：规模化增长，实现 200 家客户，5000 个活跃实例

| 里程碑 | 时间 | 关键交付物 | 验收标准 |
|-------|------|-----------|---------|
| **M7** | 第 7-8 个月 | • 自动扩缩容<br>• Skill 市场（Beta） | • 实例根据负载自动调整资源<br>• 管理员可上传企业 Skills |
| **M8** | 第 9-10 个月 | • 多云支持（TKE + 阿里云 ACK）<br>• 高级监控（链路追踪） | • 支持阿里云 ACK 集群<br>• 集成 Jaeger 链路追踪 |
| **M9** | 第 11-12 个月 | • Skill 市场（正式版）<br>• 自助式故障诊断<br>• 企业定制化配置 | • 100+ 企业 Skills<br>• 用户可自助排查常见问题 |

---

## 十一、商业化方案

### 11.1 计费模式

#### 方案 1：按实例规格计费（月付/年付）

| 规格 | CPU/内存 | 月价格 | 年价格（8 折） | 适用场景 | 免费 API 额度 |
|-----|---------|-------|---------------|---------|--------------|
| **基础版** | 1C2G | ¥200 | ¥1,920 | 轻度使用（开发者） | 1000 次/月 |
| **标准版** | 2C4G | ¥500 | ¥4,800 | 日常运维（SRE） | 5000 次/月 |
| **企业版** | 4C8G | ¥1000 | ¥9,600 | 重度使用（核心团队） | 20000 次/月 |

#### 方案 2：按使用量计费（超出免费额度后）

| 计费项 | 单价 | 备注 |
|-------|------|------|
| **API 调用次数** | ¥0.01/次 | 超出免费额度后收费 |
| **存储费用** | ¥0.5/GB/月 | 超出 10GB 后收费 |
| **流量费用** | ¥0.8/GB | 外网出流量 |

#### 企业套餐（推荐）

| 套餐 | 实例数 | 规格组合 | 年价格 | 目标客户 |
|-----|--------|---------|-------|---------|
| **创业套餐** | 10 个 | 5 基础版 + 5 标准版 | ¥36,000/年 | 50-100 人团队 |
| **标准套餐** | 50 个 | 10 基础版 + 30 标准版 + 10 企业版 | ¥150,000/年 | 200-500 人团队 |
| **旗舰套餐** | 100 个 | 20 基础版 + 50 标准版 + 30 企业版 | ¥280,000/年 | 500-1000 人团队 |

---

### 11.2 成本构成分析

#### 单实例成本（标准版 2C4G）

| 成本项 | 月成本 | 年成本 | 说明 |
|-------|--------|--------|------|
| **TKE 计算资源** | ¥150 | ¥1,800 | 2C4G Pod，按需计费 |
| **TKE 存储（CBS）** | ¥30 | ¥360 | 10GB 云硬盘 |
| **流量费用** | ¥20 | ¥240 | 外网出流量（预估） |
| **监控/日志服务** | ¥10 | ¥120 | Prometheus + CLS |
| **运维成本** | ¥40 | ¥480 | 人力成本摊销 |
| **合计** | ¥250 | ¥3,000 | 毛利率 50%（售价 ¥500/月） |

---

### 11.3 收入预测

#### 12 个月收入预测

| 月份 | 企业客户数 | 活跃实例数 | 月收入 | 累计 ARR |
|-----|-----------|-----------|-------|---------|
| M1-M3 | 10 | 50 | ¥25,000 | ¥300,000 |
| M4-M6 | 50 | 500 | ¥250,000 | ¥3,000,000 |
| M7-M9 | 120 | 2000 | ¥1,000,000 | ¥12,000,000 |
| M10-M12 | 200 | 5000 | ¥2,500,000 | ¥30,000,000 |

**年度目标**：ARR ¥3000 万，毛利率 50%，净利润 ¥1000 万

---

## 十二、待确认问题

### 12.1 需要与业务方确认的问题

1. **企业身份体系集成深度**
   - MVP 阶段是否必须支持 LDAP 组织架构同步？
   - 还是先做简单的 SSO 登录 + 手动维护部门信息？

2. **实例生命周期管理策略**
   - 临时实例到期后是否需要数据备份？备份保留多久？
   - 长期实例的"休眠"策略：多少天未使用触发休眠？

3. **初期 POC 客户名单**
   - 腾讯内部哪 3 个业务线作为 POC 客户？
   - 外部 7 家客户的行业分布（互联网/金融/制造业）？

### 12.2 需要与技术团队确认的问题

1. **TKE 集群资源规划**
   - 当前 TKE 集群的资源上限（CPU/内存/存储）？
   - 是否需要在多个地域部署（北京/上海/广州）？

2. **OpenClaw 镜像定制**
   - OpenClaw 官方镜像是否支持多租户模式？
   - 需要定制哪些企业级功能（如 LDAP 集成、审计日志）？

3. **监控和日志方案**
   - TKE 监控服务（Prometheus）的数据保留周期？
   - TKE 日志服务（CLS）的日志存储成本？

### 12.3 需要与设计团队确认的问题

1. **前端设计规范**
   - 是否复用腾讯云控制台的设计规范（TDesign）？
   - 移动端适配的优先级（MVP 是否包含）？

2. **用户体验细节**
   - 申请表单的字段是否需要"保存草稿"功能？
   - 审批流程是否需要支持"转审"（转给其他管理员）？

---

## 附录

### A. 关键术语表

| 术语 | 英文 | 说明 |
|-----|------|------|
| **OpenClaw** | - | 新一代 AI 运维助手，支持自然语言执行运维任务 |
| **富容器模式** | All-in-One Container | 将 Gateway、Agent、Skills、Memory 打包在一个容器中 |
| **TKE** | Tencent Kubernetes Engine | 腾讯云容器服务 |
| **Namespace** | - | K8s 命名空间，用于资源隔离 |
| **RBAC** | Role-Based Access Control | 基于角色的访问控制 |
| **NetworkPolicy** | - | K8s 网络策略，用于网络隔离 |
| **PVC** | PersistentVolumeClaim | K8s 持久化存储声明 |
| **Ingress** | - | K8s 外部访问入口，负责域名和 HTTPS 配置 |
| **ARR** | Annual Recurring Revenue | 年度经常性收入 |
| **NPS** | Net Promoter Score | 净推荐值，衡量用户满意度 |

---

### B. 参考文档

1. **OpenClaw 官方文档**：https://openclaw.dev/docs
2. **TKE API 文档**：https://cloud.tencent.com/document/product/457
3. **Kubernetes RBAC 文档**：https://kubernetes.io/docs/reference/access-authn-authz/rbac/
4. **Prometheus 监控方案**：https://prometheus.io/docs/
5. **企业微信 Bot API**：https://work.weixin.qq.com/api/doc

---

### C. 联系方式

| 角色 | 姓名 | 邮箱 | 企业微信 |
|-----|------|------|---------|
| **产品经理** | [待填写] | pm@company.com | @pm |
| **技术负责人** | [待填写] | tech@company.com | @tech |
| **安全负责人** | [待填写] | security@company.com | @security |
| **设计负责人** | [待填写] | design@company.com | @design |

---

**文档结束**

**下一步动作**：
1. 召开需求评审会（参与人：产品、研发、设计、测试、安全）
2. 确认 MVP 功能范围和排期
3. 启动 M1 开发（预计 2026-03-11 开始）

**（本 PRD 已完整，包含 Part 1 和 Part 2 所有内容）**
