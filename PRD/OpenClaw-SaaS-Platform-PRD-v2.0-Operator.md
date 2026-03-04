# OpenClaw SaaS 平台产品需求文档 v2.0（基于 Operator）

## 文档信息
- **版本**：V2.0（集成 OpenClaw K8s Operator）
- **基于**：V1.0 PRD
- **修订日期**：2026-03-04
- **状态**：待评审
- **产品代号**：OpenClaw Cloud Platform (OCP)
- **重要变更**：采用 OpenClaw K8s Operator 简化部署和运维

---

## 📋 版本变更说明

### V2.0 核心变更

**采用 OpenClaw K8s Operator 的理由**：
- ✅ **简化部署**：从 9 个手动步骤减少到 3 个自动步骤
- ✅ **降低维护成本**：后端代码减少 85%（200+ 行 → 30 行）
- ✅ **增强安全性**：内置安全加固（非Root、Seccomp、只读文件系统）
- ✅ **自动化运维**：自动备份、自动升级、自动回滚
- ✅ **提升可靠性**：部署成功率从 80% 提升到 99%+

---

## 一、需求背景

### 1.1 背景描述

随着企业数字化转型加速，越来越多的企业采用容器化技术（如 TKE）来管理云原生应用。然而，传统的运维方式面临三大挑战：

1. **运维复杂度高**：SRE 需要掌握 Kubernetes、监控、日志分析等多种技能，学习曲线陡峭
2. **工具碎片化**：企业内部存在多个运维工具（监控系统、发布平台、配置中心），切换成本高
3. **协作效率低**：开发、测试、运维团队使用不同的工具，信息传递存在断层

**OpenClaw** 作为新一代 AI 运维助手，能够通过自然语言理解用户意图，自动执行运维任务（如扩容、日志分析、故障排查），显著降低运维门槛。但企业在部署 OpenClaw 时面临新的问题：

- **部署困难**：需要懂 Kubernetes、Helm Chart、微服务架构才能部署
- **运维负担**：每个团队独立部署，导致版本不一致、安全配置参差不齐
- **成本失控**：缺乏统一的资源管理和成本监控，容易造成资源浪费
- **安全合规风险**：多租户场景下，数据隔离和权限管理复杂

### 1.2 问题定义

**核心问题**：企业如何让数百名员工（SRE、开发者、测试工程师）快速、安全、低成本地使用 OpenClaw？

**关键矛盾**：
- **员工视角**："我只想要一个能用的 OpenClaw，不想学 K8s"
- **IT 管理员视角**："我要保证安全合规、成本可控、版本统一，不能让员工随意部署"

### 1.3 业务目标

#### 短期目标（6 个月）
1. **产品验证**：完成 MVP 版本，支持 10 家企业 POC（每家 5-10 个用户）
2. **商业化启动**：获得 50 家付费客户，ARR 达到 ¥500 万
3. **技术验证**：证明基于 Operator 的自动化部署能力，实现 **2 分钟内创建实例**（优化 60%）

#### 长期目标（12 个月）
1. **规模化增长**：200 家企业客户，5000 个活跃实例，ARR ¥3000 万
2. **生态建设**：建立 Skill 市场，沉淀 100+ 企业级运维 Skills
3. **行业标准**：成为 "AI 运维助手 SaaS 化" 的行业标杆产品

#### 业务价值
- **对企业客户**：降低 80% 的部署和运维成本，提升 3 倍运维效率
- **对 TKE 平台**：增加 TKE 的使用粘性，带动容器化业务增长
- **对腾讯云**：新的 SaaS 收入来源，年收入潜力 ¥1 亿+

---

## 二、用户分析

### 2.1 目标用户

#### 用户类型 1：企业员工（主要使用者）

| 维度 | 描述 |
|-----|------|
| **典型角色** | SRE 工程师、后端开发者、测试工程师、DevOps 工程师 |
| **年龄/工作年限** | 25-35 岁，工作 2-8 年 |
| **技术背景** | 熟悉 Linux 和基础运维命令，对 K8s 了解程度不一（20% 精通，50% 基础，30% 不懂） |
| **日常工作** | 部署应用、排查故障、扩容缩容、查看日志和监控、发布变更 |
| **核心诉求** | "给我一个能用的 AI 助手，别让我学 K8s" |
| **痛点** | 每次部署新工具都需要找运维帮忙，等待时间长；学习新工具成本高 |

**典型用户画像：张三（SRE 工程师）**
> 张三负责公司电商业务的容器化运维，每天需要处理 10+ 次扩容、20+ 次日志分析、5+ 次故障排查。之前需要登录多个系统（TKE 控制台、监控平台、日志平台）才能完成一个任务，效率很低。听说 OpenClaw 能通过聊天完成运维任务，想尝试，但不知道怎么部署。

#### 用户类型 2：企业 IT 管理员（运营管理者）

