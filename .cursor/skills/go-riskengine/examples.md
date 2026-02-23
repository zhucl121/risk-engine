# RiskEngine – Implementation Examples

## 1. Adding a New Rule

**File:** `internal/rule/rules/device_multi_account.go`

```go
package rules

import (
    "context"
    "github.com/zhucl121/risk-engine/internal/engine"
    "github.com/zhucl121/risk-engine/internal/feature"
    "github.com/zhucl121/risk-engine/internal/rule"
)

// DeviceMultiAccountRule rejects when a device is linked to too many accounts.
type DeviceMultiAccountRule struct {
    MaxAccounts int
    WindowDays  int
}

func (r *DeviceMultiAccountRule) ID()       string { return "DEVICE_MULTI_ACCOUNT_001" }
func (r *DeviceMultiAccountRule) Name()     string { return "设备关联多账号检测" }
func (r *DeviceMultiAccountRule) Priority() int    { return 100 }

func (r *DeviceMultiAccountRule) Evaluate(ctx context.Context, rctx *rule.Context) (*rule.Result, error) {
    count := rctx.Features.GetInt(feature.KeyDeviceLinkedAccounts7d)
    if count <= int64(r.MaxAccounts) {
        return &rule.Result{RuleID: r.ID(), Hit: false}, nil
    }
    return &rule.Result{
        RuleID:   r.ID(),
        Hit:      true,
        Score:    850,
        Decision: engine.DecisionReject,
        RiskCode: "DEVICE_MULTI_ACCOUNT",
    }, nil
}
```

**Register in** `internal/rule/registry.go`:
```go
func DefaultRules(cfg *config.RuleConfig) []Rule {
    return []Rule{
        &rules.DeviceMultiAccountRule{
            MaxAccounts: cfg.DeviceMultiAccount.MaxAccounts,
            WindowDays:  cfg.DeviceMultiAccount.WindowDays,
        },
        // ... other rules
    }
}
```

---

## 2. Parallel Feature Fetch Pattern

**File:** `internal/feature/service.go`

```go
func (s *service) Fetch(ctx context.Context, req *engine.DecisionRequest) (feature.Map, error) {
    result := make(feature.Map, 64)
    var mu sync.Mutex

    g, gctx := errgroup.WithContext(ctx)
    for _, f := range s.fetchers {
        f := f  // capture
        g.Go(func() error {
            fCtx, cancel := context.WithTimeout(gctx, f.Timeout())
            defer cancel()

            m, err := f.Fetch(fCtx, req)
            if err != nil {
                // degraded: log and continue; never block decision
                s.logger.Warn("feature fetch timeout", zap.String("fetcher", f.Name()), zap.Error(err))
                s.metrics.FetchError(f.Name())
                return nil  // non-fatal
            }
            mu.Lock()
            for k, v := range m {
                result[k] = v
            }
            mu.Unlock()
            return nil
        })
    }
    _ = g.Wait()
    return result, nil
}
```

---

## 3. Rule Hot-Reload (Double Buffer)

```go
// internal/rule/evaluator.go
type atomicEvaluator struct {
    rules atomic.Pointer[[]Rule]
}

func (e *atomicEvaluator) Reload(rules []Rule) error {
    sorted := sortByPriority(rules)
    e.rules.Store(&sorted)
    return nil
}

func (e *atomicEvaluator) Evaluate(ctx context.Context, rctx *Context) ([]*Result, error) {
    rules := *e.rules.Load()
    results := make([]*Result, 0, len(rules))
    for _, r := range rules {
        res, err := r.Evaluate(ctx, rctx)
        if err != nil {
            return nil, fmt.Errorf("rule %s: %w", r.ID(), err)
        }
        if res.Hit && res.Decision == engine.DecisionReject {
            return []*Result{res}, nil  // short-circuit on hard reject
        }
        results = append(results, res)
    }
    return results, nil
}
```

---

## 4. Redis Sliding Window (Velocity)

**File:** `pkg/sliding/window.go`

```go
// Count atomically increments and returns the count within [now-window, now].
// Uses Redis ZREMRANGEBYSCORE + ZADD + ZCARD in a Lua script for atomicity.
func (w *RedisWindow) Count(ctx context.Context, key string, window time.Duration) (int64, error) {
    now := time.Now().UnixMilli()
    min := now - window.Milliseconds()
    // luaScript: ZREMRANGEBYSCORE, ZADD, EXPIRE, ZCARD in one call
    result, err := w.rdb.EvalSha(ctx, w.sha, []string{key},
        min, now, now, int(window.Seconds())+1).Int64()
    if err != nil {
        return 0, fmt.Errorf("sliding window count: %w", err)
    }
    return result, nil
}
```

---

## 5. Tiered List Check

```go
// internal/list/service.go
func (s *svc) Check(ctx context.Context, q *list.Query) (list.Status, error) {
    key := q.Kind + ":" + q.Value

    // L1: Bloom filter (in-process, ~1µs)
    if !s.bloom.Test([]byte(key)) {
        return list.StatusNotFound, nil  // definitely not in any list
    }

    // L2: Redis (~2ms)
    if st, err := s.redis.Get(ctx, key); err == nil {
        return st, nil
    }

    // L3: DB exact match (~10ms)
    return s.db.Lookup(ctx, q)
}
```

---

## 6. Config-Driven PolicySet

> See full example: `configs/policies/payment_checkout.yaml`

Key structure:
```yaml
scene: PAYMENT_CHECKOUT
version: "1.0.0"
pipeline:
  - name: list_check   # LIST | RULE | MODEL | AGGREGATE
    kind: LIST
    timeout: 5ms
    onFailure: SKIP    # SKIP = degraded pass; FAIL = hard fail
  - name: rule_engine
    kind: RULE
    ruleGroup: payment_rules
    timeout: 25ms
  - name: model_scoring
    kind: MODEL
    models: [payment_fraud_xgb, device_risk_lr]
    timeout: 30ms
    parallel: true     # concurrent with rule_engine
  - name: aggregate
    kind: AGGREGATE
    strategy: HIGHEST_RISK
fallback: PASS_WITH_MONITOR
```
