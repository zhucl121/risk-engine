# RiskEngine

[![Go 版本](https://img.shields.io/badge/go-1.24+-00ADD8.svg?logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-Apache%202.0-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/zhucl121/risk-engine)](https://goreportcard.com/report/github.com/zhucl121/risk-engine)
[![CI](https://github.com/zhucl121/risk-engine/actions/workflows/ci.yml/badge.svg)](https://github.com/zhucl121/risk-engine/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/zhucl121/risk-engine/graph/badge.svg)](https://codecov.io/gh/zhucl121/risk-engine)

基于 Go 语言开发的高性能开源**风控决策引擎**，专为支付、营销、交易等场景的实时欺诈检测设计。

**P99 < 60ms · 20,000+ TPS · 零分配 DSL · 热更新**

> English documentation: [README.md](README.md)

---

## 概述

RiskEngine 将规则引擎、ML 模型和名单服务编排成可配置的 DAG 流水线。每个风控场景完全声明式 — 策略通过 YAML 定义，运行时热更新无需重启。

```
请求 ──► Orchestrator（DAG）
           ├── 路由层   Canary  ──► hash 稳定分流（同用户同桶）
           │           A/B     ──► 随机分流
           │
           ├── 流水线   LIST   ──► Redis 黑/灰/白名单
           │           RULE   ──► RiskDSL（ANTLR4，编译为 Go 闭包）
           │           MODEL  ──► ONNX 评分（熔断器保护）
           │
           ├── Shadow  ──► 并行陪跑，写入 shadow_audit
           └── 挑战者  ──► 并行评估，写入 cc_audit

FeatureService  ──► 并行特征拉取（Redis 速率计数、gRPC Feature Store …）
AuditWriter     ──► 异步 channel → 结构化日志 / Kafka
```

---

## 功能特性

| 分类 | 功能 | 说明 |
|------|------|------|
| **DSL** | 自研 RiskDSL | ANTLR4-Go 表达式语言，加载时编译为 Go 闭包。P99 29ns（简单条件）/ 97ns（含特征读取），**零堆分配** |
| **DSL** | 丰富的操作符 | 比较、逻辑运算、`in` / `not in` 集合判断、三元表达式 `?:`、取反 |
| **DSL** | 内置函数库 | 字符串 · 数学 · 时间 · 类型转换 · 风控专属（`inList` · `velocity` · `modelScore` · `geoIP` · `within`） |
| **特征** | 并行拉取 | 所有特征源并发查询，单源超时 fail-open，不阻塞决策主路径 |
| **特征** | 速率计数器 | Redis Lua 原子滑动窗口计数，支持任意时间粒度（1分钟 / 1小时 / 24小时） |
| **特征** | 独立 Feature Store | 可选 gRPC 服务（`VelocityGroup`、`UserProfileGroup`），插件化 `FeatureGroup` 接口 |
| **名单** | 名单服务 | Redis 黑名单 / 灰名单 / 白名单，O(1) 查询 |
| **编排** | 灰度分流（Canary） | SHA-256 hash 取模**稳定路由** — 同一用户每次落同一桶；支持 `userID` / `deviceID` / `sessionID` / `IP` / `extra.<key>` 作为 hash key，实验间独立 salt |
| **编排** | A/B 测试 | 随机流量分流，实验组标记写入 `RiskReasons`，无需重启 |
| **编排** | 陪跑 / Shadow | 新策略并行执行，不影响主决策；结果写入 `shadow_audit` 供离线分析 |
| **编排** | 冠军-挑战者 | 挑战者后台并发执行，冠军决策返回给调用方；双方结果写入 `cc_audit`，含 `agreement` 一致率字段 |
| **编排** | 聚合策略 | `HIGHEST_RISK`（默认）· `WEIGHTED`（加权求和）· `RULE_FIRST`（规则优先，无命中再用模型分） |
| **编排** | 步骤级控制 | 每步可配 DSL `condition`（false 时跳过）、`retry`（最大重试 + 延迟）、`onFailure`（SKIP / REJECT / FALLBACK） |
| **数据** | Extra 参数规格 | 每场景字段规格存 MySQL — 必需 / 可选（含默认值）；DB 驱动类型转换注入 `feature.Map`；每 30s 热重载 |
| **数据** | 参数映射 | 步骤级 `ParamMapping`，将请求 / 特征字段映射到下游服务入参 |
| **数据** | 场景无关设计 | 请求结构体无业务字段；金额、商户 ID 等均通过 `extra` 传递，适用任意风控场景 |
| **韧性** | 熔断器 | `gobreaker` 按步骤熔断（名单、模型）；状态以 Prometheus Gauge 暴露 |
| **韧性** | 限流 | 两级令牌桶：全局 5,000 RPS + 单 IP 100 RPS；超限返回 HTTP 429 |
| **热更新** | 规则与策略 | 内存原子指针替换 — 零停机，< 30s 传播 |
| **可观测性** | 指标 | Prometheus：决策延迟、规则命中、特征错误、活跃请求 |
| **可观测性** | 链路追踪 | OpenTelemetry trace header 提取，传播至所有下游调用 |
| **可观测性** | 审计日志 | 三条异步 channel：`audit`（主决策）· `shadow_audit` · `cc_audit` → 结构化日志 / Kafka |
| **运维** | 健康探针 | `/api/v1/livez` · `/api/v1/readyz`（含 Redis 依赖检查），Kubernetes 原生支持 |
| **运维** | 双协议 | HTTP/JSON（Gin）+ gRPC（`Evaluate` / `BatchEvaluate` / `Health`） |
| **运维** | 云原生 | 可配置 drain 超时的优雅关闭 |

---

## 快速开始

### 环境依赖

- Go 1.24+
- Redis 7+
- MySQL 8+（用于 Extra 参数规格，可选）

### 本地运行

```bash
git clone https://github.com/zhucl121/risk-engine.git
cd risk-engine

cp configs/config.example.yaml configs/config.local.yaml
docker compose -f deployments/docker/compose.dev.yaml up -d
go run ./cmd/server -config configs/config.local.yaml
```

### 发起一次风控决策

所有场景业务字段（包括金额）均通过 `extra` 传递：

```bash
curl -s -X POST http://localhost:8080/api/v1/decision \
  -H "Content-Type: application/json" \
  -d '{
    "scene_code": "PAYMENT_CHECKOUT",
    "user_id":    "u123456",
    "device_id":  "d-abc-def",
    "ip":         "203.0.113.1",
    "extra": {
      "amount":       "9900",
      "merchant_id":  "M001",
      "product_type": "GOODS"
    }
  }'
```

```json
{
  "request_id":   "01HZ...",
  "decision":     "PASS",
  "risk_score":   120,
  "risk_level":   "LOW",
  "hit_rules":    [],
  "model_scores": { "payment_fraud_xgb": 0.08 },
  "risk_reasons": [],
  "cost_ms":      23
}
```

| 决策值 | 含义 |
|--------|------|
| `PASS` | 通过，正常放行 |
| `REJECT` | 拒绝，命中高风险规则或黑名单 |
| `MANUAL_REVIEW` | 人工审核，灰名单或低置信度命中 |

---

## 策略配置

策略通过 YAML 定义并热更新，单个 `PolicySet` 覆盖某场景的完整决策生命周期。

```yaml
- sceneCode: payment
  version: "1.0.0"
  fallback: MANUAL_REVIEW       # 流水线无法完成时的兜底决策
  strategy: HIGHEST_RISK        # HIGHEST_RISK | WEIGHTED | RULE_FIRST

  extraSchema:                  # 静态类型提示（DB 规格优先）
    amount:      int
    merchant_id: string

  pipeline:
    - name: blacklist_check
      kind: LIST
      timeoutMs: 20
      onFailure: SKIP
      listQueryFields:          # 自定义查询维度（默认：user/device/ip）
        - extra.merchant_id
        - request.ip

    - name: payment_rules
      kind: RULE
      ruleGroup: payment
      timeoutMs: 50
      condition: "extra.amount > 0"   # 为 false 时跳过该步骤
      retry:
        maxAttempts: 2
        delayMs: 5

    - name: fraud_model
      kind: MODEL
      models: [payment_fraud_v2]
      timeoutMs: 80
      weight: 0.7               # WEIGHTED 策略时生效
      params:                   # 下游服务参数映射
        merchant: extra.merchant_id
        channel:  "WEB"

  # ── 流量路由（同一请求最多命中一个） ──────────────────────────────────────────

  canary:                       # hash 稳定，渐进式灰度
    enabled: true
    canaryVersion: "v2.1.0"
    trafficPct: 10              # 10% 用户进入灰度，无需重启即可调整
    hashKey: userID             # userID | deviceID | sessionID | ip | extra.<key>
    salt: "payment_canary_v2"   # 每个实验用独立 salt
    canaryPipeline:
      - { name: model_v3, kind: MODEL, models: [payment_fraud_v3] }

  abTest:                       # 随机，每次请求独立
    enabled: false
    experimentId: payment-model-v3
    splitPct: 0.05
    experimentPipeline:
      - { name: model_v3, kind: MODEL, models: [payment_fraud_v3] }

  # ── 后台并行评估（不影响主决策） ──────────────────────────────────────────────

  shadowPolicies:               # 陪跑 → shadow_audit
    - sceneCode: payment_new_policy
      version: "draft-1"

  championChallenger:           # 冠军-挑战者 → cc_audit
    enabled: true
    experimentID: "fraud_model_v3_eval"
    challengers:
      - challengerID: "model_v3_candidate"
        trafficPct: 20
        hashKey: userID
        salt: "cc_fraud_v3"
        pipeline:
          - { name: model_v3, kind: MODEL, models: [payment_fraud_v3] }
```

### 路由优先级

```
Canary（hash 稳定，用户维度）
  ↓ 未命中
A/B Test（随机，请求维度）
  ↓ 未命中
主 Pipeline
```

| 模式 | 影响主决策 | 路由方式 | 适用场景 |
|------|-----------|---------|---------|
| A/B 测试 | ✅ 实验组走不同 pipeline | 随机 | 对称实验，两组均为生产策略 |
| Canary 灰度 | ✅ 灰度用户走新 pipeline | Hash 稳定 | 渐进发布，逐步扩量 |
| Shadow 陪跑 | ❌ | 全量（或指定场景） | 上线前验证，离线对比分析 |
| 冠军-挑战者 | ❌ 挑战者结果不返回 | Hash 稳定 | 策略评估，统计显著性对比 |

---

## RiskDSL

规则条件使用 **RiskDSL** 编写 — 由 ANTLR4 生成的解析器在规则加载时编译为 Go 闭包，运行时**零堆分配**。

### 语法

```
# 比较与逻辑运算
extra.amount > 10000 && velocity("pay", user_id, "1h") > 5

# 集合判断
extra.product_type in ["DIGITAL", "VIRTUAL"]
extra.channel not in ["OFFLINE"]

# 三元表达式
extra.vip_level == "gold" ? 200 : 500

# 取反
!isEmpty(user_id)
```

### 内置函数

**字符串** — `contains` · `startsWith` · `endsWith` · `match` · `lower` · `upper` · `trim` · `strlen` · `isEmpty`

**数学** — `abs` · `ceil` · `floor` · `round` · `sqrt` · `min` · `max` · `clamp`

**时间** — `now` · `nowMs` · `daysSince` · `hoursSince` · `secondsSince` · `toUnix` · `hour` · `weekday`

**类型转换** — `toInt` · `toFloat` · `toString` · `toBool` · `isNull` · `coalesce` · `ifThen`

**风控专属** — `inList(kind, value)` · `velocity(prefix, id, window)` · `modelScore(name)` · `geoIP(ip)` · `within(lat, lon, clat, clon, km)`

### 规则示例

```
# 高频支付检测
velocity("pay_count", user_id, "1h") > 10 && extra.amount > 5000

# 黑名单 + 高额非 VIP 交易
inList("blacklist_user", user_id) || (extra.amount > 50000 && extra.vip_level not in ["gold", "platinum"])

# 夜间大额转账（23点 ~ 6点）
extra.amount > 10000 && (hour(now()) >= 23 || hour(now()) <= 6)

# 新设备大额操作
daysSince(extra.device_register_time) < 7 && extra.amount > 20000
```

自定义函数注册：

```go
registry.RegisterFunc("myScore",
    func(ctx context.Context, rt *dsl.Runtime, args []dsl.Value) (dsl.Value, error) {
        return dsl.IntValue(42), nil
    })
```

完整语法文档：[docs/dsl-guide.md](docs/dsl-guide.md)

---

## Feature Store

Feature Store 是可选的独立 gRPC 服务，引擎始终将所有注册数据源的结果合并到同一个 `feature.Map`。

```
进程内（默认）                       独立服务（可选）
────────────────────                 ─────────────────────────────
feature.Service                      cmd/featurestore
  ├── VelocityFetcher → Redis          └── FeatureStoreService
  └── 自定义 Fetcher                         ├── VelocityGroup   → Redis
                                             └── UserProfileGroup → Redis JSON
```

配置接入：

```yaml
feature_store:
  enabled: true
  addr: "localhost:9100"
  request_timeout: "20ms"
  groups:
    - { name: velocity,     timeout: 10ms }
    - { name: user_profile, timeout: 15ms }
```

自定义 FeatureGroup：

```go
type MyGroup struct{}

func (g *MyGroup) Name() string { return "my_group" }
func (g *MyGroup) Fetch(ctx context.Context, entity *riskv1.EntityContext) (
    map[string]*riskv1.FeatureValue, []string, error) {
    return map[string]*riskv1.FeatureValue{
        "credit_score": {Value: &riskv1.FeatureValue_IntVal{IntVal: 750}},
    }, nil, nil
}
```

---

## API 参考

### HTTP

```
POST /api/v1/decision         发起风控决策
GET  /api/v1/livez            存活探针
GET  /api/v1/readyz           就绪探针（含 Redis 依赖检查）
GET  /metrics                 Prometheus 指标
```

**决策请求字段**

| 字段 | 类型 | 说明 |
|------|------|------|
| `scene_code` | string | 场景码 *（必填）* |
| `user_id` | string | 用户 ID |
| `device_id` | string | 设备 ID |
| `session_id` | string | 会话 ID |
| `ip` | string | 客户端 IP |
| `extra` | `map[string]string` | 所有场景业务字段（金额、商户 ID 等） |

**管理 API**

```
GET    /admin/v1/rules                             查询规则列表
POST   /admin/v1/rules                             创建规则
PUT    /admin/v1/rules/:id                         更新规则
DELETE /admin/v1/rules/:id                         删除规则
POST   /admin/v1/rules/:id/enable                  启用规则
POST   /admin/v1/rules/:id/disable                 禁用规则
POST   /admin/v1/rules/validate                    校验 DSL 表达式

GET    /admin/v1/scenes/:scene/extra-params        查询场景 Extra 参数规格
POST   /admin/v1/scenes/:scene/extra-params        新增参数规格
PUT    /admin/v1/scenes/:scene/extra-params/:key   更新参数规格
DELETE /admin/v1/scenes/:scene/extra-params/:key   删除参数规格
```

### gRPC

默认端口 `:9090`，Proto 定义见 [api/grpc/proto/decision.proto](api/grpc/proto/decision.proto)。

```protobuf
service DecisionService {
  rpc Evaluate(DecisionRequest)           returns (DecisionResponse);
  rpc BatchEvaluate(BatchDecisionRequest) returns (BatchDecisionResponse);
  rpc Health(HealthRequest)               returns (HealthResponse);
}
```

---

## 性能指标

测试环境：8 核 / 32 GB 虚拟机 · Go 1.24 · Redis 7（本地）

| 场景 | P50 | P99 | TPS |
|------|-----|-----|-----|
| 仅名单查询 | 0.8 ms | 1.5 ms | 80,000 |
| 仅规则（100 条） | 3 ms | 8 ms | 45,000 |
| 规则 + 模型评分 | 15 ms | 35 ms | 25,000 |
| 完整流水线（名单 + 规则 + 模型） | 22 ms | 55 ms | 20,000 |

**RiskDSL** 基准（单核，无 I/O）：

| 表达式 | P99 | 分配 |
|--------|-----|------|
| 简单条件（`extra.amount > 1000`） | 29 ns | 0 |
| 含特征读取（`velocity(...) > 10`） | 97 ns | 0 |
| 集合判断（`x in [...]`，5 元素） | 43 ns | 0 |

---

## 项目结构

```
risk-engine/
├── cmd/
│   ├── server/          主 HTTP + gRPC 服务入口
│   └── featurestore/    独立 Feature Store gRPC 服务
├── internal/
│   ├── engine/          顶层 DecisionEngine
│   ├── orchestrator/    DAG 执行器 · 路由 · Shadow · 冠军-挑战者
│   ├── rule/            规则存储、评估器、热更新
│   ├── feature/         并行特征拉取
│   │   └── fetchers/    VelocityFetcher …
│   ├── featurestore/    gRPC 客户端 + 服务端 + FeatureGroup 注册表
│   ├── model/           模型注册表、ONNX 评分接口
│   ├── list/            Redis 名单服务
│   ├── scene/           Extra 参数规格（DB 持久化，热重载）
│   ├── audit/           异步审计写入（audit / shadow_audit / cc_audit）
│   ├── resilience/      熔断器
│   ├── middleware/      限流 · 指标 · 追踪 · 日志
│   └── health/          存活 / 就绪检查器
├── pkg/
│   ├── dsl/             RiskDSL（语法文件 · AST · 代码生成 · 内置函数）
│   ├── sliding/         Redis Lua 滑动窗口计数器
│   └── pool/            对象池工具
├── api/
│   ├── grpc/            Proto 定义 + 生成代码 + Server 实现
│   └── http/            Gin Handler（决策 · 管理 · 健康检查）
├── configs/
│   ├── config.example.yaml
│   ├── migrations/      SQL 迁移脚本
│   └── policies/        PolicySet YAML 文件
├── deployments/         Docker / Kubernetes 部署清单
└── docs/                架构文档 · DSL 指南 · 设计方案
```

---

## 参与贡献

欢迎提交 PR，请在提交前阅读 [CONTRIBUTING.md](CONTRIBUTING.md)。

```bash
make setup   # 安装工具链（golangci-lint · mockery · protoc）
make test    # 运行单元测试
make lint    # 运行代码检查
make bench   # 运行基准测试
```

---

## 许可证

Apache License 2.0，详见 [LICENSE](LICENSE)。

---

## 致谢

- [antlr/antlr4](https://github.com/antlr/antlr4) · [antlr4-go/antlr](https://github.com/antlr/antlr4/tree/master/runtime/Go/antlr) — DSL 解析器运行时
- [sony/gobreaker](https://github.com/sony/gobreaker) — 熔断器
- [prometheus/client_golang](https://github.com/prometheus/client_golang) — Prometheus 指标
- [redis/go-redis](https://github.com/redis/go-redis) — Redis 客户端
- [gin-gonic/gin](https://github.com/gin-gonic/gin) — HTTP 框架
- 设计参考了哔哩哔哩、字节跳动、美团风控团队的业界实践