| 维度 | 描述 |
|-----|------|
| **典型角色** | IT 运维经理、云平台负责人、安全合规官、成本管理者 |
| **年龄/工作年限** | 30-45 岁，工作 5-15 年 |
| **技术背景** | 精通企业 IT 架构、K8s、安全合规、成本管理 |
| **日常工作** | 审批员工的资源申请、监控系统健康、控制成本、应对安全审计 |
| **核心诉求** | "让员工能自助使用，但我要有全局控制力" |
| **痛点** | 员工自己部署工具容易出安全问题；成本难以分摊到部门；工具升级困难 |

**典型用户画像：李经理（IT 运维经理）**
> 李经理管理 200 人的技术团队，每周收到 30+ 个工具申请（Jenkins、GitLab、监控系统等）。之前的做法是运维团队手动部署，每个申请需要 2-3 天。现在希望让员工自助申请，但担心成本失控和安全风险。

### 2.2 使用场景

#### 场景 1：新员工快速获取 OpenClaw 实例（高频场景）

**参与角色**：张三（新入职的 SRE）、李经理（IT 管理员）

**场景流程**：

```
1. 员工申请（3 分钟）
   ├─ 张三登录平台（SSO 单点登录）
   ├─ 点击"申请 OpenClaw"
   ├─ 填写表单：
   │  ├─ 实例名称：openclaw-zhangsan
   │  ├─ 用途说明："用于 TKE 集群日常运维"
   │  ├─ 规格选择：标准版（2C4G）
   │  └─ 使用时长：长期
   └─ 提交申请

2. 管理员审批（10 分钟）
   ├─ 李经理收到企业微信通知
   ├─ 打开审批工作台，查看申请详情
   └─ 点击"批准"

3. 自动部署（2 分钟）⚡ 优化 60%
   ├─ 后端创建 OpenClawInstance CR
   ├─ Operator 自动协调创建所有资源
   └─ 健康检查：等待 Pod Running

4. 开始使用（1 分钟）
   ├─ 张三收到企业微信通知："你的 OpenClaw 已就绪！"
   ├─ 访问地址：https://openclaw-zhangsan.tke-cloud.com
   └─ 在企业微信里 @OpenClaw 开始使用
```

**关键成功因素**：
- 申请到使用全程 < 15 分钟（原 30 分钟）
- 无需填写复杂的技术参数
- 自动配置企业微信集成

---

## 三、需求概述

### 3.1 需求描述

**产品定位**：OpenClaw-as-a-Service（OaaS） —— 企业级 AI 运维助手 SaaS 管理平台（基于 Operator）

**一句话描述**：让企业员工像申请云主机一样，一键获取属于自己的 OpenClaw AI 运维助手，由 Operator 自动化管理，IT 管理员可统一控制。

**核心能力**：
1. **自助申请**：员工填写简单表单，**2 分钟**获得可用的 OpenClaw 实例
2. **声明式部署**：通过 OpenClawInstance CR 管理实例，Operator 自动协调所有资源
3. **统一管理**：管理员一个控制台管理所有实例（审批、监控、升级、成本）
4. **安全隔离**：每个员工的实例完全隔离（Namespace + RBAC + NetworkPolicy + 容器安全加固）
5. **自动化运维**：自动备份、自动升级、自动回滚、自动扩缩容
6. **成本可控**：按实例规格计费，支持配额管理和成本分摊

### 3.2 核心价值

#### 对员工（使用者）
- **降低使用门槛**：无需懂 K8s，只需填写表单即可获得 OpenClaw
- **快速交付**：从申请到使用 < 15 分钟（优化 50%）
- **开箱即用**：自动配置企业微信集成，内置企业常用 Skills
- **成本透明**：实时查看自己的使用量和费用

#### 对 IT 管理员（运营者）
- **减少运维工作量**：从人工部署变为 Operator 自动化，节省 90% 时间
- **全局可控**：一个大盘监控所有实例，异常自动告警
- **简化升级流程**：通过 Operator 统一升级，支持灰度发布和自动回滚
- **安全合规**：自动记录审计日志，容器安全加固开箱即用

#### 对企业（决策者）
- **提升运维效率**：员工使用 AI 助手后,运维效率提升 3 倍
- **降低 IT 成本**：减少运维团队的重复性工作，人力成本降低 50%
- **统一工具体系**：避免工具碎片化，降低培训成本

---

## 四、功能需求

### 4.1 功能清单

