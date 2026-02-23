# Spec: Decision Engine

**Status**: Active  
**Last updated**: 2026-02

## Overview

The Decision Engine (`internal/engine`) is the top-level orchestrator. It accepts a `DecisionRequest`, coordinates all sub-systems, and returns a `DecisionResult` within the latency budget.

## Core Invariants (never break)

1. `Engine.Evaluate()` is always safe for concurrent use.
2. A decision is always returned — even on partial sub-system failure (degraded mode).
3. `DecisionResult.RequestID` is always set (UUID v4, generated if not provided by caller).
4. `DecisionResult.CostMs` reflects actual wall-clock decision time.
5. `DecisionResult.Decision` is always one of: `PASS`, `REJECT`, `MANUAL_REVIEW`.

## Interface Contract

```go
type Engine interface {
    Evaluate(ctx context.Context, req *DecisionRequest) (*DecisionResult, error)
    Reload(ctx context.Context) error
    Health() HealthStatus
}
```

`Evaluate` returns an error ONLY for infrastructure failures (e.g. context already cancelled on entry).  
Business risk decisions are always encoded in `DecisionResult`, never in the error.

## Decision Result Fields

| Field | Type | Constraint |
|-------|------|-----------|
| `RequestID` | string | UUID v4; never empty |
| `Decision` | enum | PASS \| REJECT \| MANUAL_REVIEW |
| `RiskScore` | int | 0–1000 inclusive |
| `RiskLevel` | enum | LOW \| MEDIUM \| HIGH \| CRITICAL |
| `HitRules` | []string | May be empty; never nil |
| `ModelScores` | map[string]float64 | May be empty; never nil |
| `RiskReasons` | []string | Machine-readable codes only |
| `CostMs` | int64 | >= 0 |

## Performance Requirements

| Metric | Target |
|--------|--------|
| P99 end-to-end | < 60ms |
| P99 end-to-end (list-only) | < 5ms |
| Throughput (8-core) | ≥ 20,000 TPS |
| Cold-start | < 2s (binary ready to serve) |

## Degraded Mode Behaviour

| Failure | Behaviour |
|---------|-----------|
| FeatureService timeout | Use zero-value features; continue |
| RuleEvaluator panic | Recover; return `ErrInternal`; log stack |
| ModelRegistry unavailable | Skip model scores; aggregate without them |
| ListService unavailable | Skip list check (log warn); apply SKIP policy |
| All sub-systems fail | Return scene-level `fallback` decision |

## Hot-Reload Requirement

`Engine.Reload()` must propagate new rules/policies within `engine.reload_interval` (default 30s).  
In-flight requests complete with the old configuration; new requests use the new configuration immediately after the atomic swap.
