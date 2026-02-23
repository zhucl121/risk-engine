# RiskEngine

[![Go Version](https://img.shields.io/badge/go-1.24+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-Apache%202.0-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/yourorg/riskengine)](https://goreportcard.com/report/github.com/yourorg/riskengine)
[![CI](https://github.com/yourorg/riskengine/actions/workflows/ci.yml/badge.svg)](https://github.com/yourorg/riskengine/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/yourorg/riskengine/graph/badge.svg)](https://codecov.io/gh/yourorg/riskengine)

**RiskEngine** is a high-performance, open-source risk decision engine written in Go. It is designed for real-time fraud detection in payment, marketing promotion, and transaction scenarios. It delivers **P99 < 60ms** decisions at **20,000+ TPS** on commodity hardware.

> 中文文档：[README_zh.md](README_zh.md)

---

## Features

- **Multi-strategy decision**: Rule engine + ML model scoring + list service — orchestrated in a configurable DAG pipeline
- **Hot-reload**: Rules and models update without service restart; change propagates in < 30 seconds
- **Self-hosted RiskDSL**: Custom ANTLR4-Go expression language compiled to Go closures at load time; P99 29ns (simple), 97ns (with features), 0 allocs per evaluation
- **Parallel feature fetching**: All feature sources queried concurrently; per-source timeout degradation never blocks the decision
- **Standalone Feature Store**: Optional external feature service called over gRPC; supports `VelocityGroup` (sliding-window counters) and `UserProfileGroup` (Redis JSON hash); fail-open on timeout
- **Velocity counters**: Redis Lua atomic sliding-window counters at any granularity (1min / 1hour / 24h); backed by `pkg/sliding`
- **List service**: Redis-backed blacklist / graylist / whitelist with O(1) lookup
- **Extra parameter spec management**: Per-scene Extra field rules stored in MySQL — mark fields as required (missing → request rejected) or optional with a default value; hot-reloaded every 30 s with a background watcher
- **Extra → feature injection**: `DecisionRequest.Extra` fields are automatically injected into `feature.Map` as `extra.<key>` with DB-driven type coercion (string / int / float / bool); per-step `ParamMapping` remaps fields for downstream services
- **A/B testing**: Traffic-split routing per `PolicySet.ABTest`; experiment group tagged in `RiskReasons`; no restart required
- **Circuit breaker**: `gobreaker`-backed per-step breakers (list, model); state exposed as Prometheus gauge
- **Rate limiting**: Two-tier token bucket — global (5000 RPS) + per-IP (100 RPS); HTTP 429 on exhaustion
- **Observability**: Prometheus metrics (decision latency, rule hits, feature errors, active requests), OpenTelemetry tracing headers, structured Zap logging, async audit writer
- **Health probes**: `/api/v1/livez` (liveness) and `/api/v1/readyz` (readiness with Redis check) for Kubernetes
- **Dual protocol**: HTTP/JSON (Gin) + gRPC (`DecisionService`: Evaluate / BatchEvaluate / Health)
- **Cloud-native**: Graceful shutdown with configurable drain timeout; Kubernetes-ready

---

## Quick Start

### Prerequisites

- Go 1.24+
- Redis 7+
- Kafka 3+ (for audit trail; optional — dev mode uses structured-log fallback)

### Run locally

```bash
git clone https://github.com/yourorg/riskengine.git
cd riskengine

# copy and edit config
cp configs/config.example.yaml configs/config.local.yaml

# start dependencies
docker compose -f deployments/docker/compose.dev.yaml up -d

# run
go run ./cmd/server -config configs/config.local.yaml
```

### Make a decision

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

Response:

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

---

## Architecture

```
Request → API Layer (Gin HTTP + gRPC)
                  ↓
     RateLimit / Metrics / Tracing Middleware
                  ↓
          DecisionEngine
                  ↓
          Orchestrator (DAG + A/B routing)
         ↙         ↓            ↘
   ListSvc    RuleEngine     ModelEngine
  (Redis+CB)  (RiskDSL)     (ONNX+CB)
         ↘         ↓            ↙
          FeatureService (parallel)
          ↙        ↓           ↘
   VelocityFetcher  UserProfile  DeviceInfo
   (sliding.Window)
                  ↓
           Redis (primary store)
                  ↓
         AuditWriter (async channel → log / Kafka)
```

- **CB** = circuit breaker (gobreaker); trips after 5 consecutive failures, recovers after 30s
- **A/B routing**: `PolicySet.ABTest.SplitPct` fraction of traffic runs `ExperimentPipeline`

For a detailed architecture and design rationale, see [docs/architecture.md](docs/architecture.md).

---

## Configuration

All configuration is YAML-driven. See [configs/config.example.yaml](configs/config.example.yaml) for the full reference with inline comments.

Key sections:

| Section | Description |
|---------|-------------|
| `server.addr` | HTTP listen address (default `:8080`) |
| `server.grpc_addr` | gRPC listen address (default `:9090`) |
| `redis` | Connection pool, cluster / sentinel mode |
| `kafka` | Brokers, audit topic (optional) |
| `engine.policy_dir` | Directory of `*.yaml` PolicySet files |
| `rules` | Rule groups, hot-reload interval |
| `models` | ONNX model paths |
| `feature.redis_timeout` | Per-fetcher Redis timeout (default `10ms`) |

---

## Extending RiskEngine

### Add a rule

1. Implement `rule.Rule` in `internal/rule/rules/<name>.go`
2. Register in `internal/rule/registry.go`
3. Add YAML configuration in `configs/rules/<group>.yaml`
4. Write unit tests and a benchmark

See [docs/adding-rules.md](docs/adding-rules.md) for a step-by-step walkthrough.

### Add a feature fetcher

1. Implement `feature.Fetcher` in `internal/feature/fetchers/<name>.go`
2. Register via `featureSvc.Register(...)` in `cmd/server/main.go`

See `internal/feature/fetchers/velocity_fetcher.go` as a reference implementation.

### Use the standalone Feature Store

Enable in `configs/config.yaml`:

```yaml
feature_store:
  enabled: true
  addr: "localhost:50052"
  dial_timeout: 3s
  request_timeout: 50ms
  groups:
    - name: velocity      # must match a registered FeatureGroup name
      timeout: 20ms
    - name: user_profile
      timeout: 30ms
```

Start the Feature Store server:

```bash
go run ./cmd/featurestore -config configs/config.yaml
```

Implement a custom `FeatureGroup` by satisfying `store.FeatureGroup` and registering it with `store.DefaultRegistry`.

### Configure Extra parameter specs (database)

Each scene's `Extra` fields can be declared in the `scene_extra_params` table (run `configs/migrations/002_create_scene_extra_params.sql` first).
Use the admin API to manage specs at runtime:

```bash
# List specs for a scene
GET /admin/v1/scenes/payment/extra-params

# Create / update a spec
POST /admin/v1/scenes/payment/extra-params
{
  "param_key":   "merchant_id",
  "param_type":  "string",
  "required":    true,
  "description": "Merchant identifier — required"
}

# Optional field with a default value
POST /admin/v1/scenes/payment/extra-params
{
  "param_key":   "product_type",
  "param_type":  "string",
  "required":    false,
  "default_val": "GOODS",
  "description": "Product category, defaults to GOODS"
}
```

When a required field is absent the engine returns `ErrMissingRequiredExtra` (HTTP 400). When an optional field is absent its `default_val` is filled in before type-coercion and feature injection.

### Add an ML model

1. Export your model to ONNX format
2. Place the `.onnx` file in `configs/models/`
3. Register metadata in `configs/models/registry.yaml`

The engine hot-loads new/updated models without restart.

---

## Performance

Benchmarks run on an 8-core / 32 GB VM (Go 1.24, Redis 7 local):

| Scenario | P50 | P99 | TPS |
|----------|-----|-----|-----|
| List check only | 0.8ms | 1.5ms | 80,000 |
| Rules only (100 rules) | 3ms | 8ms | 45,000 |
| Rules + model scoring | 15ms | 35ms | 25,000 |
| Full pipeline (list + rules + model) | 22ms | 55ms | 20,000 |

Run benchmarks locally:

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
├── internal/              # Private application code
│   ├── engine/            # Top-level DecisionEngine, request lifecycle
│   ├── rule/              # Rule storage, evaluator, hot-reload
│   ├── feature/           # Parallel feature fetching
│   │   └── fetchers/      # Concrete fetchers (VelocityFetcher, ...)
│   ├── featurestore/      # gRPC Feature Store client, fetcher adapter, server impl
│   │   └── store/         # FeatureGroup registry (VelocityGroup, UserProfileGroup)
│   ├── model/             # Model registry, ONNX scorer interface
│   ├── list/              # Redis-backed list service (blacklist/graylist)
│   ├── orchestrator/      # DAG executor, A/B routing, policy registry, Extra injection
│   ├── scene/             # Per-scene Extra param specs (DB-backed, hot-reloaded)
│   ├── audit/             # Async channel audit writer (→ log / Kafka)
│   ├── metrics/           # Prometheus collectors
│   ├── middleware/         # Gin middleware (RequestID, Metrics, RateLimit, Logger, Tracing)
│   ├── health/            # Liveness / readiness checkers
│   ├── resilience/        # Circuit breaker (gobreaker wrapper)
│   └── config/            # Config loader
├── pkg/                   # Public, reusable packages
│   ├── dsl/               # Self-hosted RiskDSL (ANTLR4-Go, AST, codegen)
│   │   ├── grammar/       # RiskDSL.g4 grammar file
│   │   ├── parser/        # ANTLR4-generated parser (do not edit)
│   │   ├── ast/           # AST node types
│   │   └── builtins/      # Built-in risk functions (inList, velocity, ...)
│   ├── sliding/           # Redis Lua sliding-window velocity counter
│   ├── bloom/             # In-process Bloom filter
│   └── pool/              # Object pool utilities
├── api/
│   ├── grpc/              # proto definitions + generated code + server impl
│   │   ├── proto/         # decision.proto
│   │   ├── v1/            # protoc-generated Go (decision.pb.go, *_grpc.pb.go)
│   │   └── server/        # DecisionServer implementation
│   └── http/              # Gin HTTP handlers
│       ├── v1/            # Decision API, health, livez, readyz
│       └── admin/v1/      # Rule management CRUD API
├── configs/
│   ├── config.example.yaml
│   ├── migrations/        # Database migration scripts (SQL)
│   └── policies/          # PolicySet YAML files (loaded at startup)
├── deployments/           # Docker and Kubernetes manifests
├── docs/                  # Architecture docs, design documents
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

---

## License

Apache License 2.0. See [LICENSE](LICENSE) for details.

---

## Acknowledgements

- [antlr/antlr4](https://github.com/antlr/antlr4) + [antlr4-go/antlr](https://github.com/antlr/antlr4/tree/master/runtime/Go/antlr) – RiskDSL parser runtime
- [bits-and-blooms/bloom](https://github.com/bits-and-blooms/bloom) – Bloom filter
- [sony/gobreaker](https://github.com/sony/gobreaker) – circuit breaker
- [prometheus/client_golang](https://github.com/prometheus/client_golang) – metrics
- [redis/go-redis](https://github.com/redis/go-redis) – Redis client
- Design inspired by industry practices at Bilibili, ByteDance, and Meituan risk teams