| 功能模块 | 功能描述 | 优先级 | MVP 包含 | 变更说明 |
|---------|---------|--------|---------|---------|
| **用户门户** |  |  |  |  |
| 用户注册/登录 | SSO 单点登录（支持企业微信、飞书、钉钉、LDAP） | P0 | ✅ | 无变更 |
| 实例申请 | 填写表单申请 OpenClaw 实例（规格、用途、时长） | P0 | ✅ | 后端简化 |
| 我的实例列表 | 查看自己的所有实例（状态、规格、费用） | P0 | ✅ | 无变更 |
| 实例操作 | 启动/停止/重启/删除实例 | P0 | ✅ | 由 Operator 处理 |
| 实例详情 | 查看实例的访问地址、Token、配置信息 | P0 | ✅ | 无变更 |
| Skills 管理 | 管理 Skills（安装/卸载/更新） | P1 | ✅ | 🆕 声明式管理 |
| 自适应配置 | 允许实例自我配置（白名单验证） | P1 | ✅ | 🆕 Operator 新增 |
| 使用统计 | 查看 API 调用量、费用明细、使用趋势 | P1 | ✅ | 无变更 |
| **管理员控制台** |  |  |  |  |
| 审批管理 | 查看待审批申请，批准/拒绝 | P0 | ✅ | 无变更 |
| 全局监控大盘 | 总实例数、活跃实例、成本、异常告警 | P0 | ✅ | 集成 Operator 指标 |
| 实例管理 | 查看所有实例，强制停止/删除异常实例 | P0 | ✅ | 通过 CR 管理 |
| 版本管理 | 统一升级实例，灰度发布 | P1 | ✅ | 🆕 Operator AutoUpdate |
| 备份恢复 | 自动备份和恢复实例数据 | P1 | ✅ | 🆕 Operator 原生支持 |
| 成本分析 | 按部门/用户统计费用，导出报表 | P1 | ✅ | 无变更 |
| 审计日志 | 记录所有操作日志，支持搜索和导出 | P1 | ✅ | 无变更 |
| **自动化运维** |  |  |  |  |
| 一键部署 | 自动在 TKE 创建 OpenClawInstance CR | P0 | ✅ | 简化为单 CR |
| 健康检查 | 定期检查实例健康，自动重启异常实例 | P0 | ✅ | Operator 自动处理 |
| 自动备份 | 定期备份 + 删除前备份 + 升级前备份 | P1 | ✅ | 🆕 S3 兼容存储 |
| 自动升级 | 自动检测新版本并升级 | P1 | ✅ | 🆕 Operator 新增 |
| 自动回滚 | 升级失败自动回滚 | P1 | ✅ | 🆕 Operator 新增 |
| 监控集成 | Prometheus + Grafana + 告警规则 | P1 | ✅ | Operator 自动配置 |

---

### 4.2 核心功能详细说明

#### 功能 1：实例申请流程（员工端）

**功能目标**：让员工通过简单的表单，快速申请一个可用的 OpenClaw 实例

**🆕 V2.0 变更**：
- 后端代码从 200+ 行减少到 30 行
- 部署成功率从 80% 提升到 99%+
- 部署时间从 5 分钟减少到 2 分钟

**操作流程**：

```
步骤 1-2：进入申请页面 + 填写表单（无变更）

步骤 3：提交申请
├─ 后端处理（简化）：
│  ├─ 创建申请记录（状态：待审批）
│  └─ 发送企业微信通知给管理员
└─ 无需预检查 TKE 资源（Operator 会处理）

步骤 4：等待审批（无变更）

步骤 5：自动部署（2 分钟，优化 60%）
└─ 后端调用 K8s API：
   ├─ 创建 Namespace
   ├─ 创建 OpenClawInstance CR（包含所有配置）
   └─ Operator 自动协调创建所有资源

步骤 6：开始使用
└─ 企业微信通知："你的 OpenClaw 已就绪！🎉"
```

---

#### 功能 2：声明式实例管理（核心变更）

**功能目标**：通过 OpenClawInstance CR 管理实例生命周期

**OpenClawInstance CR 示例**：

```yaml
apiVersion: openclaw.rocks/v1alpha1
kind: OpenClawInstance
metadata:
  name: zhangsan
  namespace: openclaw-zhangsan
  labels:
    department: research-dev
    user: zhangsan
spec:
  # 1. 规格配置
  resources:
    requests:
      cpu: "2"
      memory: "4Gi"
    limits:
      cpu: "2"
      memory: "4Gi"
  
  # 2. 存储配置
  storage:
    persistence:
      enabled: true
      size: 10Gi
      storageClassName: tke-cbs
    orphan: true  # 删除实例时保留数据
  
  # 3. 企业微信集成
  envFrom:
    - secretRef:
        name: wechat-secret
  
  # 4. 网络配置
  networking:
    ingress:
      enabled: true
      host: openclaw-zhangsan.tke-cloud.com
      tls:
        enabled: true
        issuer: letsencrypt-prod
    service:
      type: ClusterIP
  
  # 5. 监控配置
  observability:
    metrics:
      enabled: true
      serviceMonitor: true
      prometheusRule: true
      grafanaDashboard: true
  
  # 6. 自动备份
  autoUpdate:
    backup:
      enabled: true
      onDelete: true  # 删除前备份
      schedule: "0 2 * * *"  # 每天凌晨2点
      s3:
        endpoint: "https://cos.ap-guangzhou.myqcloud.com"
        bucket: "openclaw-backups"
        secretRef:
          name: s3-credentials
  
  # 7. Skills 预装
  skills:
    - source: "clawhub:tke-toolkit@latest"
    - source: "clawhub:log-analyzer@latest"
    - source: "clawhub:cost-optimizer@latest"
  
  # 8. 自我配置能力
  selfConfigure:
    enabled: true
    allowedActions:
      - skills  # 允许安装 Skills
      - envVars  # 允许修改环境变量
      - workspaceFiles  # 允许上传配置文件
```

