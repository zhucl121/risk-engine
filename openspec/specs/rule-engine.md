# Spec: Rule Engine

**Status**: Active  
**Last updated**: 2026-02

## Overview

The rule engine evaluates a prioritised set of stateless rules against a feature map. Rules are defined in YAML, compiled to `expr.Program` objects, and hot-reloaded via an atomic pointer swap.

## Rule Interface Contract

```go
type Rule interface {
    ID()       string  // globally unique, e.g. "DEVICE_MULTI_ACCOUNT_001"
    Name()     string  // human-readable Chinese or English
    Priority() int     // higher = evaluated first
    Evaluate(ctx context.Context, rctx *Context) (*Result, error)
}
```

**Statelessness**: rules MUST NOT hold mutable state. All state lives in `feature.Map`.

## Rule Result Contract

| Field | Constraint |
|-------|-----------|
| `Hit` | true only when the condition is satisfied |
| `Decision` | REJECT \| MANUAL_REVIEW \| empty (for informational rules) |
| `RiskCode` | Machine-readable; UPPER_SNAKE_CASE; never free text |
| `Score` | Added to aggregate risk score only when `Hit == true` |

## Hot-Reload

- Rules stored in `configs/rules/<group>.yaml`
- File watcher polls every `engine.reload_interval`
- New rules compiled → syntax-checked → loaded into atomic.Pointer in one swap
- In-flight evaluations finish with old rule set

## Short-Circuit

When `shortCircuit: true` on the Evaluator (default for payment scenes):  
First rule returning `Decision == REJECT` stops all further rule evaluation.

## YAML Rule DSL

```yaml
- id: EXAMPLE_RULE_001
  name: "示例规则"
  priority: 100
  condition: "features['velocity.pay_count_1min'] > 5"
  action:
    decision: REJECT
    riskCode: HIGH_VELOCITY_PAYMENT
    score: 900
```

**Condition expression**: uses `pkg/expr` (antonmedv/expr). Available variables:
- `features` — `feature.Map` (use `features['key']`)
- `amount` — `int64` from `DecisionRequest.Amount`
- `userID`, `deviceID`, `ip` — strings from request

## Performance Requirements

| Metric | Target |
|--------|--------|
| Evaluate 100 rules | P99 < 10ms |
| Single rule eval (simple expr) | < 500ns |
| Hot-reload (file → live) | < 30s |
| Memory overhead per rule | < 50KB |
