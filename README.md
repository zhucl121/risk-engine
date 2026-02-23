# RiskEngine

[![Go Version](https://img.shields.io/badge/go-1.24+-00ADD8.svg?logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-Apache%202.0-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/zhucl121/risk-engine)](https://goreportcard.com/report/github.com/zhucl121/risk-engine)
[![CI](https://github.com/zhucl121/risk-engine/actions/workflows/ci.yml/badge.svg)](https://github.com/zhucl121/risk-engine/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/zhucl121/risk-engine/graph/badge.svg)](https://codecov.io/gh/zhucl121/risk-engine)

A high-performance, open-source **risk decision engine** written in Go — designed for real-time fraud detection in payment, marketing, and transaction scenarios.

**P99 < 60 ms · 20,000+ TPS · Zero-allocation DSL · Hot-reload**

> 中文文档：[README_zh.md](README_zh.md)

---

## Overview

RiskEngine orchestrates rules, ML models, and list lookups into a configurable DAG pipeline. Each decision scene is fully declarative — policies are defined in YAML and reloaded at runtime without restart.

```
Request ──► Orchestrator (DAG)
              ├── Routing   Canary  ──► hash-stable user bucket
              │             A/B     ──► random split
              │
              ├── Pipeline  LIST   ──► Redis blacklist / graylist
              │             RULE   ──► RiskDSL (ANTLR4, compiled closures)
              │             MODEL  ──► ONNX scoring (circuit-breaker protected)
              │
              ├── Shadow    ──► parallel dry-run, writes shadow_audit
              └── Challenger ──► parallel evaluation, writes cc_audit

FeatureService  ──► parallel fetch (Redis velocity, gRPC Feature Store, ...)
AuditWriter     ──► async channel → structured log / Kafka
```

---

## Features

| Category | Feature | Description |
|----------|---------|-------------|
| **DSL** | Self-hosted RiskDSL | ANTLR4-Go expression language — compiled to Go closures at load time. P99 29 ns (simple condition), 97 ns (with feature read), **0 allocations** per evaluation |
| **DSL** | Rich operator set | Comparison, boolean logic, `in` / `not in` set membership, ternary `?:`, negation |
| **DSL** | Built-in functions | String · Math · Time · Type conversion · Risk-specific (`inList`, `velocity`, `modelScore`, `geoIP`, `within`) |
| **Features** | Parallel fetching | All sources queried concurrently; per-source timeout with fail-open, never blocks the decision path |
| **Features** | Velocity counters | Redis Lua atomic sliding-window counters at any granularity (1 min / 1 h / 24 h) |
| **Features** | Standalone Feature Store | Optional gRPC service (`VelocityGroup`, `UserProfileGroup`); pluggable `FeatureGroup` interface |
| **Lists** | List service | Redis blacklist / graylist / whitelist, O(1) lookup |
| **Orchestration** | Canary routing | SHA-256 hash–based **stable routing** — same user always in the same bucket; configurable `hashKey` (userID / deviceID / sessionID / IP / extra field) and per-experiment `salt` |
| **Orchestration** | A/B testing | Random traffic split; experiment group labelled in `RiskReasons`; no restart |
| **Orchestration** | Shadow / dry-run | New policy runs in parallel without affecting the production decision; results written to `shadow_audit` for offline analysis |
| **Orchestration** | Champion-Challenger | Challenger pipelines run concurrently; champion decision returned; both outcomes written to `cc_audit` with `agreement` flag |
| **Orchestration** | Aggregation strategies | `HIGHEST_RISK` · `WEIGHTED` (weighted sum) · `RULE_FIRST` (rules take priority over models) |
| **Orchestration** | Step-level control | Per-step DSL `condition` (skip when false), `retry` (maxAttempts + delayMs), `onFailure` (SKIP / REJECT / FALLBACK) |
| **Data** | Extra parameter specs | Per-scene field rules in MySQL — required / optional with default; DB-driven type coercion into `feature.Map`; hot-reloaded every 30 s |
| **Data** | Parameter mapping | Per-step `ParamMapping` remaps request / feature fields to downstream service parameters |
| **Data** | Domain-agnostic | No business-domain fields on the request struct; all scene-specific data (amount, merchant ID, …) live in `Extra` |
| **Resilience** | Circuit breaker | `gobreaker`-backed per-step breakers; state exposed as Prometheus gauge |
| **Resilience** | Rate limiting | Two-tier token bucket: global 5,000 RPS + per-IP 100 RPS; HTTP 429 on exhaustion |
| **Hot-reload** | Rules & policies | In-memory atomic pointer swap — zero downtime, < 30 s propagation |
| **Observability** | Metrics | Prometheus: decision latency, rule hits, feature errors, active requests |
| **Observability** | Tracing | OpenTelemetry trace header extraction; propagated to all downstream calls |
| **Observability** | Audit trail | Three async channels: `audit` (main) · `shadow_audit` · `cc_audit` → structured log / Kafka |
| **Operations** | Health probes | `/api/v1/livez` · `/api/v1/readyz` (Redis dependency check) — Kubernetes-native |
| **Operations** | Dual protocol | HTTP/JSON (Gin) + gRPC (`Evaluate` / `BatchEvaluate` / `Health`) |
| **Operations** | Cloud-native | Graceful shutdown with configurable drain timeout |

---

## Quick Start

### Prerequisites

- Go 1.24+
- Redis 7+
- MySQL 8+ (for Extra param specs; optional)

### Run locally

```bash
git clone https://github.com/zhucl121/risk-engine.git
cd risk-engine

cp configs/config.example.yaml configs/config.local.yaml
docker compose -f deployments/docker/compose.dev.yaml up -d
go run ./cmd/server -config configs/config.local.yaml
```

### Make a decision

All scene-specific fields (including amount) are passed through `extra`:

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

| Decision | Meaning |
|----------|---------|
| `PASS` | Approved — allow the action |
| `REJECT` | Denied — high-risk rule or blacklist match |
| `MANUAL_REVIEW` | Escalated — graylist match or low-confidence hit |

---

## Policy Configuration

Policies are defined in YAML and hot-reloaded. A single `PolicySet` covers the full decision lifecycle for one scene.

```yaml
- sceneCode: payment
  version: "1.0.0"
  fallback: MANUAL_REVIEW       # returned when the pipeline cannot complete
  strategy: HIGHEST_RISK        # HIGHEST_RISK | WEIGHTED | RULE_FIRST

  extraSchema:                  # static type hints (DB spec takes precedence)
    amount:      int
    merchant_id: string

  pipeline:
    - name: blacklist_check
      kind: LIST
      timeoutMs: 20
      onFailure: SKIP
      listQueryFields:          # custom lookup dimensions (default: user/device/ip)
        - extra.merchant_id
        - request.ip

    - name: payment_rules
      kind: RULE
      ruleGroup: payment
      timeoutMs: 50
      condition: "extra.amount > 0"   # step skipped when false
      retry:
        maxAttempts: 2
        delayMs: 5

    - name: fraud_model
      kind: MODEL
      models: [payment_fraud_v2]
      timeoutMs: 80
      weight: 0.7               # used by WEIGHTED strategy
      params:                   # parameter mapping for downstream calls
        merchant: extra.merchant_id
        channel:  "WEB"

  # ── Traffic routing (at most one is active per request) ─────────────────────

  canary:                       # hash-stable, gradual rollout
    enabled: true
    canaryVersion: "v2.1.0"
    trafficPct: 10              # 10 % of users; increment without restart
    hashKey: userID             # userID | deviceID | sessionID | ip | extra.<key>
    salt: "payment_canary_v2"   # unique per experiment
    canaryPipeline:
      - { name: model_v3, kind: MODEL, models: [payment_fraud_v3] }

  abTest:                       # random per request
    enabled: false
    experimentId: payment-model-v3
    splitPct: 0.05
    experimentPipeline:
      - { name: model_v3, kind: MODEL, models: [payment_fraud_v3] }

  # ── Background evaluation (never affect main decision) ──────────────────────

  shadowPolicies:               # dry-run → shadow_audit
    - sceneCode: payment_new_policy
      version: "draft-1"

  championChallenger:           # evaluation → cc_audit
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

### Routing priority

```
Canary  (hash-stable, user-level)
  ↓ not matched
A/B Test (random, request-level)
  ↓ not matched
Main Pipeline
```

| Mode | Affects production decision | Routing | Use case |
|------|-----------------------------|---------|----------|
| A/B Test | ✅ experiment group runs different pipeline | Random | Symmetric experiments |
| Canary | ✅ canary users run new pipeline | Hash-stable | Gradual rollout |
| Shadow | ❌ | All requests | Pre-release validation |
| Champion-Challenger | ❌ challenger result discarded | Hash-stable | Statistical comparison |

---

## RiskDSL

Rule conditions are written in **RiskDSL** — a domain-specific language compiled to Go closures by an ANTLR4-generated parser. Zero allocations at evaluation time.

### Syntax

```
# Comparison & logic
extra.amount > 10000 && velocity("pay", user_id, "1h") > 5

# Set membership
extra.product_type in ["DIGITAL", "VIRTUAL"]
extra.channel not in ["OFFLINE"]

# Ternary
extra.vip_level == "gold" ? 200 : 500

# Negation
!isEmpty(user_id)
```

### Built-in functions

**String** — `contains` · `startsWith` · `endsWith` · `match` · `lower` · `upper` · `trim` · `strlen` · `isEmpty`

**Math** — `abs` · `ceil` · `floor` · `round` · `sqrt` · `min` · `max` · `clamp`

**Time** — `now` · `nowMs` · `daysSince` · `hoursSince` · `secondsSince` · `toUnix` · `hour` · `weekday`

**Convert** — `toInt` · `toFloat` · `toString` · `toBool` · `isNull` · `coalesce` · `ifThen`

**Risk** — `inList(kind, value)` · `velocity(prefix, id, window)` · `modelScore(name)` · `geoIP(ip)` · `within(lat, lon, clat, clon, km)`

### Example rules

```
# High-frequency payment
velocity("pay_count", user_id, "1h") > 10 && extra.amount > 5000

# Blacklist or large amount from non-VIP
inList("blacklist_user", user_id) || (extra.amount > 50000 && extra.vip_level not in ["gold", "platinum"])

# Late-night large transfer
extra.amount > 10000 && (hour(now()) >= 23 || hour(now()) <= 6)

# New device + large amount
daysSince(extra.device_register_time) < 7 && extra.amount > 20000
```

Custom functions can be registered in Go:

```go
registry.RegisterFunc("myScore",
    func(ctx context.Context, rt *dsl.Runtime, args []dsl.Value) (dsl.Value, error) {
        return dsl.IntValue(42), nil
    })
```

Full reference: [docs/dsl-guide.md](docs/dsl-guide.md)

---

## Feature Store

The Feature Store is an optional standalone gRPC service for feature fetching. The engine always merges results from all registered sources into a single `feature.Map`.

```
In-process (default)                  Standalone (optional)
─────────────────────                 ────────────────────────────
feature.Service                       cmd/featurestore
  ├── VelocityFetcher → Redis           └── FeatureStoreService
  └── custom fetchers                         ├── VelocityGroup   → Redis
                                              └── UserProfileGroup → Redis JSON
```

Enable via config:

```yaml
feature_store:
  enabled: true
  addr: "localhost:9100"
  request_timeout: "20ms"
  groups:
    - { name: velocity,      timeout: 10ms }
    - { name: user_profile,  timeout: 15ms }
```

Implement a custom group by satisfying the `store.FeatureGroup` interface:

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

## API Reference

### HTTP

```
POST /api/v1/decision         Make a risk decision
GET  /api/v1/livez            Liveness probe
GET  /api/v1/readyz           Readiness probe (Redis dependency check)
GET  /metrics                 Prometheus metrics endpoint
```

**Decision request fields**

| Field | Type | Description |
|-------|------|-------------|
| `scene_code` | string | Scene identifier *(required)* |
| `user_id` | string | User identifier |
| `device_id` | string | Device identifier |
| `session_id` | string | Session identifier |
| `ip` | string | Client IP |
| `extra` | `map[string]string` | All scene-specific data — amount, merchant ID, etc. |

**Admin API**

```
GET    /admin/v1/rules                             List rules
POST   /admin/v1/rules                             Create rule
PUT    /admin/v1/rules/:id                         Update rule
DELETE /admin/v1/rules/:id                         Delete rule
POST   /admin/v1/rules/:id/enable                  Enable rule
POST   /admin/v1/rules/:id/disable                 Disable rule
POST   /admin/v1/rules/validate                    Validate DSL expression

GET    /admin/v1/scenes/:scene/extra-params        List Extra param specs
POST   /admin/v1/scenes/:scene/extra-params        Create param spec
PUT    /admin/v1/scenes/:scene/extra-params/:key   Update param spec
DELETE /admin/v1/scenes/:scene/extra-params/:key   Delete param spec
```

### gRPC

Default port `:9090`. See [api/grpc/proto/decision.proto](api/grpc/proto/decision.proto).

```protobuf
service DecisionService {
  rpc Evaluate(DecisionRequest)           returns (DecisionResponse);
  rpc BatchEvaluate(BatchDecisionRequest) returns (BatchDecisionResponse);
  rpc Health(HealthRequest)               returns (HealthResponse);
}
```

---

## Performance

Environment: 8-core / 32 GB VM · Go 1.24 · Redis 7 (local)

| Scenario | P50 | P99 | TPS |
|----------|-----|-----|-----|
| List check only | 0.8 ms | 1.5 ms | 80,000 |
| Rules only (100 rules) | 3 ms | 8 ms | 45,000 |
| Rules + model scoring | 15 ms | 35 ms | 25,000 |
| Full pipeline (list + rules + model) | 22 ms | 55 ms | 20,000 |

**RiskDSL** benchmarks (single core, no I/O):

| Expression | P99 | Allocs |
|------------|-----|--------|
| Simple condition (`extra.amount > 1000`) | 29 ns | 0 |
| Feature read (`velocity(...) > 10`) | 97 ns | 0 |
| Set membership (`x in [...]`, 5 elements) | 43 ns | 0 |

---

## Project Structure

```
risk-engine/
├── cmd/
│   ├── server/          Main HTTP + gRPC server
│   └── featurestore/    Standalone Feature Store gRPC server
├── internal/
│   ├── engine/          Top-level DecisionEngine
│   ├── orchestrator/    DAG executor · routing · shadow · champion-challenger
│   ├── rule/            Rule storage, evaluator, hot-reload
│   ├── feature/         Parallel feature fetching
│   │   └── fetchers/    VelocityFetcher, ...
│   ├── featurestore/    gRPC client + server + FeatureGroup registry
│   ├── model/           Model registry, ONNX scorer
│   ├── list/            Redis list service
│   ├── scene/           Extra param specs (DB-backed, hot-reloaded)
│   ├── audit/           Async audit writer (audit / shadow_audit / cc_audit)
│   ├── resilience/      Circuit breaker
│   ├── middleware/      Rate limit · metrics · tracing · logging
│   └── health/          Liveness / readiness checkers
├── pkg/
│   ├── dsl/             RiskDSL (grammar · AST · codegen · builtins)
│   ├── sliding/         Redis Lua sliding-window counter
│   └── pool/            Object pool utilities
├── api/
│   ├── grpc/            Proto definitions + generated code + server
│   └── http/            Gin handlers (decision · admin · health)
├── configs/
│   ├── config.example.yaml
│   ├── migrations/      SQL migration scripts
│   └── policies/        PolicySet YAML files
├── deployments/         Docker + Kubernetes manifests
└── docs/                Architecture · DSL guide · design documents
```

---

## Contributing

Contributions are welcome. Please read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a pull request.

```bash
make setup   # install tools (golangci-lint, mockery, protoc)
make test    # run unit tests
make lint    # run linter
make bench   # run benchmarks
```

---

## License

Apache License 2.0 — see [LICENSE](LICENSE).

---

## Acknowledgements

- [antlr/antlr4](https://github.com/antlr/antlr4) · [antlr4-go/antlr](https://github.com/antlr/antlr4/tree/master/runtime/Go/antlr) — DSL parser runtime
- [sony/gobreaker](https://github.com/sony/gobreaker) — circuit breaker
- [prometheus/client_golang](https://github.com/prometheus/client_golang) — metrics
- [redis/go-redis](https://github.com/redis/go-redis) — Redis client
- [gin-gonic/gin](https://github.com/gin-gonic/gin) — HTTP framework
- Design inspired by risk engineering practices at Bilibili, ByteDance, and Meituan