**Operator 自动管理的资源**：
- ✅ StatefulSet（主容器 + Nginx 网关）
- ✅ PVC（自动绑定 TKE CBS）
- ✅ Service + Ingress（自动申请 HTTPS 证书）
- ✅ NetworkPolicy（默认拒绝 + 白名单）
- ✅ RBAC（ServiceAccount + Role + RoleBinding）
- ✅ Gateway Token Secret（自动生成）
- ✅ ServiceMonitor + PrometheusRule（监控告警）

---

#### 功能 3：自动版本管理（管理员端）

**功能目标**：统一升级所有实例，支持灰度发布和自动回滚

**🆕 V2.0 新增功能**：

```yaml
# 管理员控制台配置全局升级策略
apiVersion: openclaw.rocks/v1alpha1
kind: OpenClawInstanceTemplate
metadata:
  name: global-upgrade-policy
spec:
  # 自动检测新版本
  autoUpdate:
    enabled: true
    interval: 1h  # 每小时检查一次
    
    # 灰度策略
    canary:
      enabled: true
      stages:
        - name: "test-env"
          selector:
            matchLabels:
              env: test
          percentage: 100%  # 测试环境全量升级
          waitDuration: 3d  # 观察 3 天
          
        - name: "prod-canary"
          selector:
            matchLabels:
              env: production
          percentage: 10%  # 生产环境先升级 10%
          waitDuration: 1d
          
        - name: "prod-full"
          percentage: 100%  # 全量升级
    
    # 回滚策略
    rollback:
      enabled: true
      trigger:
        - type: "crashLoop"  # Pod 崩溃循环
          threshold: 3
        - type: "errorRate"  # 错误率过高
          threshold: 5%
      
    # 升级前自动备份
    backup:
      enabled: true
      onUpgrade: true
```

**收益**：
- ✅ 自动检测新版本（无需人工检查）
- ✅ 声明式灰度策略（自动执行）
- ✅ 自动回滚（升级失败时）
- ✅ 升级前自动备份（数据安全）

---

## 五、技术方案与部署架构

### 5.1 整体架构设计

#### 5.1.1 系统架构图（v2.0）

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
│                            │  │ • 实例编排服务 🆕 简化85%      │  │
│                            │  │   └─ 创建 OpenClawInstance CR │  │
│                            │  │ • 监控数据聚合服务               │  │
│                            │  │ • 成本计算服务                  │  │
│                            │  │ • 通知服务（企业微信/邮件）      │  │
│                            │  └──────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                   🆕 OpenClaw Operator 层                        │
├─────────────────────────────────────────────────────────────────┤
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  OpenClaw Operator                                        │   │
│  ├──────────────────────────────────────────────────────────┤   │
│  │  监听 OpenClawInstance CR 变化                             │   │
│  │  ├─ 创建 StatefulSet + PVC                                │   │
│  │  ├─ 创建 Service + Ingress                                │   │
│  │  ├─ 创建 NetworkPolicy + RBAC                             │   │
│  │  ├─ 创建 ServiceMonitor + PrometheusRule                  │   │
│  │  ├─ 自动备份（S3）                                         │   │
│  │  ├─ 自动升级 + 回滚                                        │   │
│  │  └─ 健康检查 + 自愈                                        │   │
│  └──────────────────────────────────────────────────────────┘   │
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
│  OpenClaw 实例（由 Operator 自动管理）                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │ Namespace:   │  │ Namespace:   │  │ Namespace:   │         │
│  │ openclaw-    │  │ openclaw-    │  │ openclaw-    │         │
│  │ zhangsan     │  │ lisi         │  │ wangwu       │         │
│  ├──────────────┤  ├──────────────┤  ├──────────────┤         │
│  │ StatefulSet: │  │ StatefulSet: │  │ StatefulSet: │         │
│  │ • Gateway    │  │ • Gateway    │  │ • Gateway    │         │
│  │ • Agent      │  │ • Agent      │  │ • Agent      │         │
│  │ • Skills     │  │ • Skills     │  │ • Skills     │         │
│  │ • Memory     │  │ • Memory     │  │ • Memory     │         │
│  ├──────────────┤  ├──────────────┤  ├──────────────┤         │
│  │ 🆕 安全加固： │  │ 🆕 安全加固： │  │ 🆕 安全加固： │         │
│  │ • 非Root运行 │  │ • 非Root运行 │  │ • 非Root运行 │         │
│  │ • 只读文件系统│  │ • 只读文件系统│  │ • 只读文件系统│         │
│  │ • Seccomp   │  │ • Seccomp   │  │ • Seccomp   │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
└─────────────────────────────────────────────────────────────────┘
```

**关键变更**：
1. 新增 **OpenClaw Operator 层**：负责实例生命周期管理
2. 后端服务简化：从直接管理 9+ 个 K8s 资源 → 创建单个 CR
3. 容器安全加固：Operator 自动应用安全最佳实践

---

### 5.2 部署架构详解

#### 5.2.1 OpenClaw Operator 安装

**前置条件**：
- TKE 集群版本 ≥ 1.20
- Helm 3.x
- cert-manager（用于自动 HTTPS 证书）

**安装步骤**：

```bash
# 1. 安装 cert-manager（如果未安装）
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# 2. 添加 OpenClaw Helm Repo
helm repo add openclaw https://openclaw-rocks.github.io/helm-charts
helm repo update

