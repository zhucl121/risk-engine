# Architecture

## Overview

RiskEngine is a stateless, horizontally scalable service. Each instance holds in-memory state (rule programs, Bloom filter, model weights) that is periodically refreshed from configuration sources. All durable state lives in Redis, Kafka, and the database.

## Component Map

```
┌──────────────────────────────────────────────────────────────────┐
│                       API Layer                                   │
│            Gin HTTP /api/v1   │   gRPC :9090                     │
└────────────────────────┬─────────────────────────────────────────┘
                         │
┌────────────────────────▼─────────────────────────────────────────┐
│                   DecisionEngine (internal/engine)                │
│  1. Validate request  2. Route to PolicySet  3. Execute Pipeline  │
└───────┬──────────────────────┬──────────────────────┬────────────┘
        │                      │                      │
┌───────▼──────┐  ┌────────────▼───────┐  ┌──────────▼──────────┐
│  ListService  │  │   RuleEvaluator    │  │   ModelRegistry     │
│  (list/)      │  │   (rule/)          │  │   (model/)          │
│               │  │                   │  │                      │
│  L1: Bloom    │  │  atomic.Pointer    │  │  ONNX Runtime pool  │
│  L2: Redis    │  │  []Rule (sorted)   │  │  champion/challenger│
│  L3: DB       │  │  expr.Program      │  │                     │
└───────┬───────┘  └────────────┬───────┘  └──────────┬──────────┘
        │                       │                      │
        └───────────────────────┼──────────────────────┘
                                │
┌───────────────────────────────▼──────────────────────────────────┐
│                   FeatureService (feature/)                       │
│  errgroup parallel fetch · per-fetcher timeout · degradation      │
│                                                                   │
│  RedisVelocityFetcher  │  UserProfileFetcher  │  DeviceFetcher    │
└───────────────────────────────┬──────────────────────────────────┘
                                │
        ┌───────────────────────┼─────────────────────┐
        │                       │                      │
    Redis Cluster           HBase / DB            External APIs
  (velocity, session)      (user profile)       (IP intel, device)
                                │
┌───────────────────────────────▼──────────────────────────────────┐
│                   AuditWriter (audit/)                            │
│  async batch → Kafka topic "risk.decisions" (at-least-once)      │
└──────────────────────────────────────────────────────────────────┘
```

## Data Flow

1. **Request arrives** at the API layer (HTTP POST or gRPC Evaluate).
2. **DecisionEngine** validates the request, assigns a RequestID if absent, and looks up the PolicySet for the requested SceneCode.
3. **FeatureService** fires all registered Fetchers in parallel, each with its own deadline. Timeouts result in zero-value features (degraded mode).
4. **Orchestrator** executes the PolicySet's DAG:
   - **ListService** (L1 Bloom → L2 Redis → L3 DB): fast-path rejection for known bad entities.
   - **RuleEvaluator**: atomically reads the current rule set and evaluates rules by priority; short-circuits on REJECT.
   - **ModelRegistry**: routes to champion (or challenger) scorer; runs ONNX inference in a CGO-isolated goroutine pool.
5. **Aggregator** merges results according to the configured AggregationStrategy (HIGHEST_RISK | WEIGHTED | RULE_FIRST).
6. **DecisionResult** is serialised and returned to the caller.
7. **AuditWriter** enqueues a masked Record to Kafka for downstream analytics and model feedback loops.

## Hot Reload

Rules and policies are watched for changes every `engine.reload_interval` (default 30s):

```
File watcher detects change
       ↓
Load and compile new rules/policies
       ↓
Validate (syntax check, expression compile)
       ↓
Atomic swap (atomic.Pointer.Store)
       ↓
In-flight requests continue with old version
New requests use new version immediately
```

## Concurrency Model

- One goroutine per incoming HTTP/gRPC request; no goroutine pool at the API layer.
- Feature fetching uses `errgroup` for structured concurrency with context propagation.
- Rule evaluation is fully in-memory; no goroutine creation per rule.
- Model inference runs in a fixed-size goroutine pool (`pkg/pool`) to isolate CGO thread locking.
- AuditWriter uses a background goroutine with a buffered channel; callers never block on Kafka I/O.

## Failure Modes

| Component | Failure | Behaviour |
|-----------|---------|-----------|
| Redis (features) | Timeout | Zero-value features; log warn; continue |
| Redis (list L2) | Timeout | Fall through to L3 (DB) |
| DB (list L3) | Timeout | Return StatusNotFound; log error |
| Rule evaluator | Panic | Recovered; return ErrInternal |
| Model scorer | Timeout | Skip model score; aggregate without it |
| Kafka audit | Full queue | Log error; drop record (audit is non-blocking) |

## Performance Design Decisions

See `.cursor/rules/performance.mdc` for coding-level patterns.

Key architectural decisions:
- **Object pools** (`sync.Pool`) for `DecisionContext` and `FeatureMap` to reduce GC pressure.
- **Bloom filter** as L1 list cache: eliminates Redis RTT for ~99.9% of clean traffic.
- **Atomic pointer swap** for rule hot-reload: zero lock contention during normal evaluation.
- **CGO pool** for ONNX: prevents unbounded OS thread creation that would stall the Go scheduler.
- **GOGC=200 + GOMEMLIMIT**: traded memory for lower GC frequency; tunable per deployment.
