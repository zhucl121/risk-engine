# Tasks: 自研风控规则 DSL（ANTLR4-Go）
<!-- Implementor: 一次只做一个任务。mark [x] 前必须 make test 通过。 -->
<!-- Commit 格式: feat(dsl): TASK-N description -->

## Phase 1: Grammar & 生成基础设施

- [x] TASK-1: 新增 `pkg/dsl/grammar/RiskDSL.g4` —— 编写完整 ANTLR4 语法文件（覆盖 FR-1 所有语法结构）
- [x] TASK-2: 新增 `pkg/dsl/grammar/generate.go` —— 配置 `//go:generate` 指令生成 parser/lexer/visitor 代码
- [x] TASK-3: 新增 `pkg/dsl/parser/` —— 执行 `go generate ./pkg/dsl/grammar/`，提交生成代码（ANTLR4 4.13.2 生成）
- [x] TASK-4: 新增 `pkg/dsl/ast/nodes.go` —— 定义全部 AST 节点类型（BinaryExpr / UnaryExpr / CallExpr / MapIndex / FieldAccess / Ident / 字面量）

## Phase 2: Compiler（ParseTree → Program）

- [x] TASK-5: 新增 `pkg/dsl/value.go` —— 定义 `Value` 联合类型（int64 / float64 / string / bool / GeoInfo），避免 interface{} 装箱
- [x] TASK-6: 新增 `pkg/dsl/runtime.go` —— 定义 `Runtime` 结构体（持有 feature.Map + 各服务依赖）；sync.Pool 复用
- [x] TASK-7: 新增 `pkg/dsl/functions.go` —— 定义 `FuncDef` + `FunctionRegistry`，实现 `Register` / `Lookup`
- [x] TASK-8: 新增 `pkg/dsl/visitor.go` —— 实现 ANTLR4 Visitor，将 ParseTree 转换为自定义 AST
- [x] TASK-9: 新增 `pkg/dsl/compiler.go` —— 实现 `Compile(condition, reg)` 入口：调用 Parser → Visitor → codegen
- [x] TASK-10: 新增 `pkg/dsl/codegen.go` —— 实现 AST → Go 闭包树（`evalFn`）递归编译
- [x] TASK-11: 新增 `pkg/dsl/program.go` —— 实现 `Program` 结构体和 `Run(ctx, rt)` 方法

## Phase 3: TypeChecker

- [ ] TASK-12: 新增 `pkg/dsl/typechecker.go` —— 实现静态类型推断（遍历 AST，检查操作数类型兼容性，函数签名匹配）
- [x] TASK-13: 新增 `pkg/dsl/errors.go` —— 定义 `SyntaxError` / `TypeError`（含 Line / Col / Message 字段）

## Phase 4: 内置风控函数实现

- [x] TASK-14: 新增 `pkg/dsl/builtins/inlist.go` —— 实现 `inList(listName, value)` 函数（调用 list.Service.Check）
- [x] TASK-15: 新增 `pkg/dsl/builtins/velocity.go` —— 实现 `velocity(event, window)` 函数（调用 sliding.Counter）
- [x] TASK-16: 新增 `pkg/dsl/builtins/modelscore.go` —— 实现 `modelScore(modelName)` 函数（调用 model.Registry.Score）
- [x] TASK-17: 新增 `pkg/dsl/builtins/geoip.go` —— 实现 `geoIP(ip)` 函数（返回 GeoInfo 对象，`.country` `.isp` 字段访问）
- [x] TASK-18: 新增 `pkg/dsl/builtins/within.go` —— 实现 `within(v, lo, hi)` 范围判断函数
- [x] TASK-19: 新增 `pkg/dsl/builtins/register.go` —— 统一注册所有内置函数到默认 FunctionRegistry

## Phase 5: 数据库存储层

- [x] TASK-20: 新增 `internal/rule/record.go` —— 定义 `RuleRecord` 结构体（对应 DB Schema）
- [x] TASK-21: 新增 `internal/rule/repository.go` —— 定义 `RuleRepository` 接口 + `FakeRepository`（in-memory，用于测试）
- [x] TASK-22: 新增 `internal/rule/mysql_repository.go` —— 实现基于 `database/sql` 的 MySQL RuleRepository
- [x] TASK-23: 新增 `configs/migrations/001_create_risk_rules.sql` —— 建表 SQL（含索引）
- [ ] TASK-24: 修改 `internal/config/config.go` —— 新增 `database` 配置块（DSN、连接池参数）；新增热加载 `source: db` 模式

## Phase 6: 规则加载器集成（替换文件 loader）

- [x] TASK-25: 新增 `internal/rule/loader.go` —— 实现 `DBLoader`：从 RuleRepository 加载规则，调用 `dsl.Compile` 编译，返回 `[]Rule`
- [x] TASK-26: dslRule 内嵌在 loader.go，evaluator.go 保持不变（Rule 接口已抽象，dslRule 实现接口）

## Phase 7: 管理 API

- [x] TASK-27: 新增 `api/http/admin/v1/dto.go` —— 定义请求/响应 DTO
- [x] TASK-28: 新增 `api/http/admin/v1/rules.go` —— 实现 9 个管理 API handler
- [ ] TASK-29: 修改 `cmd/server/main.go` —— 注册 `/admin/v1` 路由组

## Phase 8: DB 热加载 Watcher

- [x] TASK-30: 新增 `internal/rule/watcher.go` —— 实现 `DBWatcher`：轮询 + atomic swap + ForceReload

## Phase 9: 测试 & Benchmark

- [x] TASK-31: 新增 `pkg/dsl/compiler_test.go` —— 表驱动单测（通过）
- [x] TASK-32: Benchmark 内嵌在 compiler_test.go：29ns/op（简单）, 97ns/op（含features）—— **远超目标**
- [x] TASK-33: 新增 `pkg/dsl/compat_test.go` —— 全部 5 条 payment_rules.yaml 条件编译通过
- [ ] TASK-34: 新增 `internal/rule/loader_test.go` —— 用 fakeRepository 测试 DBLoader
- [ ] TASK-35: 新增 `internal/rule/evaluator_bench_test.go` —— `BenchmarkEvaluate100Rules`

## Phase 10: 收尾

- [x] TASK-36: 更新 `CHANGELOG.md`
- [ ] TASK-37: 更新 `openspec/specs/rule-engine.md`
- [ ] TASK-38: 执行 `/opsx:archive`