# 3. 安装 OpenClaw Operator
helm install openclaw-operator openclaw/openclaw-operator \
  --namespace openclaw-system \
  --create-namespace \
  --set cloudProvider=tencent \
  --set tke.clusterEndpoint=https://tke-api.tke-cloud.com \
  --set metrics.enabled=true \
  --set webhook.enabled=true

# 4. 验证安装
kubectl get pods -n openclaw-system
# 输出：
# NAME                                  READY   STATUS    RESTARTS   AGE
# openclaw-operator-7c8d9f7b5d-xxxxx   1/1     Running   0          1m
```

---

#### 5.2.2 自动化部署流程（2 分钟内完成，优化 60%）

```
管理员批准申请
       ▼
┌──────────────────────────────────────────────────────────────┐
│  步骤 1：创建 Namespace + OpenClawInstance CR（10 秒）        │
│  ├─ 后端调用 K8s API 创建 Namespace                           │
│  ├─ 创建 OpenClawInstance CR（包含所有配置）                  │
│  └─ Operator 自动协调创建所有资源                             │
└──────────────────────────────────────────────────────────────┘
       ▼
┌──────────────────────────────────────────────────────────────┐
│  步骤 2：等待 Operator 完成部署（1.5 分钟）                   │
│  ├─ Operator 自动创建：                                       │
│  │  ├─ StatefulSet（主容器 + Nginx 网关）                     │
│  │  ├─ PVC（自动绑定 TKE CBS）                                │
│  │  ├─ Service + Ingress（自动申请 HTTPS 证书）               │
│  │  ├─ NetworkPolicy（默认拒绝 + 白名单）                     │
│  │  ├─ RBAC（ServiceAccount + Role + RoleBinding）           │
│  │  ├─ Gateway Token Secret（自动生成）                       │
│  │  └─ ServiceMonitor + PrometheusRule                        │
│  ├─ Init Container 自动安装 Skills                            │
│  └─ 等待 Pod Running + 健康检查                               │
└──────────────────────────────────────────────────────────────┘
       ▼
┌──────────────────────────────────────────────────────────────┐
│  步骤 3：发送通知（10 秒）                                    │
│  ├─ 企业微信通知："你的 OpenClaw 已就绪！"                     │
│  ├─ 邮件通知：包含访问地址和快速入门指南                        │
│  └─ 更新平台状态：实例状态 = Running                           │
└──────────────────────────────────────────────────────────────┘

总耗时：约 2 分钟（P50），< 3 分钟（P95）
原流程：约 5 分钟（P50），< 10 分钟（P95）
优化：60% 时间减少
```

**后端代码简化示例**：

```go
// 原方案：需要创建 9+ 个资源（200+ 行代码）
func DeployInstance(req *DeployRequest) error {
    // 1. 创建 Namespace（20 行）
    // 2. 创建 PVC（30 行）
    // 3. 创建 Secret（15 行）
    // 4. 创建 Deployment（60 行）
    // 5. 创建 Service（20 行）
    // 6. 创建 Ingress（30 行）
    // 7. 创建 NetworkPolicy（25 行）
    // 8. 健康检查循环（30 行）
    // ... 200+ 行代码
}

