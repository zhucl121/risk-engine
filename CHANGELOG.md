# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Velocity Fetcher** (`internal/feature/fetchers`): sliding-window Redis counter for payment (1min/1h/24h) and promo (24h) velocity features; fail-open on Redis errors; `Increment()` API for transaction-commit path
- **Prometheus Metrics** (`internal/metrics`): four golden signals + business indicators — `decision_duration_seconds`, `decisions_total`, `rule_hits_total`, `feature_fetch_errors_total`, `active_requests`, `http_request_duration_seconds`, `rate_limited_total`, `circuit_breaker_state`; `/metrics` endpoint via `promhttp`
- **Two-tier Rate Limiter** (`internal/middleware.RateLimit`): global (5000 RPS) + per-IP (100 RPS) token-bucket with idle-entry cleanup; HTTP 429 + counter
- **Health Checks** (`internal/health`): `Checker` interface, `RedisChecker`, `FuncChecker`, `CompositeChecker`; `/api/v1/livez` (always 200) and `/api/v1/readyz` (503 on dependency failure) for Kubernetes probes
- **gRPC DecisionService** (`api/grpc/server`, `api/grpc/v1`): proto-generated stubs + `DecisionServer` implementing `Evaluate`, `BatchEvaluate`, `Health`; started alongside HTTP server with `GracefulStop`
- **A/B Test Traffic Splitting** (`internal/orchestrator`): `ABTestConfig.ExperimentPipeline`; `rand`-based bucket assignment; experiment group tagged with `abtest:{experimentID}` in `RiskReasons`; YAML `abTest` block parsed in registry loader
- **Circuit Breaker** (`internal/resilience`): `gobreaker`-backed `Breaker.Execute(ctx, fn)`; state changes recorded in `circuit_breaker_state` gauge; wired into `dispatchModel` and `dispatchList` in pipeline
- **Policy YAML Loading Fix** (`internal/orchestrator.LoadFromReader`): multi-document YAML (`---` separator) support; `loadPolicies()` in `main.go` delegates to `LoadFromYAML()` per file; `configs/policies/payment.yaml` example

### Changed
- `ABTestConfig` gained `ExperimentPipeline []Step` field for alternative step sequence
- `Handler.WithHealthChecker()` fluent option wires `CompositeChecker` to `/readyz`
- `orchestrator.Deps` gained `Breakers map[string]*resilience.Breaker` for per-step circuit breakers
- `go.mod` upgraded: `golang.org/x/time`, `google.golang.org/grpc`, added `github.com/sony/gobreaker`, `github.com/prometheus/client_golang`


- Self-hosted RiskDSL expression engine (`pkg/dsl`) built on ANTLR4-Go 4.13.2
  - Grammar file `pkg/dsl/grammar/RiskDSL.g4` (reproduced via `go generate`)
  - ANTLR4 ParseTree → custom AST → Go closure tree (zero-alloc hot path)
  - `pkg/dsl.Compile(condition, reg)` — compile-once, evaluate-many
  - `pkg/dsl.Program.Run(ctx, rt)` — P99 29ns (simple), 97ns (with features), 0 allocs
  - Built-in risk functions: `inList`, `velocity`, `modelScore`, `geoIP`, `within`
  - `FunctionRegistry` — plugin-style function extension without grammar changes
  - `Runtime` pool via `sync.Pool` to avoid per-request allocation
  - Backward-compatible with all existing `payment_rules.yaml` conditions (5/5 pass)
- Database-backed rule storage (`internal/rule`)
  - `RuleRepository` interface with `FakeRepository` (tests) and `MySQLRepository` (production)
  - DB migration: `configs/migrations/001_create_risk_rules.sql`
  - `DBLoader` — loads and compiles rules from DB; skips rules with invalid DSL
  - `DBWatcher` — polls DB every 30s for changes, triggers atomic hot-reload
  - `ForceReload` — called by management API for immediate reload
- Management REST API (`api/http/admin/v1/rules`)
  - 9 endpoints: List / Get / Create / Update / Delete / Validate / Enable / Disable / Reload
  - DSL validation on write — invalid conditions rejected before DB persist
  - Optimistic locking via `version` field prevents concurrent update conflicts
  - `condition_ast` JSON field supports visual rule builder round-trip

### Changed
- `pkg/dsl` replaces `pkg/expr` (antonmedv/expr) as the rule condition evaluator
  - `pkg/expr` retained for backward compatibility; will be archived after full migration
- `.golangci.yml` — excludes `pkg/dsl/parser/` (ANTLR4-generated code) from linter

### Added (dependencies)
- `github.com/antlr4-go/antlr/v4 v4.13.1`

- Initial project scaffold with full package structure
- Core interfaces: Engine, RuleEvaluator, FeatureFetcher, ModelScorer, ListService
- HTTP API v1: `/api/v1/decision`, `/api/v1/health`, `/api/v1/metrics`
- gRPC API: DecisionService proto definition
- Redis sliding-window velocity counter (`pkg/sliding`)
- In-process Bloom filter list cache (`pkg/bloom`)
- Goroutine pool for CGO isolation (`pkg/pool`)
- Expression evaluator wrapping antonmedv/expr (`pkg/expr`)
- Parallel feature fetching with per-fetcher timeout degradation
- Rule hot-reload via atomic double-buffer swap
- Tiered list service: Bloom (L1) → Redis (L2) → DB (L3)
- Champion-Challenger A/B routing for ML models
- DAG-based orchestration pipeline with YAML config
- Kafka audit writer with PII masking
- Prometheus metrics + OpenTelemetry tracing
- Docker Compose dev environment
- Kubernetes deployment manifests with HPA
- GitHub Actions CI: test, lint, benchmark, Docker build
- Cursor rules and agent skills for development acceleration
