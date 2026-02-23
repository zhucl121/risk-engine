# RiskEngine

[![Go 版本](https://img.shields.io/badge/go-1.24+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-Apache%202.0-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/yourorg/riskengine)](https://goreportcard.com/report/github.com/yourorg/riskengine)
[![CI](https://github.com/yourorg/riskengine/actions/workflows/ci.yml/badge.svg)](https://github.com/yourorg/riskengine/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/yourorg/riskengine/graph/badge.svg)](https://codecov.io/gh/yourorg/riskengine)

**RiskEngine** 是一款基于 Go 语言开发的高性能开源风控决策引擎，专为支付、营销活动、交易等场景的实时欺诈检测设计。在普通硬件上可达到 **P99 < 60ms**、**20,000+ TPS** 的决策能力。

> English documentation: [README.md](README.md)

---

## 功能特性

| 特性 | 说明 |
|------|------|
| **多策略决策** | 规则引擎 + ML 模型评分 + 名单服务，通过可配置 DAG 流水线协同编排 |
| **热更新** | 规则和模型无需重启即可生效，变更在 30 秒内传播完毕 |
| **自研 RiskDSL** | 基于 ANTLR4-Go 的表达式语言，加载时编译为 Go 闭包；P99 29ns（简单条件）/ 97ns（含特征读取），零堆分配 |
| **并行特征拉取** | 所有特征源并发查询，单源超时降级，绝不阻塞决策主路径 |
| **独立 Feature Store** | 可选的外部特征服务，通过 gRPC 调用；内置 `VelocityGroup`（滑动窗口计数）和 `UserProfileGroup`（Redis JSON Hash）；超时时自动 fail-open |
| **速率计数器** | Redis Lua 原子滑动窗口计数，支持任意时间粒度（1分钟 / 1小时 / 24小时） |
| **名单服务** | Redis 黑名单 / 灰名单 / 白名单，O(1) 查询 |
| **Extra 参数规格管理** | 每个场景的 Extra 字段规格存储于 MySQL，支持必需校验（缺失则拒绝请求）和可选默认值填充；后台每 30s 热重载 |
| **Extra → 特征注入** | `DecisionRequest.Extra` 字段自动注入 `feature.Map`（key 为 `extra.<字段名>`），类型由 DB 规格驱动（string/int/float/bool）；步骤级 `ParamMapping` 支持向下游服务映射任意字段 |
| **A/B 测试** | 按 `PolicySet.ABTest` 配置流量分流，实验组结果标记在 `RiskReasons` 中，无需重启 |
| **熔断器** | 基于 `gobreaker` 的按步骤熔断（名单、模型），状态通过 Prometheus Gauge 暴露 |
| **限流** | 两级令牌桶：全局 5000 RPS + 单 IP 100 RPS，超限返回 HTTP 429 |
| **可观测性** | Prometheus 指标（决策延迟、规则命中、特征错误、活跃请求）、OpenTelemetry 链路追踪、结构化 Zap 日志、异步审计写入 |
| **健康探针** | `/api/v1/livez`（存活探针）和 `/api/v1/readyz`（就绪探针，含 Redis 依赖检查），原生支持 Kubernetes |
| **双协议** | HTTP/JSON（Gin）+ gRPC（`DecisionService`：Evaluate / BatchEvaluate / Health） |
| **云原生** | 可配置 drain 超时的优雅关闭，开箱即用的 Kubernetes 部署 |

---

## 快速开始

### 环境依赖

- Go 1.24+
- Redis 7+
- Kafka 3+（用于审计日志，可选 — 开发模式使用结构化日志作为降级方案）

### 本地运行

```bash
git clone https://github.com/yourorg/riskengine.git
cd riskengine

# 复制并编辑配置文件
cp configs/config.example.yaml configs/config.local.yaml

# 启动依赖服务
docker compose -f deployments/docker/compose.dev.yaml up -d

# 启动服务
go run ./cmd/server -config configs/config.local.yaml
```

### 发起一次风控决策

```bash
curl -X POST http://localhost:8080/api/v1/decision \
  -H "Content-Type: application/json" \
  -d '{
    "scene_code": "PAYMENT_CHECKOUT",
    "user_id": "u123456",
    "device_id": "d-abc-def",
    "ip": "1.2.3.4",
    "amount": 9900
  }'
```

返回示例：

```json
{
  "request_id": "01HZ...",
  "decision": "PASS",
  "risk_score": 120,
  "risk_level": "LOW",
  "hit_rules": [],
  "model_scores": {"payment_fraud_xgb": 0.08},
  "risk_reasons": [],
  "cost_ms": 23
}
```

决策值说明：

| 值 | 含义 |
|----|------|
| `PASS` | 通过，正常放行 |
| `REJECT` | 拒绝，命中高风险规则或黑名单 |
| `MANUAL_REVIEW` | 人工审核，处于灰名单或规则命中但置信度不足 |

---

## 系统架构

```
请求 → API 层（Gin HTTP + gRPC）
                  ↓
     限流 / 指标 / 链路追踪 中间件
                  ↓
          DecisionEngine（决策引擎）
                  ↓
          Orchestrator（DAG 编排 + A/B 路由）
         ↙         ↓              ↘
   名单服务      规则引擎         模型引擎
  (Redis+熔断)  (RiskDSL)      (ONNX+熔断)
         ↘         ↓              ↙
          FeatureService（并行特征拉取）
          ↙        ↓             ↘
   VelocityFetcher  用户画像    设备信息
   (滑动窗口计数)
                  ↓
           Redis（主存储）
                  ↓
     AuditWriter（异步 channel → 结构化日志 / Kafka）
```

**关键说明：**
- **熔断器**：连续失败 5 次后断开，30 秒后进入半开探测
- **A/B 路由**：`PolicySet.ABTest.SplitPct` 比例的流量走 `ExperimentPipeline`，实验组结果附加标记

详细架构设计与决策依据见 [docs/architecture.md](docs/architecture.md)。

---

## 核心概念

### PolicySet（策略集）

每个场景（`scene_code`）对应一个 `PolicySet`，定义完整的决策流水线：

```yaml
- sceneCode: payment          # 场景码
  version: "1.0.0"
  fallback: MANUAL_REVIEW     # 流水线异常时的兜底决策
  pipeline:
    - name: blacklist_check   # 步骤名称
      kind: LIST              # 步骤类型: LIST / RULE / MODEL / AGGREGATE
      timeoutMs: 20           # 单步超时
      onFailure: SKIP         # 失败策略: SKIP / REJECT / FALLBACK
    - name: payment_rules
      kind: RULE
      ruleGroup: payment
      timeoutMs: 50
    - name: risk_model
      kind: MODEL
      models: [payment_fraud_v2]
      timeoutMs: 80
      strategy: HIGHEST_RISK  # 聚合策略: HIGHEST_RISK / MAJORITY_VOTE
  abTest:
    enabled: false
    experimentId: payment-model-v3
    splitPct: 0.05            # 5% 流量走实验组
    experimentPipeline: [...]
```

### RiskDSL（风控表达式语言）

规则条件使用自研的 RiskDSL 编写，支持：

```
# 基础比较
amount > 10000

# 逻辑组合
amount > 5000 AND velocity("pay", user_id, "1h") > 10

# 内置函数
inList("blacklist_ip", ip) OR modelScore("fraud_v2") > 0.85

# 地理位置
geoIP(ip) == "CN" AND within(lat, lon, 39.9, 116.4, 50)
```

内置函数说明：

| 函数 | 说明 |
|------|------|
| `inList(kind, value)` | 查询名单服务 |
| `velocity(prefix, id, window)` | 读取滑动窗口计数 |
| `modelScore(name)` | 获取 ML 模型评分（0~1） |
| `geoIP(ip)` | 返回 IP 归属国家码 |
| `within(lat, lon, clat, clon, km)` | 地理围栏判断 |

---

## 配置说明

所有配置均为 YAML 格式，完整配置项及注释见 [configs/config.example.yaml](configs/config.example.yaml)。

核心配置项：

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `server.addr` | HTTP 监听地址 | `:8080` |
| `server.grpc_addr` | gRPC 监听地址 | `:9090` |
| `redis` | Redis 连接池、集群 / Sentinel 模式 | — |
| `kafka` | Broker 地址、审计 Topic（可选） | — |
| `engine.policy_dir` | PolicySet YAML 文件目录 | `configs/policies/` |
| `rules` | 规则分组、热更新间隔 | — |
| `models` | ONNX 模型路径 | — |
| `feature.redis_timeout` | 单个特征拉取 Redis 超时 | `10ms` |

---

## API 接口

### HTTP API

#### 发起决策

```
POST /api/v1/decision
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `scene_code` | string | 场景码（必填） |
| `user_id` | string | 用户 ID |
| `device_id` | string | 设备 ID |
| `ip` | string | 客户端 IP |
| `amount` | int64 | 交易金额（最小货币单位，如分） |
| `extra` | map | 场景自定义扩展字段 |

#### 健康检查

```
GET /api/v1/livez    # 存活探针，进程存活即返回 200
GET /api/v1/readyz   # 就绪探针，依赖不可用返回 503
GET /metrics         # Prometheus 指标
```

#### 规则管理（Admin）

```
GET    /admin/v1/rules              # 查询规则列表
POST   /admin/v1/rules              # 创建规则
PUT    /admin/v1/rules/:id          # 更新规则
DELETE /admin/v1/rules/:id          # 删除规则
POST   /admin/v1/rules/:id/enable   # 启用规则
POST   /admin/v1/rules/:id/disable  # 禁用规则
POST   /admin/v1/rules/validate     # 校验 DSL 表达式
```

### gRPC API

服务定义见 [api/grpc/proto/decision.proto](api/grpc/proto/decision.proto)。

```protobuf
service DecisionService {
  rpc Evaluate(DecisionRequest) returns (DecisionResponse);
  rpc BatchEvaluate(BatchDecisionRequest) returns (BatchDecisionResponse);
  rpc Health(HealthRequest) returns (HealthResponse);
}
```

gRPC 默认监听 `:9090`。

---

## Feature Store（独立特征服务）

特征拉取支持两种模式，可按需选择或混合使用：

### 模式对比

```
模式一（默认）：进程内 Fetcher，直连 Redis
RiskEngine 进程
  └── feature.Service
        ├── VelocityFetcher  ──→ Redis（pkg/sliding）
        └── PromoVelocityFetcher ──→ Redis

模式二：外部 Feature Store，通过 gRPC 调用
RiskEngine 进程                     Feature Store 进程（cmd/featurestore）
  └── feature.Service                     └── FeatureStoreService
        └── FeatureStoreFetcher ──gRPC──→       ├── VelocityGroup   → Redis
                                                └── UserProfileGroup → Redis JSON
```

### 启动独立 Feature Store

```bash
# Feature Store 默认监听 :9100
go run ./cmd/featurestore -config configs/config.yaml
```

### 决策引擎接入 Feature Store

在 `configs/config.yaml` 中开启：

```yaml
feature_store:
  enabled: true
  addr: "localhost:9100"    # sidecar 用 localhost，k8s 用 ClusterIP DNS
  request_timeout: "20ms"
  groups:
    - name: "user_profile"  # 对应 Feature Store 中注册的 FeatureGroup.Name()
      timeout: "15ms"
    - name: "velocity"
      timeout: "10ms"
```

启动后，`feature.Service` 会同时并发调用进程内 Fetcher 和 gRPC Fetcher，结果合并入同一个 `feature.Map`。

### 自定义 FeatureGroup

在 Feature Store 服务中实现 `store.FeatureGroup` 接口：

```go
type MyGroup struct{ db *sql.DB }

func (g *MyGroup) Name() string { return "my_group" }

func (g *MyGroup) Fetch(ctx context.Context, entity *riskv1.EntityContext) (
    map[string]*riskv1.FeatureValue, []string, error,
) {
    // 从数据库 / 外部 API 查询特征
    return map[string]*riskv1.FeatureValue{
        "my_feature": {Value: &riskv1.FeatureValue_IntVal{IntVal: 100}},
    }, nil, nil
}
```

注册到 store：

```go
registry.Register(&MyGroup{db: db})
```

Feature Store 重启后决策引擎会自动重连（gRPC retry policy 已内置）。

### 健康检查

Feature Store 客户端自动接入 `/api/v1/readyz`：

```json
{
  "ready": false,
  "checks": [
    {"name": "redis", "healthy": true},
    {"name": "feature-store", "healthy": false, "message": "grpc health: connection refused"}
  ]
}
```

---

## 扩展开发

### 添加规则

1. 在 `internal/rule/rules/<name>.go` 实现 `rule.Rule` 接口
2. 在 `internal/rule/registry.go` 中注册
3. 在 `configs/rules/<group>.yaml` 添加规则配置
4. 编写单元测试和基准测试

详细步骤参考 [docs/adding-rules.md](docs/adding-rules.md)。

### 添加特征拉取器

1. 在 `internal/feature/fetchers/<name>.go` 实现 `feature.Fetcher` 接口：

```go
type Fetcher interface {
    Name()    string
    Timeout() time.Duration
    Fetch(ctx context.Context, req *engine.DecisionRequest) (feature.Map, error)
}
```

2. 在 `cmd/server/main.go` 中注册：

```go
featureSvc.Register(fetchers.NewMyFetcher(deps))
```

参考实现：`internal/feature/fetchers/velocity_fetcher.go`

### 配置 Extra 参数规格（数据库）

先执行迁移脚本：

```bash
mysql -u root riskengine < configs/migrations/002_create_scene_extra_params.sql
```

通过管理 API 维护规格：

```bash
# 查询场景的 Extra 参数规格
GET /admin/v1/scenes/payment/extra-params

# 新增必需字段
POST /admin/v1/scenes/payment/extra-params
{
  "param_key":   "merchant_id",
  "param_type":  "string",
  "required":    true,
  "description": "商户 ID，必填"
}

# 新增可选字段（带默认值）
POST /admin/v1/scenes/payment/extra-params
{
  "param_key":   "product_type",
  "param_type":  "string",
  "required":    false,
  "default_val": "GOODS",
  "description": "商品类型，默认 GOODS"
}

# 更新字段
PUT /admin/v1/scenes/payment/extra-params/product_type
{ "default_val": "DIGITAL" }

# 软删除字段
DELETE /admin/v1/scenes/payment/extra-params/product_type
```

**运行时行为：**

| 场景 | 行为 |
|------|------|
| 必需字段缺失 | 返回 `ErrMissingRequiredExtra`（HTTP 400），包含字段名和场景码 |
| 可选字段缺失且有默认值 | 自动填充 `default_val`，再做类型转换注入 `feature.Map` |
| 可选字段缺失且无默认值 | 不注入，规则引用 `extra.<key>` 得到零值 |
| 字段已存在 | 直接做类型转换，跳过默认值填充 |

DB 规格与 YAML 静态 `ExtraSchema` 合并，**DB 优先**。规格缓存在内存中，后台每 30s 增量热重载。

### 添加 ML 模型

1. 将模型导出为 ONNX 格式
2. 将 `.onnx` 文件放入 `configs/models/`
3. 在 `configs/models/registry.yaml` 中注册元数据

引擎支持热加载，更新模型文件后无需重启。

### 扩展 RiskDSL 函数

在 `pkg/dsl/builtins/` 中注册自定义函数：

```go
registry.Register("myFunc", func(ctx context.Context, rt dsl.Runtime, args []dsl.Value) (dsl.Value, error) {
    // 实现自定义逻辑
    return dsl.IntValue(42), nil
})
```

---

## 性能指标

测试环境：8 核 / 32 GB 虚拟机（Go 1.24，Redis 7 本地部署）

| 场景 | P50 | P99 | TPS |
|------|-----|-----|-----|
| 仅名单查询 | 0.8ms | 1.5ms | 80,000 |
| 仅规则（100 条） | 3ms | 8ms | 45,000 |
| 规则 + 模型评分 | 15ms | 35ms | 25,000 |
| 完整流水线（名单 + 规则 + 模型） | 22ms | 55ms | 20,000 |

DSL 表达式基准（单核，无 Redis 依赖）：

| 场景 | 耗时 | 堆分配 |
|------|------|--------|
| 简单条件（`amount > 1000`） | P99 29ns | 0 allocs |
| 含特征读取（`velocity(...)>10`） | P99 97ns | 0 allocs |

本地运行基准测试：

```bash
make bench
```

---

## 项目结构

```
riskengine/
├── cmd/
│   ├── server/            # 主 HTTP + gRPC 服务入口
│   └── featurestore/      # 独立 Feature Store gRPC 服务入口
├── internal/              # 私有业务代码
│   ├── engine/            # 顶层 DecisionEngine，请求生命周期
│   ├── rule/              # 规则存储、评估器、热更新
│   ├── feature/           # 并行特征拉取
│   │   └── fetchers/      # 具体拉取器（VelocityFetcher 等）
│   ├── featurestore/      # gRPC Feature Store 客户端、Fetcher 适配器、Server 实现
│   │   └── store/         # FeatureGroup 注册表（VelocityGroup / UserProfileGroup）
│   ├── model/             # 模型注册表、ONNX 评分接口
│   ├── list/              # Redis 名单服务（黑名单 / 灰名单）
│   ├── orchestrator/      # DAG 执行器、A/B 路由、策略注册表、Extra 注入
│   ├── scene/             # 场景级 Extra 参数规格（DB 持久化，带热重载）
│   ├── audit/             # 异步 channel 审计写入（→ 日志 / Kafka）
│   ├── metrics/           # Prometheus 指标定义
│   ├── middleware/        # Gin 中间件（RequestID / Metrics / RateLimit / Logger / Tracing）
│   ├── health/            # 存活 / 就绪检查器
│   ├── resilience/        # 熔断器（gobreaker 封装）
│   └── config/            # 配置加载器
├── pkg/                   # 可复用公共包
│   ├── dsl/               # 自研 RiskDSL（ANTLR4-Go / AST / 代码生成）
│   │   ├── grammar/       # RiskDSL.g4 语法文件
│   │   ├── parser/        # ANTLR4 自动生成的 Parser（请勿手动修改）
│   │   ├── ast/           # AST 节点类型
│   │   └── builtins/      # 内置风控函数（inList / velocity / ...）
│   ├── sliding/           # Redis Lua 滑动窗口速率计数器
│   ├── bloom/             # 进程内 Bloom 过滤器
│   └── pool/              # 对象池工具
├── api/
│   ├── grpc/              # Proto 定义 + 生成代码 + Server 实现
│   │   ├── proto/         # decision.proto
│   │   ├── v1/            # protoc 生成的 Go 代码
│   │   └── server/        # DecisionServer 实现
│   └── http/              # Gin HTTP Handler
│       ├── v1/            # 决策 API / 健康检查 / livez / readyz
│       └── admin/v1/      # 规则管理 CRUD API + Extra 参数规格管理 API
├── configs/
│   ├── config.example.yaml
│   ├── migrations/        # 数据库迁移脚本（SQL）
│   └── policies/          # PolicySet YAML 文件（启动时自动加载）
├── deployments/           # Docker 和 Kubernetes 部署清单
├── docs/                  # 架构文档、设计方案
└── openspec/              # 变更提案与设计规范
```

---

## 开发指南

### 环境搭建

```bash
make setup     # 安装工具链（golangci-lint / mockery / protoc）
make test      # 运行单元测试
make lint      # 运行代码检查
make bench     # 运行基准测试
```

### 开发规范

- 每实现一个完整功能模块，进行一次 `git commit`
- 新增非平凡功能前，须在 `openspec/changes/<name>/proposal.md` 中创建提案
- 触碰 `engine/` `rule/` `feature/` `model/` `list/` 包下接口时，须有 `design.md`（含 Architect REVIEW 章节）
- 实现代码须对应 `tasks.md` 中一个未完成的 `[ ]` 任务

### 如何贡献

欢迎提交 PR！请在提交前阅读 [CONTRIBUTING.md](CONTRIBUTING.md)。

---

## 许可证

Apache License 2.0，详见 [LICENSE](LICENSE)。

---

## 致谢

- [antlr/antlr4](https://github.com/antlr/antlr4) + [antlr4-go/antlr](https://github.com/antlr/antlr4/tree/master/runtime/Go/antlr) — RiskDSL Parser 运行时
- [bits-and-blooms/bloom](https://github.com/bits-and-blooms/bloom) — Bloom 过滤器
- [sony/gobreaker](https://github.com/sony/gobreaker) — 熔断器
- [prometheus/client_golang](https://github.com/prometheus/client_golang) — Prometheus 指标
- [redis/go-redis](https://github.com/redis/go-redis) — Redis 客户端
- 设计参考了哔哩哔哩、字节跳动、美团风控团队的业界实践