// 🆕 新方案：只需创建 1 个 CR（30 行代码）
func DeployInstance(req *DeployRequest) error {
    instance := &openclawv1alpha1.OpenClawInstance{
        ObjectMeta: metav1.ObjectMeta{
            Name:      req.Username,
            Namespace: fmt.Sprintf("openclaw-%s", req.Username),
            Labels: map[string]string{
                "department": req.Department,
                "user":       req.Username,
            },
        },
        Spec: openclawv1alpha1.OpenClawInstanceSpec{
            Resources: req.ToResourceRequirements(),
            Storage:   req.ToStorageSpec(),
            EnvFrom:   req.ToEnvFrom(),
            Networking: openclawv1alpha1.NetworkingSpec{
                Ingress: openclawv1alpha1.IngressSpec{
                    Enabled: true,
                    Host:    fmt.Sprintf("openclaw-%s.tke-cloud.com", req.Username),
                    TLS:     openclawv1alpha1.TLSSpec{Enabled: true},
                },
            },
            Observability: openclawv1alpha1.ObservabilitySpec{
                Metrics: openclawv1alpha1.MetricsSpec{
                    Enabled:           true,
                    ServiceMonitor:    true,
                    PrometheusRule:    true,
                    GrafanaDashboard:  true,
                },
            },
        },
    }
    
    // 一行代码创建，Operator 自动协调
    return k8sClient.Create(ctx, instance)
}
```

**收益**：
- ✅ 代码量减少 85%（200+ 行 → 30 行）
- ✅ 降低维护成本
- ✅ 减少 Bug 风险
- ✅ 自动处理边界情况

---

### 5.3 多租户安全与隔离

#### 5.3.1 五层隔离机制（新增容器安全层）

| 隔离层级 | 隔离方式 | 作用 | 实现细节 |
|---------|---------|------|---------|
| **1. Namespace 隔离** | K8s Namespace | 逻辑隔离，不同用户的资源互不可见 | 每个实例独立 Namespace（openclaw-{username}） |
| **2. RBAC 权限隔离** | K8s RBAC | 员工只能操作自己的实例 | • 员工角色：仅能访问自己的 Namespace<br>• 管理员角色：可访问所有 Namespace |
| **3. 网络隔离** | NetworkPolicy | 不同实例间网络不互通 | • Ingress：仅允许 Nginx Ingress 访问<br>• Egress：仅允许访问 TKE API、企业微信 API |
| **4. 数据隔离** | 独立 PVC | 每个实例的数据独立存储 | 每个实例挂载独立的 TKE CBS 云硬盘 |
| **🆕 5. 容器安全** | Security Context + Seccomp | 防止容器逃逸和提权攻击 | • 非 Root 运行（UID 1000）<br>• 只读根文件系统<br>• Drop All Capabilities<br>• Seccomp 过滤系统调用 |

#### 5.3.2 容器安全加固（Operator 自动应用）

```yaml
# Operator 自动生成的 StatefulSet（安全加固）
apiVersion: apps/v1
kind: StatefulSet
spec:
  template:
    spec:
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        fsGroup: 1000
        seccompProfile:
          type: RuntimeDefault
      
      containers:
      - name: openclaw
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
          capabilities:
            drop:
            - ALL
        
        volumeMounts:
        - name: data
          mountPath: /app/data  # 唯一可写目录
        - name: tmp
          mountPath: /tmp  # 临时目录
      
      volumes:
      - name: tmp
        emptyDir: {}
```

**安全加固对比**：

| 安全措施 | V1.0 手动配置 | V2.0 Operator 自动 |
|---------|-------------|------------------|
| 非 Root 运行 | ❌ 未配置 | ✅ 默认启用（UID 1000） |
| 只读根文件系统 | ❌ 未配置 | ✅ 默认启用 |
| Drop Capabilities | ❌ 未配置 | ✅ 默认 Drop ALL |
| Seccomp 过滤 | ❌ 未配置 | ✅ 默认 RuntimeDefault |
| NetworkPolicy | 🟡 需手动创建 | ✅ 自动生成白名单 |

---

### 5.4 监控与告警

#### 5.4.1 监控架构（Operator 集成）

```
┌─────────────────────────────────────────────────────────┐
│  OpenClaw 实例                                           │
│  ├─ 内置 Prometheus 指标暴露（/metrics）                  │
│  └─ 结构化 JSON 日志（stdout）                           │
└─────────────────────────────────────────────────────────┘
                     ▼
┌─────────────────────────────────────────────────────────┐
│  🆕 ServiceMonitor（Operator 自动创建）                   │
│  ├─ 自动发现实例 Metrics 端点                            │
│  └─ Prometheus 自动采集                                  │
└─────────────────────────────────────────────────────────┐
                     ▼
┌─────────────────────────────────────────────────────────┐
│  Prometheus                                              │
│  ├─ 存储指标（15 天保留期）                               │
│  ├─ 🆕 PrometheusRule（Operator 自动创建告警规则）       │
│  └─ AlertManager（发送告警）                             │
└─────────────────────────────────────────────────────────┘
                     ▼
