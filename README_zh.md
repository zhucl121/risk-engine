# RiskEngine

[![Go 版本](https://img.shields.io/badge/go-1.24+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-Apache%202.0-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/zhucl121/risk-engine)](https://goreportcard.com/report/github.com/zhucl121/risk-engine)
[![CI](https://github.com/zhucl121/risk-engine/actions/workflows/ci.yml/badge.svg)](https://github.com/zhucl121/risk-engine/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/zhucl121/risk-engine/graph/badge.svg)](https://codecov.io/gh/zhucl121/risk-engine)

**RiskEngine** 是一款基于 Go 语言开发的高性能开源风控决策引擎，专为支付、营销活动、交易等场景的实时欺诈检测设计。在普通硬件上可达到 **P99 < 60ms**、**20,000+ TPS** 的决策能力。

> English documentation: [README.md](README.md)

---

## 功能特性

### 核心决策

| 特性 | 说明 |
|------|------|
| **多策略决策** | 规则引擎 + ML 模型评分 + 名单服务，通过可配置 DAG 流水线协同编排 |
| **热更新** | 规则和模型无需重启即可生效，变更在 30 秒内传播完毕 |
| **自研 RiskDSL** | 基于 ANTLR4-Go 的表达式语言，加载时编译为 Go 闭包；P99 29ns（简单条件）/ 97ns（含特征读取），零堆分配 |
| **并行特征拉取** | 所有特征源并发查询，单源超时降级，绝不阻塞决策主路径 |
| **独立 Feature Store** | 可选的外部特征服务，通过 gRPC 调用；内置 `VelocityGroup`（滑动窗口计数）和 `UserProfileGroup`（Redis JSON Hash）；超时时自动 fail-open |
| **速率计数器** | Redis Lua 原子滑动窗口计数，支持任意时间粒度（1分钟 / 1小时 / 24小时） |
| **名单服务** | Redis 黑名单 / 灰名单 / 白名单，O(1) 查询 |

### 策略编排

| 特性 | 说明 |
|------|------|
| **A/B 测试** | 随机流量分流；实验组结果标记在 `RiskReasons` 中，无需重启 |
| **灰度分流（Canary）** | 基于 SHA-256 hash 的**确定性分流**，同一用户每次都落入同一桶（稳定路由）；支持 userID / deviceID / sessionID / IP / `extra.<key>` 等多维度 hash key；通过独立 salt 隔离不同实验的桶位 |
| **陪跑 / Shadow 模式** | 新策略并行执行，结果不影响主决策；全量写入 `shadow_audit` 日志，离线分析后再决定上线 |
| **冠军-挑战者** | 多个挑战者策略后台并发执行，冠军结果返回给调用方；双方结果写入 `cc_audit`，含 `agreement` 字段，用于统计胜率、翻转率、延迟差 |
| **步骤级条件** | 每个步骤可配置 DSL 表达式 `condition`，条件为 false 时跳过该步骤 |
| **自动重试** | 步骤失败时按 `maxAttempts` + `delayMs` 自动重试 |
| **多聚合策略** | `HIGHEST_RISK`（默认）/ `WEIGHTED`（加权求和）/ `RULE_FIRST`（规则优先，无命中再用模型分数） |
| **降级策略** | 步骤失败策略：`SKIP`（跳过）/ `REJECT`（高风险兜底）/ `FALLBACK`（场景兜底决策） |
| **熔断器** | 基于 `gobreaker` 的按步骤熔断（名单、模型），状态通过 Prometheus Gauge 暴露 |

### 参数与数据

| 特性 | 说明 |
|------|------|
| **Extra 参数规格管理** | 每个场景的 Extra 字段规格存储于 MySQL，支持必需校验（缺失则拒绝请求）和可选默认值填充；后台每 30s 热重载 |
| **Extra → 特征注入** | `DecisionRequest.Extra` 字段自动注入 `feature.Map`（key 为 `extra.<字段名>`），类型由 DB 规格驱动（string / int / float / bool）；步骤级 `ParamMapping` 支持向下游服务映射任意字段 |
| **场景无关设计** | 决策请求无业务耦合字段（金额、订单类型等均放入 `extra`），引擎适用于支付、营销、登录等任意风控场景 |

### 可观测性与运维

| 特性 | 说明 |
|------|------|
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
git clone https://github.com/zhucl121/risk-engine.git
cd riskengine

# 复制并编辑配置文件
cp configs/config.example.yaml configs/config.local.yaml

# 启动依赖服务
docker compose -f deployments/docker/compose.dev.yaml up -d

# 启动服务
go run ./cmd/server -config configs/config.local.yaml
```

### 发起一次风控决策

所有业务字段（包括金额等）均通过 `extra` 传入，引擎不预设业务含义字段：

```bash
curl -X POST http://localhost:8080/api/v1/decision \
  -H "Content-Type: application/json" \
  -d '{
    "scene_code": "PAYMENT_CHECKOUT",
    "user_id":    "u123456",
    "device_id":  "d-abc-def",
    "ip":         "1.2.3.4",
    "extra": {
      "amount":      "9900",
      "merchant_id": "M001",
      "product_type": "GOODS"
    }
  }'
```

返回示例：

```json
{
  "request_id": "01HZ...",
  "decision":   "PASS",
  "risk_score": 120,
  "risk_level": "LOW",
  "hit_rules":  [],
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
请求 → API 层（Gin HTTP / gRPC）
              ↓
   限流 / 指标 / 链路追踪 中间件
              ↓
        DecisionEngine
              ↓
        Orchestrator（DAG 编排）
         ├─ 路由层：Canary > A/B Test > 主 Pipeline
         │
         ├─ 主 Pipeline（顺序 + 并行步骤）
         │    ├── LIST   →  名单服务（Redis + 熔断）
         │    ├── RULE   →  规则引擎（RiskDSL）
         │    └── MODEL  →  模型引擎（ONNX + 熔断）
         │
         ├─ Shadow Pipeline（后台并行，不影响主决策）
         │    └── → shadow_audit 日志
         │
         └─ Champion-Challenger（后台并发，挑战者结果不返回）
              └── → cc_audit 日志

  FeatureService（并行特征拉取）
   ├── VelocityFetcher（滑动窗口计数）→ Redis
   ├── UserProfile → Redis JSON
   └── FeatureStoreFetcher → gRPC Feature Store

AuditWriter（异步 channel → 结构化日志 / Kafka）
  ├── audit（主决策记录）
  ├── shadow_audit（陪跑记录）
  └── cc_audit（冠军-挑战者记录）
```

详细架构设计与决策依据见 [docs/architecture.md](docs/architecture.md)。

---

## 核心概念

### PolicySet（策略集）

每个场景（`scene_code`）对应一个 `PolicySet`，通过 YAML 定义完整的决策流水线：

```yaml
- sceneCode: payment
  version: "1.0.0"
  fallback: MANUAL_REVIEW     # 流水线异常时的兜底决策
  strategy: HIGHEST_RISK      # 聚合策略: HIGHEST_RISK | WEIGHTED | RULE_FIRST
  extraSchema:
    amount: int               # 静态类型声明（DB 规格优先）
    merchant_id: string

  pipeline:
    - name: blacklist_check
      kind: LIST
      timeoutMs: 20
      onFailure: SKIP
      listQueryFields:        # 自定义查询维度（默认 user/device/ip）
        - extra.merchant_id
        - request.ip

    - name: payment_rules
      kind: RULE
      ruleGroup: payment
      timeoutMs: 50
      condition: "extra.amount > 0"   # 步骤条件，false 时跳过
      retry:
        maxAttempts: 2
        delayMs: 5

    - name: risk_model
      kind: MODEL
      models: [payment_fraud_v2]
      timeoutMs: 80
      weight: 0.7             # WEIGHTED 策略时生效
      params:                 # 步骤级参数映射
        merchant: extra.merchant_id
        channel:  "WEB"

  # A/B 测试（随机分流，不稳定）
  abTest:
    enabled: true
    experimentId: payment-model-v3
    splitPct: 0.05
    experimentPipeline:
      - name: model_v3
        kind: MODEL
        models: [payment_fraud_v3]
        timeoutMs: 80

  # 灰度分流（hash 稳定，同一用户始终落同一桶）
  canary:
    enabled: true
    canaryVersion: "v2.1.0"
    trafficPct: 10            # 10% 用户进入灰度
    hashKey: userID           # 路由 key: userID | deviceID | sessionID | ip | extra.<key>
    salt: "payment_canary_v2" # 实验盐，不同实验须用不同 salt
    canaryPipeline:
      - name: new_rules
        kind: RULE
        ruleGroup: payment_v2
      - name: model_v3
        kind: MODEL
        models: [payment_fraud_v3]

  # 陪跑 / Shadow 模式（不影响主决策，结果写入 shadow_audit）
  shadowPolicies:
    - sceneCode: payment_new_policy
      version: "draft-1"

  # 冠军-挑战者（双方并行，冠军结果返回，均写 cc_audit）
  championChallenger:
    enabled: true
    experimentID: "fraud_model_v3_eval"
    challengers:
      - challengerID: "model_v3_candidate"
        trafficPct: 20         # 20% 用户参与挑战
        hashKey: userID
        salt: "cc_fraud_v3"
        pipeline:
          - name: model_v3
            kind: MODEL
            models: [payment_fraud_v3]
      - challengerID: "rule_baseline"
        trafficPct: 100
        hashKey: userID
        salt: "cc_rule_baseline"
        pipeline:
          - name: rules_only
            kind: RULE
            ruleGroup: payment
```

---

### 策略路由优先级

```
Canary（hash 稳定，用户维度定向）
  ↓ 未命中
A/B Test（随机，请求维度）
  ↓ 未命中
主 Pipeline
```

三者互斥，同一请求只走一个分支。Shadow 模式和冠军-挑战者作为独立的**旁路并发**，不参与路由竞争，所有请求都会触发（在各自的流量比例内）。

---

### 各模式对比

| 模式 | 主决策影响 | 流量路由 | 用途 |
|------|-----------|---------|------|
| **A/B 测试** | ✅ 实验组走不同 pipeline | 随机（每次请求独立） | 对称实验，两组均为生产策略 |
| **Canary 灰度** | ✅ 灰度用户走新 pipeline | Hash 稳定（同用户同桶） | 渐进发布，逐步扩大新策略覆盖 |
| **Shadow 陪跑** | ❌ 不影响主决策 | 全量（或特定场景） | 上线前验证，离线对比分析 |
| **冠军-挑战者** | ❌ 挑战者不影响主决策 | Hash 稳定（按 trafficPct） | 策略评估，统计显著性对比 |

---

### RiskDSL（风控表达式语言）

规则条件使用自研的 RiskDSL 编写。DSL 在规则加载时由 ANTLR4 编译为 Go 闭包，运行时零分配。

#### 支持的操作符

```
# 比较运算
amount > 10000
risk_score != 0

# 逻辑运算
amount > 5000 && velocity("pay", user_id, "1h") > 10
amount > 1000 || inList("blacklist_ip", ip)

# in / not in（数组成员判断）
extra.product_type in ["DIGITAL", "VIRTUAL"]
extra.channel not in ["OFFLINE", "STORE"]

# 三元表达式
extra.vip_level == "gold" ? 200 : 500

# 取反
!isEmpty(user_id)
```

#### 内置函数

**字符串函数**

| 函数 | 说明 |
|------|------|
| `contains(s, sub)` | 包含子串 |
| `startsWith(s, prefix)` | 前缀匹配 |
| `endsWith(s, suffix)` | 后缀匹配 |
| `match(s, pattern)` | 正则匹配（缓存编译） |
| `lower(s)` / `upper(s)` | 大小写转换 |
| `trim(s)` | 去除首尾空格 |
| `strlen(s)` | 字符串长度 |
| `isEmpty(s)` | 是否为空 |

**数学函数**

| 函数 | 说明 |
|------|------|
| `abs(n)` | 绝对值 |
| `ceil(n)` / `floor(n)` / `round(n)` | 取整 |
| `sqrt(n)` | 开方 |
| `min(a, b)` / `max(a, b)` | 最值 |
| `clamp(n, lo, hi)` | 区间限制 |

**时间函数**

| 函数 | 说明 |
|------|------|
| `now()` | 当前 Unix 时间戳（秒） |
| `nowMs()` | 当前毫秒时间戳 |
| `daysSince(t)` | 距今天数 |
| `hoursSince(t)` | 距今小时数 |
| `secondsSince(t)` | 距今秒数 |
| `toUnix(t)` | 时间字符串转 Unix 时间戳 |
| `hour(t)` | 提取小时（0–23） |
| `weekday(t)` | 星期几（0 = 周日，6 = 周六） |

**类型转换与条件函数**

| 函数 | 说明 |
|------|------|
| `toInt(v)` / `toFloat(v)` / `toString(v)` / `toBool(v)` | 类型转换 |
| `isNull(v)` | 是否为空值 |
| `coalesce(a, b, ...)` | 返回第一个非空值 |
| `ifThen(cond, a, b)` | 条件选择（等效三元运算） |

**风控专属函数**

| 函数 | 说明 |
|------|------|
| `inList(kind, value)` | 查询名单服务（黑/灰/白名单） |
| `velocity(prefix, id, window)` | 读取滑动窗口计数 |
| `modelScore(name)` | 获取 ML 模型评分（0~1） |
| `geoIP(ip)` | 返回 IP 归属国家码 |
| `within(lat, lon, clat, clon, km)` | 地理围栏判断 |

**规则示例**

```
# 高频支付检测（1小时内超过10次）
velocity("pay_count", user_id, "1h") > 10 && extra.amount > 5000

# 黑名单 + 金额组合
inList("blacklist_user", user_id) || (extra.amount > 50000 && extra.vip_level not in ["gold", "platinum"])

# 夜间高额交易（23点~6点）
extra.amount > 10000 && (hour(now()) >= 23 || hour(now()) <= 6)

# 新设备高额转账
daysSince(extra.device_register_time) < 7 && extra.amount > 20000

# 三元判断风险阈值
velocity("pay_fail", user_id, "24h") > (extra.vip_level == "gold" ? 20 : 5)
```

DSL 完整语法文档见 [docs/dsl-guide.md](docs/dsl-guide.md)。

---

## Feature Store（独立特征服务）

特征拉取支持两种模式，可按需选择或混合使用：

### 模式对比

```
模式一（默认）：进程内 Fetcher，直连 Redis
RiskEngine 进程
  └── feature.Service
        ├── VelocityFetcher  ──→ Redis
        └── ...

模式二：外部 Feature Store，通过 gRPC 调用
RiskEngine 进程                     Feature Store 进程
  └── feature.Service                 └── FeatureStoreService
        └── FeatureStoreFetcher ─gRPC─→    ├── VelocityGroup   → Redis
                                           └── UserProfileGroup → Redis JSON
```

### 启动独立 Feature Store

```bash
go run ./cmd/featurestore -config configs/config.yaml
```

### 配置决策引擎接入

```yaml
feature_store:
  enabled: true
  addr: "localhost:9100"
  request_timeout: "20ms"
  groups:
    - name: "user_profile"
      timeout: "15ms"
    - name: "velocity"
      timeout: "10ms"
```

### 自定义 FeatureGroup

```go
type MyGroup struct{ db *sql.DB }

func (g *MyGroup) Name() string { return "my_group" }

func (g *MyGroup) Fetch(ctx context.Context, entity *riskv1.EntityContext) (
    map[string]*riskv1.FeatureValue, []string, error,
) {
    return map[string]*riskv1.FeatureValue{
        "credit_score": {Value: &riskv1.FeatureValue_IntVal{IntVal: 750}},
    }, nil, nil
}
```

注册：

```go
store.DefaultRegistry.Register(&MyGroup{db: db})
```

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
| `session_id` | string | 会话 ID |
| `ip` | string | 客户端 IP |
| `extra` | map[string]string | 场景自定义扩展字段（包括金额等业务字段） |

> **注意**：交易金额等业务字段通过 `extra` 传递，例如 `"extra": {"amount": "9900"}`，引擎会按场景配置的 `extraSchema` 进行类型转换。

#### 健康检查

```
GET /api/v1/livez    # 存活探针，进程存活即返回 200
GET /api/v1/readyz   # 就绪探针，依赖不可用返回 503
GET /metrics         # Prometheus 指标
```

#### 规则管理（Admin）

```
GET    /admin/v1/rules                          # 查询规则列表
POST   /admin/v1/rules                          # 创建规则
PUT    /admin/v1/rules/:id                      # 更新规则
DELETE /admin/v1/rules/:id                      # 删除规则
POST   /admin/v1/rules/:id/enable               # 启用规则
POST   /admin/v1/rules/:id/disable              # 禁用规则
POST   /admin/v1/rules/validate                 # 校验 DSL 表达式
GET    /admin/v1/scenes/:scene/extra-params     # 查询场景 Extra 参数规格
POST   /admin/v1/scenes/:scene/extra-params     # 新增参数规格
PUT    /admin/v1/scenes/:scene/extra-params/:key  # 更新参数规格
DELETE /admin/v1/scenes/:scene/extra-params/:key  # 删除参数规格
```

### gRPC API

服务定义见 [api/grpc/proto/decision.proto](api/grpc/proto/decision.proto)，默认监听 `:9090`。

```protobuf
service DecisionService {
  rpc Evaluate(DecisionRequest) returns (DecisionResponse);
  rpc BatchEvaluate(BatchDecisionRequest) returns (BatchDecisionResponse);
  rpc Health(HealthRequest) returns (HealthResponse);
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

1. 实现 `feature.Fetcher` 接口：

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
# 必需字段（缺失则请求被 400 拒绝）
POST /admin/v1/scenes/payment/extra-params
{
  "param_key":   "merchant_id",
  "param_type":  "string",
  "required":    true,
  "description": "商户 ID，必填"
}

# 可选字段（缺失时自动填充 default_val）
POST /admin/v1/scenes/payment/extra-params
{
  "param_key":   "product_type",
  "param_type":  "string",
  "required":    false,
  "default_val": "GOODS",
  "description": "商品类型，默认 GOODS"
}
```

**运行时行为：**

| 场景 | 行为 |
|------|------|
| 必需字段缺失 | 返回 `ErrMissingRequiredExtra`（HTTP 400），包含字段名和场景码 |
| 可选字段缺失且有默认值 | 自动填充 `default_val`，再做类型转换注入 `feature.Map` |
| 可选字段缺失且无默认值 | 不注入，规则引用 `extra.<key>` 得到零值 |
| 字段已存在 | 直接做类型转换，跳过默认值填充 |

DB 规格与 YAML 静态 `ExtraSchema` 合并，**DB 优先**。规格缓存在内存中，后台每 30s 增量热重载。

### 配置灰度分流（Canary）

灰度分流使用 SHA-256 哈希取模，保证**同一用户每次落入同一桶**：

```yaml
canary:
  enabled: true
  canaryVersion: "v2.1.0"
  trafficPct: 10        # 10% 用户进入灰度（整数，精确控制）
  hashKey: userID       # userID | deviceID | sessionID | ip | extra.<key>
  salt: "exp_payment_canary_v2"   # 每个实验用不同 salt，避免不同实验桶位相关
  canaryPipeline:
    - name: new_rule_engine
      kind: RULE
      ruleGroup: payment_v2
```

**扩量流程**：从 5% → 10% → 25% → 50% → 100%，逐步在管理后台调整 `trafficPct`，无需重启服务。

### 配置冠军-挑战者实验

```yaml
championChallenger:
  enabled: true
  experimentID: "fraud_model_eval_q1"
  challengers:
    - challengerID: "xgb_v3"
      trafficPct: 30
      hashKey: userID
      salt: "cc_xgb_v3_q1"
      pipeline:
        - name: model_v3
          kind: MODEL
          models: [fraud_xgb_v3]
```

查询 cc_audit 日志进行统计对比：

```bash
# 按实验 ID 聚合，计算挑战者与冠军的判决一致率
jq 'select(.experiment_id == "fraud_model_eval_q1") | .agreement' cc_audit.log | \
  awk '{a[$0]++} END {print "agree:", a["true"], "disagree:", a["false"]}'
```

### 配置 Shadow 陪跑模式

```yaml
shadowPolicies:
  - sceneCode: payment_new_strategy
    version: "draft-2025Q1"
```

陪跑结果写入 `shadow_audit` 日志，字段包含 `shadow_decision`、`prod_decision`，可直接对比分析：

```bash
# 查看陪跑与主策略的判决差异
jq '{req: .request_id, prod: .production_decision, shadow: .decision}' shadow_audit.log
```

### 添加 ML 模型

1. 将模型导出为 ONNX 格式
2. 将 `.onnx` 文件放入 `configs/models/`
3. 在 `configs/models/registry.yaml` 中注册元数据

引擎支持热加载，更新模型文件后无需重启。

### 扩展 RiskDSL 函数

在 `pkg/dsl/builtins/` 中注册自定义函数：

```go
registry.RegisterFunc("myRiskScore", func(ctx context.Context, rt *dsl.Runtime, args []dsl.Value) (dsl.Value, error) {
    // 访问 request 和 features
    userID := rt.Request.UserID
    _ = userID
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
| 简单条件（`extra.amount > 1000`） | P99 29ns | 0 allocs |
| 含特征读取（`velocity(...)>10`） | P99 97ns | 0 allocs |
| `in` 数组判断（5 元素） | P99 43ns | 0 allocs |

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
├── internal/
│   ├── engine/            # 顶层 DecisionEngine，请求生命周期
│   ├── rule/              # 规则存储、评估器、热更新
│   ├── feature/           # 并行特征拉取
│   │   └── fetchers/      # 具体拉取器（VelocityFetcher 等）
│   ├── featurestore/      # gRPC Feature Store 客户端 + 服务端
│   │   └── store/         # FeatureGroup 注册表（VelocityGroup / UserProfileGroup）
│   ├── model/             # 模型注册表、ONNX 评分接口
│   ├── list/              # Redis 名单服务（黑名单 / 灰名单 / 白名单）
│   ├── orchestrator/      # DAG 执行器、策略路由（A/B + Canary）、
│   │                      # 陪跑 Shadow、冠军-挑战者、Extra 注入、熔断
│   ├── scene/             # 场景级 Extra 参数规格（DB 持久化，热重载）
│   ├── audit/             # 异步审计写入（主决策 / shadow_audit / cc_audit）
│   ├── metrics/           # Prometheus 指标定义
│   ├── middleware/        # Gin 中间件（RequestID / 指标 / 限流 / 日志 / 追踪）
│   ├── health/            # 存活 / 就绪检查器
│   ├── resilience/        # 熔断器（gobreaker 封装）
│   └── config/            # 配置加载器
├── pkg/
│   ├── dsl/               # 自研 RiskDSL（ANTLR4-Go / AST / 代码生成）
│   │   ├── grammar/       # RiskDSL.g4 语法文件
│   │   ├── parser/        # ANTLR4 自动生成的 Parser（请勿手动修改）
│   │   ├── ast/           # AST 节点类型
│   │   └── builtins/      # 内置函数（字符串 / 数学 / 时间 / 类型转换 / 风控）
│   ├── sliding/           # Redis Lua 滑动窗口速率计数器
│   ├── bloom/             # 进程内 Bloom 过滤器
│   └── pool/              # 对象池工具
├── api/
│   ├── grpc/
│   │   ├── proto/         # decision.proto
│   │   ├── v1/            # protoc 生成的 Go 代码
│   │   └── server/        # DecisionServer 实现
│   └── http/
│       ├── v1/            # 决策 API / 健康检查
│       └── admin/v1/      # 规则管理 + Extra 参数规格管理
├── configs/
│   ├── config.example.yaml
│   ├── migrations/        # 数据库迁移 SQL
│   └── policies/          # PolicySet YAML（启动时自动加载）
├── deployments/           # Docker / Kubernetes 部署清单
├── docs/                  # 架构文档、DSL 指南、设计方案
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
