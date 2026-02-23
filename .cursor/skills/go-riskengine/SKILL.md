---
name: go-riskengine
description: Expert skill for the riskengine open-source Go project. Use when implementing, reviewing, or extending any component of the risk decision engine: rule engine, model engine, feature service, orchestrator, list service, audit, API layer, or deployment config. Provides concise domain knowledge, interface contracts, and implementation patterns to minimize context token usage.
---

# RiskEngine Go Project Skill

Full design reference: `风控决策引擎深度设计方案.md`
Domain rules: `.cursor/rules/riskengine-domain.mdc`

## Quick Reference

### Project Layout
```
riskengine/
├── cmd/server/          # main.go – wire & serve
├── cmd/migrate/         # DB/config migration tool
├── internal/
│   ├── engine/          # DecisionEngine – top-level orchestration
│   ├── rule/            # Rule DSL parser, evaluator, hot-reload
│   ├── feature/         # Parallel feature fetching, velocity windows
│   ├── model/           # ONNX inference pool, champion-challenger
│   ├── list/            # Bloom+Redis+DB tiered list service
│   ├── orchestrator/    # DAG executor, A/B router, aggregator
│   ├── audit/           # Kafka audit writer, PII masking
│   ├── config/          # Config loader (Viper + Nacos/Etcd watch)
│   └── middleware/      # Rate limit, auth, trace injection
├── pkg/
│   ├── expr/            # Expression evaluator (wraps antonmedv/expr)
│   ├── sliding/         # Redis sliding-window velocity counter
│   ├── bloom/           # In-process Bloom filter
│   └── pool/            # Generic goroutine pool (ONNX CGO isolation)
├── api/
│   ├── http/v1/         # Gin handlers + OpenAPI spec
│   └── grpc/proto/      # Protobuf definitions
├── configs/             # YAML config templates
├── deployments/         # Docker, Kubernetes manifests
├── scripts/             # lint, test, benchmark scripts
└── docs/                # Architecture diagrams, ADRs
```

### Core Interfaces (never change signatures without major version bump)
See [reference.md](reference.md) for full interface definitions.

Key contracts:
- `engine.Engine` – `Evaluate(ctx, *DecisionRequest) (*DecisionResult, error)`
- `rule.Evaluator` – `Evaluate(ctx, *rule.Context) ([]*rule.Result, error)`
- `feature.Fetcher` – `Fetch(ctx, *feature.Key) (feature.Map, error)`
- `model.Scorer` – `Score(ctx, feature.Map) (map[string]float64, error)`
- `list.Service` – `Check(ctx, *list.Query) (list.Status, error)`

### Decision Flow (implement in this order when adding a new scene)
1. Add scene code to `api/http/v1/scenes.go`
2. Define `PolicySet` YAML in `configs/policies/<scene>.yaml`
3. Register feature fetchers needed for the scene
4. Add/reuse rules in `configs/rules/`
5. Wire in `internal/orchestrator/registry.go`
6. Add integration test in `internal/engine/testdata/<scene>/`

### Performance Targets
| Component | P99 Target |
|-----------|-----------|
| List check (L1 Bloom) | < 1ms |
| Feature fetch (parallel) | < 20ms |
| Rule evaluation (100 rules) | < 10ms |
| Model scoring (XGBoost) | < 15ms |
| End-to-end decision | < 60ms |
| Throughput (8-core) | > 20,000 TPS |

### Key Dependencies
```
go.uber.org/zap          # logging
github.com/gin-gonic/gin # HTTP
google.golang.org/grpc   # gRPC
github.com/redis/go-redis/v9  # Redis
golang.org/x/sync        # errgroup
github.com/antonmedv/expr     # expression engine
github.com/prometheus/client_golang  # metrics
go.opentelemetry.io/otel  # tracing
```

## Additional Resources
- Full interface contracts: [reference.md](reference.md)
- Implementation examples: [examples.md](examples.md)
- Feature implementation steps & `/opsx:` commands: skill `riskengine-workflow`