┌─────────────────────────────────────────────────────────┐
│  🆕 Grafana Dashboard（Operator 自动创建）                │
│  ├─ 实例概览仪表板                                        │
│  ├─ 资源使用趋势                                         │
│  └─ 异常告警面板                                         │
└─────────────────────────────────────────────────────────┘
```

#### 5.4.2 内置告警规则（Operator 自动配置）

| 告警名称 | 触发条件 | 严重程度 | 自动操作 |
|---------|---------|---------|---------|
| **InstanceCrashLoop** | Pod 崩溃 3 次 | Critical | 自动重启 + 通知管理员 |
| **InstanceOOM** | 内存 OOM | Critical | 自动重启 + 建议升级规格 |
| **HighCPUUsage** | CPU > 90% 持续 10 分钟 | Warning | 通知用户 + 建议升级规格 |
| **HighMemoryUsage** | 内存 > 90% 持续 10 分钟 | Warning | 通知用户 + 建议升级规格 |
| **PVCAlmostFull** | 磁盘使用率 > 85% | Warning | 通知用户清理数据 |
| **DeploymentFailed** | 部署失败 | Critical | 自动回滚 + 通知管理员 |

---

## 六、非功能需求

### 6.1 性能要求（v2.0 优化）

| 性能指标 | V1.0 目标 | V2.0 目标 | 优化幅度 |
|---------|----------|----------|---------|
| **实例创建时长** | P50 < 5 分钟，P95 < 10 分钟 | P50 < 2 分钟，P95 < 3 分钟 | 60% ⬇️ |
| **部署成功率** | ≥ 95% | ≥ 99% | 4% ⬆️ |
| **审批响应时间** | < 2 秒 | < 2 秒 | 无变更 |
| **大盘加载时间** | < 3 秒 | < 3 秒 | 无变更 |
| **API 响应时间** | P99 < 500ms | P99 < 500ms | 无变更 |

### 6.2 安全要求（v2.0 增强）

| 安全要求 | V1.0 实现 | V2.0 增强 |
|---------|---------|----------|
| **容器安全** | ❌ 未明确 | ✅ 非Root + 只读文件系统 + Seccomp |
| **网络隔离** | 🟡 手动配置 NetworkPolicy | ✅ Operator 自动生成白名单 |
| **RBAC** | 🟡 手动创建 SA/Role | ✅ Operator 自动配置最小权限 |
| **数据加密** | 🟡 V1.1 计划 | ✅ S3 备份加密 + TLS 证书 |
| **审计日志** | ✅ 记录所有 API 调用 | ✅ 无变更 |

---

## 七、排期建议（v2.0 调整）

### 7.1 迭代计划

#### 阶段 0：Operator 部署（M0，第 1-2 周）

| 里程碑 | 时间 | 关键交付物 | 验收标准 |
|-------|------|-----------|---------|
| **M0** | 第 1-2 周 | • 安装 OpenClaw Operator<br>• 验证 Operator 功能<br>• 编写后端适配代码 | • Operator 正常运行<br>• 可创建测试实例<br>• 后端可调用 K8s API 创建 CR |

#### 阶段 1：MVP 版本（M1 - M3，第 3-12 周）

| 里程碑 | 时间 | 关键交付物 | 验收标准 | 变更说明 |
|-------|------|-----------|---------|---------|
| **M1** | 第 3-6 周 | • 员工申请流程<br>• 管理员审批流程<br>• 基于 Operator 的部署 | • 审批通过后自动创建实例（成功率 > 99%） | 部署成功率从 80% → 99% |
| **M2** | 第 7-9 周 | • 我的实例列表<br>• 全局监控大盘<br>• 健康检查和告警 | • 集成 Operator 的 ServiceMonitor<br>• 显示 Grafana Dashboard | 集成 Operator 监控 |
| **M3** | 第 10-12 周 | • 成本统计<br>• 配额管理<br>• 审计日志 | • 按部门统计成本<br>• 所有操作记录审计日志 | 无变更 |

#### 阶段 2：产品化版本（M4 - M6，3-6 个月）

| 里程碑 | 时间 | 关键交付物 | 变更说明 |
|-------|------|-----------|---------|
| **M4** | 第 4 个月 | • Skills 管理<br>• 🆕 自适应配置（SelfConfig）<br>• 详细成本分析 | 利用 Operator 的 SelfConfig CRD |
| **M5** | 第 5 个月 | • 🆕 自动版本管理（AutoUpdate）<br>• 🆕 灰度发布<br>• 🆕 自动回滚 | 提前实现（原计划 M5） |
| **M6** | 第 6 个月 | • 🆕 自动备份恢复（S3）<br>• 数据加密<br>• 多集群支持 | 提前实现（原计划 M4） |

**关键变化**：
- ✅ M4 备份恢复提前到 M2（Operator 原生支持）
- ✅ M5 版本管理简化为配置 AutoUpdate 策略
- ✅ 新增 M0 阶段（Operator 部署）

---

## 八、风险与依赖（v2.0 更新）

### 8.1 风险点（v2.0 调整）

| 风险类型 | 风险描述 | 影响程度 | V2.0 应对措施 | 责任人 |
|---------|---------|---------|-------------|--------|
| **技术风险** | OpenClaw Operator 不稳定 | 🟡 中 | • 使用稳定版本（v1.0+）<br>• 与 OpenClaw 社区建立联系<br>• 准备降级方案（手动部署） | 技术负责人 |
| **技术风险** | Operator 升级导致实例异常 | 🟡 中 | • 在测试环境充分验证<br>• 灰度升级 Operator<br>• 保留旧版本 Operator | 技术负责人 |
| **技术风险** | TKE API 不稳定 | 🟡 中（降低） | • Operator 内置重试机制<br>• 降低对 TKE API 的直接依赖 | 技术负责人 |
| **业务风险** | 员工滥用配额 | 🟡 中 | • Operator 支持资源配额限制<br>• 监控异常使用 | 产品经理 |

### 8.2 外部依赖（v2.0 新增）

| 依赖项 | 依赖内容 | 关键程度 | 风险 | 应对措施 |
|-------|---------|---------|------|---------|
| **🆕 OpenClaw Operator** | 实例生命周期管理 | 🔴 关键 | Operator Bug 导致实例异常 | • 使用稳定版本<br>• 与社区保持联系<br>• 准备降级方案 |
| **🆕 cert-manager** | 自动 HTTPS 证书 | 🟡 重要 | cert-manager 故障导致证书无法申请 | • 提前申请通配符证书<br>• 降级方案：手动证书 |
| **TKE API** | 创建 Namespace 和 CR | 🔴 关键 | TKE API 变更 | • 适配 TKE 多个版本<br>• Operator 封装 API 变更 |

---

## 九、商业化方案（无变更）

### 9.1 计费模式

与 V1.0 保持一致，详见原 PRD Part 2 第十一章。

**核心优势**：
- 成本结构更优：Operator 减少人力维护成本
- 毛利率提升：自动化运维降低运营成本 5%-10%

---

## 十、附录

### A. 关键术语表（v2.0 新增）

| 术语 | 英文 | 说明 |
|-----|------|------|
| **Operator** | Kubernetes Operator | K8s 扩展，用于自动化应用的部署和运维 |
| **CR (Custom Resource)** | Custom Resource | K8s 自定义资源，如 OpenClawInstance |
| **CRD** | CustomResourceDefinition | CR 的定义 |
| **Reconciliation** | - | Operator 的协调循环，确保实际状态与期望状态一致 |
| **SelfConfig** | - | OpenClaw Operator 的自适应配置能力 |
| **AutoUpdate** | - | Operator 的自动升级功能 |
| **ServiceMonitor** | - | Prometheus Operator 的服务监控资源 |
| **PrometheusRule** | - | Prometheus 告警规则资源 |

### B. 参考文档（v2.0 新增）

1. **OpenClaw Operator GitHub**：https://github.com/openclaw-rocks/k8s-operator
2. **OpenClaw Operator 文档**：https://openclaw-rocks.github.io/k8s-operator/
3. **Kubernetes Operator 最佳实践**：https://kubernetes.io/docs/concepts/extend-kubernetes/operator/
4. **cert-manager 文档**：https://cert-manager.io/docs/

---

## 十一、待确认问题（v2.0 新增）

### 11.1 需要与 OpenClaw 社区确认的问题

1. **Operator 稳定性**
   - OpenClaw Operator 的当前稳定版本？
   - 是否有生产环境使用案例？
   - 是否支持 TKE 环境？

2. **Operator 定制需求**
   - 是否需要 Fork Operator 增加企业特性？
   - 如何与 TKE 的监控和日志服务集成？

3. **社区支持**
   - 如何获得技术支持？
   - 是否可以参与 Operator 开发？

---

## 总结

### V2.0 核心改进

1. **简化部署**：从 9 步减少到 3 步，部署时间减少 60%
2. **提升可靠性**：部署成功率从 80% 提升到 99%+
3. **降低维护成本**：后端代码减少 85%
4. **增强安全性**：容器安全加固开箱即用
5. **自动化运维**：自动备份、升级、回滚、监控

### 下一步行动

1. **与 OpenClaw 社区沟通**：确认 Operator 在 TKE 环境的兼容性
2. **M0 阶段启动**：安装和验证 Operator
3. **修改后端架构**：适配 Operator 的 CR 管理方式
4. **更新前端界面**：展示 Operator 提供的新功能（如 AutoUpdate）

---

**文档结束**

**（本 PRD 为 V2.0 版本，基于 OpenClaw K8s Operator 进行了全面优化）**
