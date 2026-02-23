# Proposal: 自研风控规则 DSL（ANTLR4-Go）

## Motivation

当前规则条件使用 `antonmedv/expr`（第三方表达式库）封装在 `pkg/expr/evaluator.go`，规则通过 YAML 文件配置。这一方案存在以下瓶颈：

1. **语法不可控**：无法添加风控专属语法糖（如 `velocity('pay', '1m') > 5`、`inList('blacklist', phone)`），只能通过 Go 函数注入绕行，语义不直观。
2. **配置方式局限**：YAML 文件只能由工程师维护，风控分析师无法自助配置规则。
3. **无管理平台**：规则变更无审批、无版本、无灰度，缺乏运营工具链。
4. **可扩展性受限**：底层库迭代路线由第三方决定，无法为未来的调试工具、规则测试沙箱、可视化构建器提供 AST 基础。

## Goals

- [ ] 自研 `pkg/dsl` 包：基于 ANTLR4-Go 生成 Parser，将规则 condition 字符串编译为 Go 闭包（`Program`），运行时无 ANTLR4 开销
- [ ] 规则存储迁移到数据库（MySQL/PG），替代 YAML 文件
- [ ] 新增管理 API `api/http/admin/v1/rules`：支持 CRUD 和 DSL 语法校验
- [ ] 支持可视化规则构建器（前端，分析师无需写 DSL）+ 高级手写 DSL 模式（工程师）
- [ ] 热加载机制从文件轮询改为 DB 轮询，保留 < 30s 传播目标

## Non-Goals

- 不实现完整的规则管理前端（前端为独立 Phase，本 change 只交付后端 API）
- 不实现 CDC（binlog 监听）热加载，初期用轮询，预留接口
- 不删除 `pkg/expr`，保留以降低迁移风险，待全量切换后再归档
- 不变更 `rule.Rule`、`engine.Engine` 接口签名

## Success Criteria

- Performance: 单条规则 `Program.Run()` P99 < 500ns（Benchmark 验证）
- Performance: 100 条规则 Evaluate P99 < 10ms（与现状持平）
- Correctness: DSL Compiler 对语法错误在 load time 报错，不在 request path panic
- Correctness: 现有 `payment_rules.yaml` 中所有 condition 表达式可无修改编译通过
- Observability: 管理 API 写操作记录审计日志
- Hot-reload: DB 变更 → 规则生效 < 30s

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|-----------|
| ANTLR4-Go 运行时并发缓存 bug | 中（仅影响 load 阶段） | Load 阶段单线程串行，不暴露于请求热路径 |
| Grammar 设计不完备需多次迭代 | 中 | Phase 1 先实现最小子集，通过测试驱动扩展 |
| DB 引入新依赖（无 DB 的测试环境） | 低 | Repository 接口化，测试用 in-memory fake |
| 与现有 expr 条件不兼容 | 低 | DSL 设计时对 `features['key']` 等现有语法做兼容测试 |
