# RiskEngine – Interface Reference

## engine package

```go
// Engine is the top-level decision orchestrator.
type Engine interface {
    Evaluate(ctx context.Context, req *DecisionRequest) (*DecisionResult, error)
    Reload(ctx context.Context) error  // hot-reload policies/rules/models
    Health() HealthStatus
}

type DecisionRequest struct {
    RequestID   string            // UUID v4, set by caller or generated
    SceneCode   string            // e.g. "PAYMENT_CHECKOUT"
    UserID      string
    DeviceID    string
    SessionID   string
    IP          string
    Amount      int64             // in cents; 0 for non-monetary scenes
    Extra       map[string]string // scene-specific fields
    ReceivedAt  time.Time
}

type DecisionResult struct {
    RequestID   string
    Decision    Decision          // PASS | REJECT | MANUAL_REVIEW
    RiskScore   int               // 0–1000
    RiskLevel   RiskLevel         // LOW | MEDIUM | HIGH | CRITICAL
    HitRules    []string          // rule IDs that fired
    ModelScores map[string]float64
    RiskReasons []string          // reason codes, e.g. "DEVICE_MULTI_ACCOUNT"
    Actions     []string          // recommended actions, e.g. "REQUIRE_OTP"
    Path        []StepTrace       // full execution trace for audit
    CostMs      int64
}

type Decision  string
const (
    DecisionPass         Decision = "PASS"
    DecisionReject       Decision = "REJECT"
    DecisionManualReview Decision = "MANUAL_REVIEW"
)
```

## rule package

```go
type Rule interface {
    ID()       string
    Name()     string
    Priority() int
    Evaluate(ctx context.Context, rctx *Context) (*Result, error)
}

type Context struct {
    Request  *engine.DecisionRequest
    Features feature.Map
}

type Result struct {
    RuleID    string
    Hit       bool
    Score     int
    Decision  engine.Decision
    RiskCode  string
    CostMs    int64
}

// Evaluator runs a set of rules and returns all results.
type Evaluator interface {
    Evaluate(ctx context.Context, rctx *Context) ([]*Result, error)
    Reload(rules []Rule) error  // atomic hot-swap
}
```

## feature package

```go
// Map is the unified feature container passed through the pipeline.
type Map map[string]Value

// Value is a tagged union to avoid interface{} boxing in hot paths.
type Value struct {
    Kind    ValueKind
    IntVal  int64
    FltVal  float64
    StrVal  string
    BoolVal bool
}

type ValueKind uint8
const (
    KindInt ValueKind = iota
    KindFloat
    KindString
    KindBool
)

// Fetcher retrieves features for a given request.
type Fetcher interface {
    Name()    string
    Timeout() time.Duration
    Fetch(ctx context.Context, req *engine.DecisionRequest) (Map, error)
}

// Service orchestrates parallel fetching across all registered Fetchers.
type Service interface {
    Fetch(ctx context.Context, req *engine.DecisionRequest) (Map, error)
    Register(f Fetcher)
}
```

## model package

```go
// Scorer runs ML inference and returns named scores.
type Scorer interface {
    Name()    string
    Score(ctx context.Context, features feature.Map) (float64, error)
}

// Registry manages champion-challenger routing across multiple scorers.
type Registry interface {
    Score(ctx context.Context, modelName string, features feature.Map) (float64, error)
    Reload(name string, path string) error  // hot-swap ONNX file
    Champion(name string) Scorer
    SetChallenger(name string, s Scorer, trafficPct float64)
}
```

## list package

```go
type Status int
const (
    StatusNotFound  Status = iota
    StatusBlacklist
    StatusWhitelist
    StatusGraylist
)

type Query struct {
    Kind  string // "user" | "device" | "ip" | "card"
    Value string
}

type Service interface {
    Check(ctx context.Context, q *Query) (Status, error)
    Add(ctx context.Context, q *Query, status Status, ttl time.Duration) error
    Remove(ctx context.Context, q *Query) error
}
```

## orchestrator package

```go
// Pipeline executes a DAG of steps and aggregates results.
type Pipeline interface {
    Execute(ctx context.Context, req *engine.DecisionRequest) (*engine.DecisionResult, error)
}

// Step is one node in the execution DAG.
type Step struct {
    Name      string
    Kind      StepKind          // LIST | RULE | MODEL | GRAPH | CUSTOM
    Timeout   time.Duration
    Parallel  bool              // run concurrently with siblings
    OnFailure FailurePolicy     // SKIP | REJECT | FALLBACK
}

// AggregationStrategy merges sub-results into final DecisionResult.
type AggregationStrategy interface {
    Aggregate(results []*StepResult) *engine.DecisionResult
}
```

## audit package

```go
type Record struct {
    RequestID   string
    SceneCode   string
    UserID      string            // masked: last 4 chars visible
    DeviceID    string
    Decision    engine.Decision
    RiskScore   int
    Features    feature.Map       // PII fields masked before write
    RuleResults []rule.Result
    ModelScores map[string]float64
    CostMs      int64
    CreatedAt   time.Time
}

type Writer interface {
    Write(ctx context.Context, r *Record) error
    Flush(ctx context.Context) error
}
```
