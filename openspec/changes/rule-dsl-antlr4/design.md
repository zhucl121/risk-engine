# Design: 自研风控规则 DSL（ANTLR4-Go）

## Approach

新建 `pkg/dsl` 包，基于 ANTLR4-Go 生成 Lexer/Parser，将规则 condition 字符串在**加载阶段**编译为 Go 闭包树（`Program`）。请求热路径只执行 Go 闭包，零 ANTLR4 运行时开销。规则存储从 YAML 文件迁移到关系型数据库，新增管理 API 支持 CRUD 和 DSL 语法校验，前端可视化构建器通过 `condition_ast` JSON 字段与后端交换结构化规则，与 DSL 字符串双向转换。

## Component Changes

| Component | Change | Reason |
|-----------|--------|--------|
| `pkg/dsl/` | 新增（完整包） | 自研 DSL 引擎核心 |
| `internal/rule/loader.go` | 新增 | 从 DB 加载规则，替代 YAML loader |
| `internal/rule/repository.go` | 新增 | 规则 DB CRUD 接口及实现 |
| `internal/rule/evaluator.go` | 修改 | 使用 `dsl.Program` 替换 `expr.Program` |
| `api/http/admin/v1/rules.go` | 新增 | 管理 API handler |
| `api/http/admin/v1/dto.go` | 新增 | 请求/响应 DTO |
| `internal/config/config.go` | 修改 | 新增 DB DSN；热加载方式改为 DB poll |
| `pkg/expr/` | 不变（暂时保留） | 平滑迁移；待全量切换后归档 |

## Interface Delta

- `engine.Engine`: **no change**
- `rule.Rule`: **no change**
- `rule.Evaluator`: **no change**（内部实现替换 Program 类型）
- `pkg/expr.Program`: 保留，不删除
- 新增 `pkg/dsl` 包公开接口（见下方）

### 新增 pkg/dsl 公开接口

```go
// Env 描述规则求值时的变量环境（用于类型检查）
type Env struct {
    Features map[string]ValueKind  // feature key → 期望类型
    Vars     map[string]ValueKind  // amount, ip, userID 等请求字段
}

// Program 是编译后的规则条件，可并发安全地多次执行
type Program struct { /* 不透明 */ }

// Run 在给定环境下求值，返回 bool 结果
// 不调用任何 ANTLR4 代码；纯 Go 闭包执行
func (p *Program) Run(ctx context.Context, rt *Runtime) (bool, error)

// Source 返回原始 DSL 字符串
func (p *Program) Source() string

// Compile 将 DSL 字符串编译为 Program；仅在加载阶段调用
// 返回 SyntaxError 或 TypeError（均带行列信息）
func Compile(condition string, env Env, reg *FunctionRegistry) (*Program, error)

// FunctionRegistry 管理可在 DSL 中调用的函数
type FunctionRegistry struct { /* 不透明 */ }

func NewFunctionRegistry() *FunctionRegistry
func (r *FunctionRegistry) Register(fn FuncDef) error

// Runtime 持有运行时依赖（list.Service、sliding.Counter 等）
// 通过 context 或直接传参注入，不在 Program 中持有
type Runtime struct {
    Features feature.Map
    Request  *engine.DecisionRequest
    List     list.Service
    Velocity sliding.Counter
    Model    model.Registry
}
```

### 新增 internal/rule/repository.go 接口

```go
type RuleRepository interface {
    ListActive(ctx context.Context, sceneCode string) ([]*RuleRecord, error)
    ListUpdatedSince(ctx context.Context, since time.Time) ([]*RuleRecord, error)
    GetByID(ctx context.Context, id int64) (*RuleRecord, error)
    Create(ctx context.Context, r *RuleRecord) (int64, error)
    Update(ctx context.Context, r *RuleRecord) error   // 乐观锁：version check
    SoftDelete(ctx context.Context, id int64) error
}
```

## Data Flow

```
规则加载阶段（每 30s 或手动触发）:
  DB.ListUpdatedSince(lastCheck)
    → for each RuleRecord:
        dsl.Compile(record.Condition, env, funcReg)
        → ANTLR4 Parse → AST → TypeCheck → Program(Go闭包)
    → atomic.Pointer.Store(newRuleSet)

请求热路径:
  ruleSet := atomicPtr.Load()
  for _, rule := range ruleSet:
    result, _ = rule.program.Run(ctx, runtime)
    // 纯 Go 闭包，零 ANTLR4 调用
```

## DSL Grammar 核心结构（RiskDSL.g4 草案）

