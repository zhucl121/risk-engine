# Spec: Feature Service

**Status**: Active  
**Last updated**: 2026-02

## Overview

The Feature Service (`internal/feature`) coordinates parallel retrieval of all feature data needed for a decision. It applies per-fetcher timeouts and degrades gracefully on partial failure.

## Fetcher Interface Contract

```go
type Fetcher interface {
    Name()    string
    Timeout() time.Duration   // must come from config, never hardcoded
    Fetch(ctx context.Context, req *engine.DecisionRequest) (Map, error)
}
```

## Feature Map

`feature.Map` is `map[string]Value` where `Value` is a tagged union (`KindInt`, `KindFloat`, `KindString`, `KindBool`).  
**No `interface{}` in the hot path.** Always use typed accessors: `GetInt`, `GetFloat`, `GetString`, `GetBool`.

## Parallel Fetch Contract

1. All registered fetchers are invoked concurrently via `errgroup`.
2. Each fetcher gets its own `context.WithTimeout(ctx, f.Timeout())`.
3. A timeout in one fetcher does NOT cancel others.
4. On timeout: log warn + return zero-value features for that fetcher's keys — never block the decision.
5. Total feature-fetch budget controlled by `feature.total_timeout` config.

## Standard Feature Keys

All keys live in `internal/feature/keys.go`. Never use raw string literals in rules or fetchers.

| Prefix | Examples |
|--------|---------|
| `device.*` | `device.linked_account_count_7d`, `device.risk_score` |
| `user.*` | `user.register_days`, `user.credit_score` |
| `ip.*` | `ip.country`, `ip.is_datacenter` |
| `velocity.*` | `velocity.pay_count_1min`, `velocity.pay_count_24hour` |
| `session.*` | `session.duration_sec` |

## Velocity Features (Sliding Window)

Velocity features use `pkg/sliding.Window`:
- Redis Sorted Set per `(userID|deviceID):<window>`
- Lua script: atomic `ZREMRANGEBYSCORE + ZADD + EXPIRE + ZCOUNT` in one RTT
- Supported windows: 1min, 5min, 1hour, 24hour
- P99 target: < 5ms (local Redis)

## Performance Requirements

| Metric | Target |
|--------|--------|
| Feature fetch (all sources, parallel) | P99 < 20ms |
| Redis velocity check | P99 < 5ms |
| External API call | P99 < 30ms (with degradation on timeout) |
| Memory per feature map | < 10KB |
