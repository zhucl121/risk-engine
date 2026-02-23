# RiskEngine

[![Go Version](https://img.shields.io/badge/go-1.22+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-Apache%202.0-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/yourorg/riskengine)](https://goreportcard.com/report/github.com/yourorg/riskengine)
[![CI](https://github.com/yourorg/riskengine/actions/workflows/ci.yml/badge.svg)](https://github.com/yourorg/riskengine/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/yourorg/riskengine/graph/badge.svg)](https://codecov.io/gh/yourorg/riskengine)

**RiskEngine** is a high-performance, open-source risk decision engine written in Go. It is designed for real-time fraud detection in payment, marketing promotion, and transaction scenarios. It delivers **P99 < 60ms** decisions at **20,000+ TPS** on commodity hardware.

---

## Features

- **Multi-strategy decision**: Rule engine + ML model scoring + tiered list service + graph risk — orchestrated in a configurable DAG pipeline
- **Hot-reload**: Rules and models update without service restart; change propagates in < 30 seconds
- **Expression DSL**: YAML-based rule configuration powered by a type-safe expression evaluator; no DRL expertise required
- **Parallel feature fetching**: All feature sources queried concurrently; per-source timeout degradation never blocks the decision
- **Velocity counters**: Redis Lua atomic sliding-window counters for rate limiting at any granularity (1min / 1hour / 24h)
- **Tiered list lookup**: In-process Bloom filter (L1) → Redis (L2) → persistent store (L3); typical list check < 1ms
- **Champion-Challenger**: Traffic-split A/B testing for rules and ML models with statistical significance tracking
- **Observability**: Prometheus metrics, OpenTelemetry tracing, structured Zap logging, Kafka audit trail
- **Cloud-native**: Docker image < 50MB; Kubernetes HPA manifests included; graceful shutdown with drain timeout

---

## Quick Start

### Prerequisites

- Go 1.22+
- Redis 7+
- Kafka 3+ (for audit trail; optional in dev mode)

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
Request → API Layer (Gin/gRPC)
               ↓
         Orchestrator (DAG)
        ↙       ↓        ↘
  ListSvc   RuleEngine  ModelEngine
        ↘       ↓        ↙
         FeatureService
               ↓
    Redis / HBase / External APIs
               ↓
         AuditWriter → Kafka
```

For a detailed architecture and design rationale, see [docs/architecture.md](docs/architecture.md).

---

## Configuration

All configuration is YAML-driven. See [configs/config.example.yaml](configs/config.example.yaml) for the full reference with inline comments.

Key sections:

| Section | Description |
|---------|-------------|
| `server` | HTTP/gRPC listen address, timeouts |
| `redis` | Connection pool, cluster mode |
| `kafka` | Brokers, audit topic |
| `engine` | Scene policies, fallback strategy |
| `rules` | Rule groups, hot-reload interval |
| `models` | ONNX model paths, champion-challenger config |
| `features` | Fetcher timeouts, velocity window sizes |

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
2. Register in `internal/feature/registry.go`

### Add an ML model

1. Export your model to ONNX format
2. Place the `.onnx` file in `configs/models/`
3. Register metadata in `configs/models/registry.yaml`

The engine hot-loads new/updated models without restart.

---

## Performance

Benchmarks run on a 8-core / 32 GB VM (Go 1.22, Redis 7 local):

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
├── cmd/               # Application entry points
├── internal/          # Private application code
│   ├── engine/        # Top-level decision orchestration
│   ├── rule/          # Rule DSL, evaluator, hot-reload
│   ├── feature/       # Parallel feature fetching
│   ├── model/         # ONNX inference, champion-challenger
│   ├── list/          # Tiered list service (Bloom+Redis+DB)
│   ├── orchestrator/  # DAG executor, A/B router
│   ├── audit/         # Kafka audit writer
│   └── config/        # Config loader
├── pkg/               # Public, reusable packages
│   ├── expr/          # Expression evaluator
│   ├── sliding/       # Redis sliding-window velocity
│   ├── bloom/         # In-process Bloom filter
│   └── pool/          # Goroutine pool for CGO isolation
├── api/               # HTTP handlers and protobuf definitions
├── configs/           # YAML config templates, rules, models
├── deployments/       # Docker and Kubernetes manifests
├── docs/              # Architecture docs, ADRs
└── scripts/           # Build, lint, benchmark scripts
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

- [antonmedv/expr](https://github.com/antonmedv/expr) – expression evaluator
- [bits-and-blooms/bloom](https://github.com/bits-and-blooms/bloom) – Bloom filter
- Design inspired by industry practices at Bilibili, ByteDance, and Meituan risk teams