```antlr
grammar RiskDSL;

// 顶层：一条规则条件是一个逻辑表达式
condition   : expr EOF ;

expr
    : expr ('&&' | '||') expr          # LogicalExpr
    | '!' expr                          # NotExpr
    | expr ('>' | '<' | '>=' | '<=' | '==' | '!=') expr  # CmpExpr
    | ID '(' argList? ')'              # CallExpr
    | ID '[' STRING ']'                # MapIndexExpr   // features['key']
    | ID ('.' ID)+                     # FieldAccessExpr
    | '(' expr ')'                     # ParenExpr
    | literal                          # LiteralExpr
    | ID                               # IdentExpr
    ;

argList : expr (',' expr)* ;

literal
    : INT    # IntLit
    | FLOAT  # FloatLit
    | STRING # StringLit
    | BOOL   # BoolLit
    ;

// Tokens
INT    : [0-9]+ ;
FLOAT  : [0-9]+ '.' [0-9]+ ;
STRING : '\'' (~'\'')* '\'' | '"' (~'"')* '"' ;
BOOL   : 'true' | 'false' ;
ID     : [a-zA-Z_][a-zA-Z0-9_]* ;
WS     : [ \t\r\n]+ -> skip ;
```

## AST 节点设计

```go
// pkg/dsl/ast/nodes.go
type Node interface{ nodeType() string }

type BinaryExpr struct { Op string; Left, Right Node }
type UnaryExpr  struct { Op string; Operand Node }
type CallExpr   struct { Name string; Args []Node }
type MapIndex   struct { Map string; Key string }       // features['key']
type FieldAccess struct { Obj string; Fields []string } // geoIP(ip).country
type Ident      struct { Name string }
type IntLit     struct { Val int64 }
type FloatLit   struct { Val float64 }
type StringLit  struct { Val string }
type BoolLit    struct { Val bool }
```

## Program（闭包树）实现

```go
// pkg/dsl/program.go
type evalFn func(ctx context.Context, rt *Runtime) (Value, error)

type Program struct {
    src  string
    eval evalFn  // 根节点闭包，递归调用子闭包
}

// Compiler 将 AST 节点递归编译为闭包
func compileNode(node ast.Node, reg *FunctionRegistry) (evalFn, error) {
    switch n := node.(type) {
    case *ast.BinaryExpr:
        left, _ := compileNode(n.Left, reg)
        right, _ := compileNode(n.Right, reg)
        return func(ctx context.Context, rt *Runtime) (Value, error) {
            l, _ := left(ctx, rt)
            r, _ := right(ctx, rt)
            return applyOp(n.Op, l, r)
        }, nil
    // ... 其他节点类型
    }
}
```

## 数据库 Schema

```sql
CREATE TABLE risk_rules (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    rule_key        VARCHAR(128) NOT NULL UNIQUE,
    name            VARCHAR(256) NOT NULL,
    group_name      VARCHAR(128) NOT NULL,
    scene_code      VARCHAR(128) NOT NULL,
    priority        INT NOT NULL DEFAULT 100,
    condition_dsl   TEXT NOT NULL,        -- DSL 字符串
    condition_ast   JSON,                 -- 可视化构建器 JSON（可空）
    action_decision VARCHAR(32) NOT NULL,
    action_risk_code VARCHAR(128),
    action_score    INT DEFAULT 0,
    status          TINYINT NOT NULL DEFAULT 1,
    version         INT NOT NULL DEFAULT 1,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_scene_status (scene_code, status),
    INDEX idx_updated_at   (updated_at)
);
```

## Alternatives Considered

### Option A（chosen）: ANTLR4-Go 生成 Parser + Go 闭包 Program
语法文件即文档，扩展语法只需改 `.g4`；Parser 仅在 load time 调用，热路径零开销。

### Option B（rejected）: 手写递归下降 Parser
优点：零外部依赖，无 ANTLR4 版本绑定。缺点：Grammar 变更时需同步改 Parser 代码，维护成本高；调试工具（parse tree 可视化）需自建。

### Option C（rejected）: 保留 antonmedv/expr + 函数注入
优点：零迁移成本。缺点：语法固定不可扩展，无 AST 访问权限，无法构建可视化构建器，无法支持 velocity/inList 等一等公民语法。

## ADR

