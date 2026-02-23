# RiskEngine

[![Go Version](https://img.shields.io/badge/go-1.24+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-Apache%202.0-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/zhucl121/risk-engine)](https://goreportcard.com/report/github.com/zhucl121/risk-engine)
[![CI](https://github.com/zhucl121/risk-engine/actions/workflows/ci.yml/badge.svg)](https://github.com/zhucl121/risk-engine/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/zhucl121/risk-engine/graph/badge.svg)](https://codecov.io/gh/zhucl121/risk-engine)

**RiskEngine** is a high-performance, open-source risk decision engine written in Go. It is designed for real-time fraud detection in payment, marketing promotion, and transaction scenarios. It delivers **P99 < 60ms** decisions at **20,000+ TPS** on commodity hardware.

> 中文文档：[README_zh.md](README_zh.md)

---

## Features

### Core Decision

| Feature | Description |
|---------|-------------|
| **Multi-strategy decision** | Rule engine + ML model scoring + list service — orchestrated in a configurable DAG pipeline |
| **Hot-reload** | Rules and models update without service restart; change propagates in < 30 s |
| **Self-hosted RiskDSL** | ANTLR4-Go expression language compiled to Go closures at load time; P99 29 ns (simple), 97 ns (with features), 0 allocs per evaluation |
| **Parallel feature fetching** | All feature sources queried concurrently; per-source timeout degradation never blocks the decision |
| **Standalone Feature Store** | Optional external feature service called over gRPC; `VelocityGroup` (sliding-window counters) and `UserProfileGroup` (Redis JSON hash); fail-open on timeout |
| **Velocity counters** | Redis Lua atomic sliding-window counters at any granularity (1 min / 1 h / 24 h) |
| **List service** | Redis-backed blacklist / graylist / whitelist with O(1) lookup |

### Policy Orchestration

| Feature | Description |
|---------|-------------|
| **A/B testing** | Random traffic splitting; experiment group tagged in `RiskReasons`; no restart required |
| **Canary routing** | Deterministic SHA-256 hash–based routing — the same user always lands in the same bucket (stable routing); supports `userID`, `deviceID`, `sessionID`, `ip`, or `extra.<key>` as the hash key; per-experiment salt prevents bucket correlation |
| **Shadow / dry-run mode** | New policies execute in parallel without affecting the production decision; full results written to `shadow_audit` log for offline analysis before promotion |
| **Champion-challenger** | Multiple challenger pipelines run concurrently in the background; only the champion decision is returned; both champion and challenger results written to `cc_audit` with an `agreement` flag for statistical comparison |
| **Step-level condition** | Each step accepts a DSL `condition` expression; step is skipped when the expression evaluates to `false` |
| **Automatic retry** | Per-step `maxAttempts` + `delayMs` retry on transient failure |
| **Aggregation strategies** | `HIGHEST_RISK` (default) / `WEIGHTED` (weighted sum) / `RULE_FIRST` (rule priority, model as fallback) |
| **Failure policies** | Per-step `onFailure`: `SKIP` (ignore) / `REJECT` (high-risk default) / `FALLBACK` (scene-level default decision) |
| **Circuit breaker** | `gobreaker`-backed per-step breakers (list, model); state exposed as Prometheus gauge |

### Parameters & Data

| Feature | Description |
|---------|-------------|
| **Extra parameter spec management** | Per-scene Extra field rules stored in MySQL — required fields (missing → 400 rejected) or optional fields with a default value; hot-reloaded every 30 s |
| **Extra → feature injection** | `DecisionRequest.Extra` fields automatically injected into `feature.Map` as `extra.<key>` with DB-driven type coercion (string / int / float / bool); per-step `ParamMapping` remaps fields for downstream services |
| **Domain-agnostic design** | No business-domain fields on the request struct; amount, order type, merchant ID, etc. all live in `extra` — making the engine suitable for payment, marketing, login, and any other risk scenario |

### Observability & Operations

| Feature | Description |
|---------|-------------|
| **Rate limiting** | Two-tier token bucket: global 5000 RPS + per-IP 100 RPS; HTTP 429 on exhaustion |
| **Observability** | Prometheus metrics (decision latency, rule hits, feature errors, active requests), OpenTelemetry tracing headers, structured Zap logging, async audit writer |
| **Health probes** | `/api/v1/livez` (liveness) and `/api/v1/readyz` (readiness with Redis check) for Kubernetes |
| **Dual protocol** | HTTP/JSON (Gin) + gRPC (`DecisionService`: Evaluate / BatchEvaluate / Health) |
| **Cloud-native** | Graceful shutdown with configurable drain timeout; Kubernetes-ready |

---

## Quick Start

### Prerequisites

- Go 1.24+
- Redis 7+
- Kafka 3+ (for audit trail; optional — dev mode uses structured-log fallback)

### Run locally

```bash
git clone https://github.com/zhucl121/risk-engine.git
cd riskengine

# copy and edit config
cp configs/config.example.yaml configs/config.local.yaml

# start dependencies
docker compose -f deployments/docker/compose.dev.yaml up -d

# run
go run ./cmd/server -config configs/config.local.yaml
```

### Make a decision

All business fields (including amount) are passed via `extra` — the engine imposes no domain-specific top-level fields:

```bash
curl -X POST http://localhost:8080/api/v1/decision \
  -H "Content-Type: application/json" \
  -d '{
    "scene_code": "PAYMENT_CHECKOUT",
    "user_id":    "u123456",
    "device_id":  "d-abc-def",
    "ip":         "1.2.3.4",
    "extra": {
      "amount":       "9900",
      "merchant_id":  "M001",
      "product_type": "GOODS"
    }
  }'
```

Response:

```json
{
  "request_id":  "01HZ...",
  "decision":    "PASS",
  "risk_score":  120,
  "risk_level":  "LOW",
  "hit_rules":   [],
  "model_scores": {"payment_fraud_xgb": 0.08},
  "risk_reasons": [],
  "cost_ms": 23
}
```

| Decision | Meaning |
|----------|---------|
| `PASS` | Approved — allow the action |
| `REJECT` | Denied — hit a high-risk rule or blacklist |
| `MANUAL_REVIEW` | Escalated — graylist match or low-confidence rule hit |

---

## Architecture

```
Request → API Layer (Gin HTTP / gRPC)
                ↓
   RateLimit / Metrics / Tracing Middleware
                ↓
          DecisionEngine
                ↓
          Orchestrator (DAG)
           ├─ Routing: Canary > A/B Test > Main Pipeline
           │
           ├─ Main Pipeline (sequential + parallel steps)
           │    ├── LIST   → List Service (Redis + circuit breaker)
           │    ├── RULE   → Rule Engine (RiskDSL)
           │    └── MODEL  → Model Engine (ONNX + circuit breaker)
           │
           ├─ Shadow Pipeline (background, never affects main decision)
           │    └── → shadow_audit log
           │
           └─ Champion-Challenger (background concurrent, challenger not returned)
                └── → cc_audit log

  FeatureService (parallel fetching)
   ├── VelocityFetcher (sliding-window counter) → Redis
   ├── UserProfile → Redis JSON
   └── FeatureStoreFetcher → gRPC Feature Store

AuditWriter (async channel → structured log / Kafka)
  ├── audit       (main decision records)
  ├── shadow_audit (shadow / dry-run records)
  └── cc_audit    (champion-challenger records)
```

- **Circuit breaker (CB)**: trips after 5 consecutive failures, probes after 30 s
- **Routing priority**: Canary (hash-stable) → A/B test (random) → Main pipeline; at most one is active per request
- **Shadow / CC**: always run as independent background goroutines, never compete with routing

For a detailed architecture and design rationale, see [docs/architecture.md](docs/architecture.md).

---

## Core Concepts

### PolicySet

Each scene (`scene_code`) maps to one `PolicySet` that defines its full decision pipeline in YAML:

```yaml
- sceneCode: payment
  version: "1.0.0"
  fallback: MANUAL_REVIEW       # decision returned when pipeline cannot complete
  strategy: HIGHEST_RISK        # HIGHEST_RISK | WEIGHTED | RULE_FIRST
  extraSchema:
    amount:      int            # static type hint (DB spec takes precedence)
    merchant_id: string

  pipeline:
    - name: blacklist_check
      kind: LIST
      timeoutMs: 20
      onFailure: SKIP
      listQueryFields:          # custom query dimensions (default: user/device/ip)
        - extra.merchant_id
        - request.ip

    - name: payment_rules
      kind: RULE
      ruleGroup: payment
      timeoutMs: 50
      condition: "extra.amount > 0"   # DSL condition; step skipped when false
      retry:
        maxAttempts: 2
        delayMs: 5

    - name: risk_model
      kind: MODEL
      models: [payment_fraud_v2]
      timeoutMs: 80
      weight: 0.7               # used by WEIGHTED strategy
      params:                   # per-step parameter mapping
        merchant: extra.merchant_id
        channel:  "WEB"

  # A/B test (random per request — non-sticky)
  abTest:
    enabled: true
    experimentId: payment-model-v3
    splitPct: 0.05
    experimentPipeline:
      - name: model_v3
        kind: MODEL
        models: [payment_fraud_v3]

  # Canary routing (hash-stable — same user always in same bucket)
  canary:
    enabled: true
    canaryVersion: "v2.1.0"
    trafficPct: 10              # 10 % of users routed to canary
    hashKey: userID             # userID | deviceID | sessionID | ip | extra.<key>
    salt: "payment_canary_v2"   # unique salt per experiment
    canaryPipeline:
      - name: new_rules
        kind: RULE
        ruleGroup: payment_v2
      - name: model_v3
        kind: MODEL
        models: [payment_fraud_v3]

  # Shadow / dry-run (parallel, no effect on production decision)
  shadowPolicies:
    - sceneCode: payment_new_policy
      version: "draft-1"

  # Champion-challenger (parallel background execution; results to cc_audit only)
  championChallenger:
    enabled: true
    experimentID: "fraud_model_v3_eval"
    challengers:
      - challengerID: "model_v3_candidate"
        trafficPct: 20
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

### Routing Priority

```
Canary (hash-stable, user-level)
  ↓ not matched
A/B Test (random, request-level)
  ↓ not matched
Main Pipeline
```

Exactly one branch is active per request. Shadow and Champion-Challenger always run as independent background goroutines (within their own `trafficPct`) and never interfere with routing.

### Mode Comparison

| Mode | Affects production decision | Routing | Use case |
|------|-----------------------------|---------|----------|
| **A/B Test** | Yes — experiment group runs different pipeline | Random (per request) | Symmetric experiments; both groups are production-ready |
| **Canary** | Yes — canary users run new pipeline | Hash-stable (same user, same bucket) | Gradual rollout; incrementally expand new policy coverage |
| **Shadow** | No | All requests (or specific scene) | Pre-release validation; offline comparison analysis |
| **Champion-Challenger** | No — challenger result discarded | Hash-stable (per `trafficPct`) | Policy evaluation; statistical significance testing |

---

### RiskDSL

Rule conditions are written in the self-hosted RiskDSL. Expressions are compiled to Go closures at rule load time by an ANTLR4-generated parser — zero allocations at evaluation time.

#### Operators

```
# Comparison
extra.amount > 10000
risk_score != 0

# Boolean logic
extra.amount > 5000 && velocity("pay", user_id, "1h") > 10
extra.amount > 1000 || inList("blacklist_ip", ip)

# in / not in (set membership)
extra.product_type in ["DIGITAL", "VIRTUAL"]
extra.channel not in ["OFFLINE", "STORE"]

# Ternary
extra.vip_level == "gold" ? 200 : 500

# Negation
!isEmpty(user_id)
```

#### Built-in Functions

**String**

| Function | Description |
|----------|-------------|
| `contains(s, sub)` | Substring check |
| `startsWith(s, prefix)` | Prefix match |
| `endsWith(s, suffix)` | Suffix match |
| `match(s, pattern)` | Regex match (compiled & cached) |
| `lower(s)` / `upper(s)` | Case conversion |
| `trim(s)` | Strip leading/trailing whitespace |
| `strlen(s)` | String length |
| `isEmpty(s)` | Empty check |

**Math**

| Function | Description |
|----------|-------------|
| `abs(n)` | Absolute value |
| `ceil(n)` / `floor(n)` / `round(n)` | Rounding |
| `sqrt(n)` | Square root |
| `min(a, b)` / `max(a, b)` | Minimum / maximum |
| `clamp(n, lo, hi)` | Clamp to range |

**Time**

| Function | Description |
|----------|-------------|
| `now()` | Current Unix timestamp (seconds) |
| `nowMs()` | Current millisecond timestamp |
| `daysSince(t)` | Days elapsed since time string `t` |
| `hoursSince(t)` | Hours elapsed since `t` |
| `secondsSince(t)` | Seconds elapsed since `t` |
| `toUnix(t)` | Parse time string to Unix timestamp |
| `hour(t)` | Extract hour (0–23) |
| `weekday(t)` | Day of week (0 = Sunday, 6 = Saturday) |

**Type conversion & conditional**

| Function | Description |
|----------|-------------|
| `toInt(v)` / `toFloat(v)` / `toString(v)` / `toBool(v)` | Type conversion |
| `isNull(v)` | Null / zero-value check |
| `coalesce(a, b, ...)` | First non-null value |
| `ifThen(cond, a, b)` | Conditional select (equivalent to ternary) |

**Risk-specific**

| Function | Description |
|----------|-------------|
| `inList(kind, value)` | Query list service (blacklist / graylist / whitelist) |
| `velocity(prefix, id, window)` | Read sliding-window counter |
| `modelScore(name)` | ML model score (0–1) |
| `geoIP(ip)` | Country code for IP |
| `within(lat, lon, clat, clon, km)` | Geofence check |

**Example rules**

```
# High-frequency payment (> 10 payments in 1 hour)
velocity("pay_count", user_id, "1h") > 10 && extra.amount > 5000

# Blacklist + amount threshold
inList("blacklist_user", user_id) || (extra.amount > 50000 && extra.vip_level not in ["gold", "platinum"])

# Late-night large transaction (11 PM – 6 AM)
extra.amount > 10000 && (hour(now()) >= 23 || hour(now()) <= 6)

# New device + large transfer
daysSince(extra.device_register_time) < 7 && extra.amount > 20000

# Ternary: VIP users allowed more failures
velocity("pay_fail", user_id, "24h") > (extra.vip_level == "gold" ? 20 : 5)
```

Full DSL syntax reference: [docs/dsl-guide.md](docs/dsl-guide.md)

---

## Feature Store

Feature fetching supports two modes that can be used independently or together.

### Comparison

```
Mode 1 (default) — in-process fetchers, direct Redis access
RiskEngine process
  └── feature.Service
        ├── VelocityFetcher  ──→ Redis
        └── ...

Mode 2 — standalone Feature Store, called over gRPC
RiskEngine process                Feature Store process
  └── feature.Service              └── FeatureStoreService
        └── FeatureStoreFetcher ─gRPC→  ├── VelocityGroup   → Redis
                                        └── UserProfileGroup → Redis JSON
```

### Start the Feature Store

```bash
go run ./cmd/featurestore -config configs/config.yaml
```

### Connect the engine

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

### Custom FeatureGroup

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

Register:

```go
store.DefaultRegistry.Register(&MyGroup{db: db})
```

---

## API Reference

### HTTP API

#### Make a decision

```
POST /api/v1/decision
```

| Field | Type | Description |
|-------|------|-------------|
| `scene_code` | string | Scene identifier (required) |
| `user_id` | string | User identifier |
| `device_id` | string | Device identifier |
| `session_id` | string | Session identifier |
| `ip` | string | Client IP address |
| `extra` | map[string]string | Scene-specific fields (amount, merchant_id, etc.) |

> **Note**: Amount and other business-domain fields belong in `extra`, e.g. `"extra": {"amount": "9900"}`. The engine type-coerces them according to the scene's `extraSchema`.

#### Health

```
GET /api/v1/livez    # liveness — returns 200 if process is alive
GET /api/v1/readyz   # readiness — returns 503 if dependencies are unavailable
GET /metrics         # Prometheus metrics
```

#### Rule management (Admin)

```
GET    /admin/v1/rules                             # list rules
POST   /admin/v1/rules                             # create rule
PUT    /admin/v1/rules/:id                         # update rule
DELETE /admin/v1/rules/:id                         # delete rule
POST   /admin/v1/rules/:id/enable                  # enable rule
POST   /admin/v1/rules/:id/disable                 # disable rule
POST   /admin/v1/rules/validate                    # validate DSL expression
GET    /admin/v1/scenes/:scene/extra-params        # list extra param specs for scene
POST   /admin/v1/scenes/:scene/extra-params        # create extra param spec
PUT    /admin/v1/scenes/:scene/extra-params/:key   # update extra param spec
DELETE /admin/v1/scenes/:scene/extra-params/:key   # delete extra param spec
```

### gRPC API

See [api/grpc/proto/decision.proto](api/grpc/proto/decision.proto). Default port `:9090`.

```protobuf
service DecisionService {
  rpc Evaluate(DecisionRequest) returns (DecisionResponse);
  rpc BatchEvaluate(BatchDecisionRequest) returns (BatchDecisionResponse);
  rpc Health(HealthRequest) returns (HealthResponse);
}
```

---

## Extending RiskEngine

### Add a rule

1. Implement `rule.Rule` in `internal/rule/rules/<name>.go`
2. Register in `internal/rule/registry.go`
3. Add YAML config in `configs/rules/<group>.yaml`
4. Write unit tests and a benchmark

See [docs/adding-rules.md](docs/adding-rules.md) for a step-by-step walkthrough.

### Add a feature fetcher

1. Implement `feature.Fetcher`:

```go
type Fetcher interface {
    Name()    string
    Timeout() time.Duration
    Fetch(ctx context.Context, req *engine.DecisionRequest) (feature.Map, error)
}
```

2. Register in `cmd/server/main.go`:

```go
featureSvc.Register(fetchers.NewMyFetcher(deps))
```

Reference implementation: `internal/feature/fetchers/velocity_fetcher.go`

### Configure Extra parameter specs (database)

Run the migration first:

```bash
mysql -u root riskengine < configs/migrations/002_create_scene_extra_params.sql
```

Manage specs via the admin API:

```bash
# Required field (missing → HTTP 400)
POST /admin/v1/scenes/payment/extra-params
{
  "param_key":   "merchant_id",
  "param_type":  "string",
  "required":    true,
  "description": "Merchant identifier — required"
}

# Optional field with default
POST /admin/v1/scenes/payment/extra-params
{
  "param_key":   "product_type",
  "param_type":  "string",
  "required":    false,
  "default_val": "GOODS",
  "description": "Product category, defaults to GOODS"
}
```

| Scenario | Behaviour |
|----------|-----------|
| Required field absent | `ErrMissingRequiredExtra` (HTTP 400) with field name and scene |
| Optional field absent, has default | `default_val` injected before type-coercion into `feature.Map` |
| Optional field absent, no default | Not injected; DSL reads `extra.<key>` as zero value |
| Field present | Type-coerced directly; default value skipped |

DB specs are merged with the static YAML `extraSchema`; **DB takes precedence**. Specs are cached in memory and background-reloaded every 30 s.

### Configure Canary routing

Canary uses SHA-256 hash modulo to guarantee the same user always lands in the same bucket:

```yaml
canary:
  enabled: true
  canaryVersion: "v2.1.0"
  trafficPct: 10          # integer 0–100; exact user percentage
  hashKey: userID         # userID | deviceID | sessionID | ip | extra.<key>
  salt: "payment_canary_v2"   # unique per experiment to avoid bucket correlation
  canaryPipeline:
    - name: new_rule_engine
      kind: RULE
      ruleGroup: payment_v2
```

**Ramp-up flow**: 5 % → 10 % → 25 % → 50 % → 100 %. Adjust `trafficPct` in the management UI; no restart required.

### Configure Champion-Challenger

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

Analyse agreement rate from `cc_audit` log:

```bash
jq 'select(.experiment_id == "fraud_model_eval_q1") | .agreement' cc_audit.log | \
  awk '{a[$0]++} END {print "agree:", a["true"], "disagree:", a["false"]}'
```

### Configure Shadow mode

```yaml
shadowPolicies:
  - sceneCode: payment_new_strategy
    version: "draft-2025Q1"
```

Shadow results appear in `shadow_audit` with `production_decision` and `decision` fields side-by-side:

```bash
jq '{req: .request_id, prod: .production_decision, shadow: .decision}' shadow_audit.log
```

### Add an ML model

1. Export your model to ONNX
2. Place the `.onnx` file in `configs/models/`
3. Register metadata in `configs/models/registry.yaml`

The engine hot-loads new or updated models without restart.

### Extend RiskDSL with custom functions

```go
registry.RegisterFunc("myRiskScore",
    func(ctx context.Context, rt *dsl.Runtime, args []dsl.Value) (dsl.Value, error) {
        userID := rt.Request.UserID
        _ = userID
        return dsl.IntValue(42), nil
    })
```

---

## Performance

Benchmarks on an 8-core / 32 GB VM (Go 1.24, Redis 7 local):

| Scenario | P50 | P99 | TPS |
|----------|-----|-----|-----|
| List check only | 0.8 ms | 1.5 ms | 80,000 |
| Rules only (100 rules) | 3 ms | 8 ms | 45,000 |
| Rules + model scoring | 15 ms | 35 ms | 25,000 |
| Full pipeline (list + rules + model) | 22 ms | 55 ms | 20,000 |

DSL expression benchmarks (single core, no Redis):

| Scenario | Latency | Allocations |
|----------|---------|-------------|
| Simple condition (`extra.amount > 1000`) | P99 29 ns | 0 allocs |
| Feature read (`velocity(...) > 10`) | P99 97 ns | 0 allocs |
| `in` set check (5 elements) | P99 43 ns | 0 allocs |

```bash
make bench
```

---

## Project Structure

```
riskengine/
├── cmd/
│   ├── server/            # Main HTTP + gRPC server entry point
│   └── featurestore/      # Standalone Feature Store gRPC server
├── internal/
│   ├── engine/            # Top-level DecisionEngine, request lifecycle
│   ├── rule/              # Rule storage, evaluator, hot-reload
│   ├── feature/           # Parallel feature fetching
│   │   └── fetchers/      # Concrete fetchers (VelocityFetcher, ...)
│   ├── featurestore/      # gRPC Feature Store client + server implementation
│   │   └── store/         # FeatureGroup registry (VelocityGroup, UserProfileGroup)
│   ├── model/             # Model registry, ONNX scorer interface
│   ├── list/              # Redis-backed list service (blacklist / graylist)
│   ├── orchestrator/      # DAG executor, policy routing (A/B + Canary),
│   │                      # shadow mode, champion-challenger, Extra injection
│   ├── scene/             # Per-scene Extra param specs (DB-backed, hot-reloaded)
│   ├── audit/             # Async channel audit writer (main / shadow_audit / cc_audit)
│   ├── metrics/           # Prometheus collectors
│   ├── middleware/        # Gin middleware (RequestID, Metrics, RateLimit, Logger, Tracing)
│   ├── health/            # Liveness / readiness checkers
│   ├── resilience/        # Circuit breaker (gobreaker wrapper)
│   └── config/            # Config loader
├── pkg/
│   ├── dsl/               # Self-hosted RiskDSL (ANTLR4-Go, AST, codegen)
│   │   ├── grammar/       # RiskDSL.g4 grammar file
│   │   ├── parser/        # ANTLR4-generated parser (do not edit)
│   │   ├── ast/           # AST node types
│   │   └── builtins/      # Built-in functions (string / math / time / convert / risk)
│   ├── sliding/           # Redis Lua sliding-window velocity counter
│   ├── bloom/             # In-process Bloom filter
│   └── pool/              # Object pool utilities
├── api/
│   ├── grpc/
│   │   ├── proto/         # decision.proto
│   │   ├── v1/            # protoc-generated Go (decision.pb.go, *_grpc.pb.go)
│   │   └── server/        # DecisionServer implementation
│   └── http/
│       ├── v1/            # Decision API, health probes
│       └── admin/v1/      # Rule management + Extra param spec management
├── configs/
│   ├── config.example.yaml
│   ├── migrations/        # Database migration scripts (SQL)
│   └── policies/          # PolicySet YAML files (loaded at startup)
├── deployments/           # Docker and Kubernetes manifests
├── docs/                  # Architecture docs, DSL guide, design documents
└── openspec/              # Change proposals and design specs
```

---

## Contributing

We welcome contributions! Please read [CONTRIBUTING.md](CONTRIBUTING.md) before submitting a pull request.

### Development setup

```bash
make setup     # install tools (golangci-lint, mockery, protoc)
make test      # run unit tests
make lint      # run linter
make bench     # run benchmarks
```

### Coding conventions

- One `git commit` per complete functional module
- Non-trivial features require a proposal in `openspec/changes/<name>/proposal.md` first
- Touching interfaces under `engine/`, `rule/`, `feature/`, `model/`, or `list/` requires a `design.md` with an Architect REVIEW section
- Implementation code must correspond to an unchecked `[ ]` task in `tasks.md`

---

## License

Apache License 2.0. See [LICENSE](LICENSE) for details.

---

## Acknowledgements

- [antlr/antlr4](https://github.com/antlr/antlr4) + [antlr4-go/antlr](https://github.com/antlr/antlr4/tree/master/runtime/Go/antlr) — RiskDSL parser runtime
- [bits-and-blooms/bloom](https://github.com/bits-and-blooms/bloom) — Bloom filter
- [sony/gobreaker](https://github.com/sony/gobreaker) — circuit breaker
- [prometheus/client_golang](https://github.com/prometheus/client_golang) — metrics
- [redis/go-redis](https://github.com/redis/go-redis) — Redis client
- Design inspired by industry practices at Bilibili, ByteDance, and Meituan risk teams
