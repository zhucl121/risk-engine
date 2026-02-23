# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
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