### ADR-001: ANTLR4 仅用于 load time，不用于 request time
**Context**: ANTLR4-Go 运行时在并发场景下有 DFA 缓存竞争问题（see GitHub issue #4344）。  
**Decision**: Compiler.Compile() 将 ParseTree 转换为纯 Go 闭包（Program）。请求热路径只调用 Program.Run()，不接触 ANTLR4 代码。  
**Consequences**: 热路径无 ANTLR4 开销；load time（热加载，每 30s 一次）可接受 ANTLR4 的解析延迟（< 100ms/100 rules）。

### ADR-002: Repository 接口化，测试用 fake
**Context**: DB 依赖会使单元测试复杂。  
**Decision**: `RuleRepository` 定义为接口，生产实现用 `database/sql`，测试用 `fakeRepository`（in-memory map）。  
**Consequences**: 所有 rule 层单测无需真实 DB；集成测试在 CI 用 Docker MySQL。

### ADR-003: condition_ast 字段可空
**Context**: 工程师手写 DSL 时无可视化 JSON；分析师使用构建器时会有。  
**Decision**: `condition_ast` 允许为 NULL；前端加载时若有则渲染构建器，若无则以文本模式显示 DSL。  
**Consequences**: 向后兼容；不强制所有规则必须有可视化表示。

## Benchmark Targets

| Path | Target |
|------|--------|
| `BenchmarkRunProgram` (单条简单条件) | < 500ns/op |
| `BenchmarkRunProgram` (含函数调用) | < 2μs/op |
| `BenchmarkEvaluate100Rules` | < 10ms/op |
| `BenchmarkCompile` (单条，load time) | < 5ms/op |

## REVIEW (Architect)
Date: 2026-02-23

### Open Questions

- [x] Q1: ANTLR4 Go 并发解析缓存 bug 是否已修复？  
  → 已确认：加载阶段为单线程串行（atomic swap 前），不走并发路径，问题不复现。ADR-001 记录此决策。
- [x] Q2: `Runtime` 结构体持有 `list.Service` / `model.Registry` 等接口引用，每次 `Run()` 都要传入，会不会造成热路径分配？  
  → `Runtime` 应由调用方通过 `sync.Pool` 复用，设计文档需在 TASK-6 中注明 Pool 策略。已在 tasks.md TASK-6 中标注。
- [x] Q3: 管理 API 是否需要鉴权？  
  → 在 TASK-29 中以可配置 token/basicAuth 实现，不硬编码。已在 tasks.md 标注。
- [ ] Q4: `condition_ast` JSON Schema 是否需要在本 change 中定义？  
  → 当前 change 后端只存 opaque JSON blob，前端负责解释。Schema 定义推迟到前端 Phase 9，不阻塞本 change。

### Risks

- Risk: Grammar 在 Phase 1 设计不完备，后续需破坏性修改 `.g4` | Mitigation: Phase 1 先以现有 `payment_rules.yaml` 中所有 condition 写 TASK-33 兼容性测试驱动 Grammar 设计，确保覆盖所有当前语法
- Risk: `database/sql` + 无连接池配置导致 DB 连接泄漏 | Mitigation: TASK-24 config 中必须包含 `max_open_conns` / `max_idle_conns` / `conn_max_lifetime`，loader_test 验证连接关闭
- Risk: 热加载 DB 轮询在规则数量大时全量重编译开销高 | Mitigation: TASK-30 `DBWatcher` 只拉取 `updated_at > lastCheck` 的增量记录，不全量扫描；TASK-34 测试增量热加载路径
- Risk: `geoIP` 函数需要外部 GeoIP 库或数据库，可能引入新依赖 | Mitigation: TASK-17 用接口隔离，默认实现可以是 stub，生产实现通过注入；不在本 change 内强制引入第三方 GeoIP 库

### Interface Impact

- `engine.Engine`: no change
- `rule.Rule`: no change
- `rule.Evaluator`: no change（签名不变，内部 Program 类型替换为 `dsl.Program`）
- `pkg/expr.Program`: no change（保留，不删除）
- `pkg/dsl`: 全新包，无存量接口变更风险
- `internal/config`: 新增字段（向后兼容，旧配置文件警告但不崩溃）

### Performance Budget

- `Program.Run()` P99: < 500ns（Go 闭包调用，等价于 antonmedv/expr 现有水平）
- 热加载编译 100 条规则: < 500ms（每条 < 5ms，ANTLR4 load time 可接受）
- 管理 API 写入: < 200ms P99（含 DSL Compile 校验 + DB 写入）
- 以上全部需 Benchmark 在 CI 中记录基线，合并时报告数字

### Decision

[x] GO — 接口边界清晰，热路径不引入 ANTLR4，性能风险可控，可进入 `/opsx:apply`
