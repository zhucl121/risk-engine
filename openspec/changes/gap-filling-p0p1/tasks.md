# Tasks: Gap-Filling P0/P1
<!-- Implementor: one task at a time. Mark [x] only after make test passes. -->
<!-- Commit format: feat(scope): TASK-N description -->

## Phase 1: Velocity Fetcher
- [ ] TASK-1: 新增 `internal/feature/fetchers/velocity_fetcher.go` — 实现 feature.Fetcher，用 sliding.Window 计算 1min/1h/24h 速率
- [ ] TASK-2: 新增 `internal/feature/fetchers/velocity_fetcher_test.go` — 单元测试（mock redis）
- [ ] TASK-3: 修改 `cmd/server/main.go` — 注册 VelocityFetcher 到 feature.Service

## Phase 2: Prometheus Metrics
- [ ] TASK-4: 新增 `internal/metrics/metrics.go` — 定义所有 Prometheus Collector（histogram/counter/gauge）
- [ ] TASK-5: 新增 `internal/middleware/metrics.go` — gin 中间件，记录请求延迟和状态码
- [ ] TASK-6: 修改 `cmd/server/main.go` — 注册 /metrics 端点，注册 metrics 中间件

## Phase 3: 限流中间件
- [ ] TASK-7: 新增 `internal/middleware/ratelimit.go` — 令牌桶限流（golang.org/x/time/rate），per-IP + 全局
- [ ] TASK-8: 修改 `cmd/server/main.go` — 注册限流中间件

## Phase 4: 健康检查深化
- [ ] TASK-9: 新增 `internal/health/checker.go` — 定义 Checker 接口 + RedisChecker + CompositeChecker
- [ ] TASK-10: 修改 `api/http/v1/handler.go` — /readyz 探活所有依赖，/livez 始终 200
- [ ] TASK-11: 修改 `cmd/server/main.go` — 注册 health checker

## Phase 5: gRPC 服务
- [ ] TASK-12: 新增 `api/grpc/server/decision_server.go` — 实现 proto 生成的 DecisionServiceServer
- [ ] TASK-13: 修改 `cmd/server/main.go` — 启动 gRPC listener（端口从 cfg.Server.GRPCAddr）

## Phase 6: A/B 测试执行
- [ ] TASK-14: 修改 `internal/orchestrator/registry.go` — 加载 ABTest 配置
- [ ] TASK-15: 修改 `internal/orchestrator/pipeline.go` — Execute 时根据 ABTestConfig 分流，实验组结果追加 experiment_id

## Phase 7: 多级降级 + 熔断器
- [ ] TASK-16: 新增 `internal/resilience/breaker.go` — 封装 gobreaker，提供 Execute(fn) 接口
- [ ] TASK-17: 修改 `internal/orchestrator/pipeline.go` — 在 dispatchModel/dispatchList 时包裹熔断器

## Phase 8: Policy YAML 加载修复
- [ ] TASK-18: 修改 `internal/orchestrator/registry.go` — LoadFromYAML 支持单文件多文档（---分隔）
- [ ] TASK-19: 修改 `cmd/server/main.go` — loadPolicies() 实际遍历目录读取文件
- [ ] TASK-20: 新增 `configs/policies/payment.yaml` — 示例 PolicySet YAML

## Phase 9: Wrap-up
- [ ] TASK-21: 更新 `CHANGELOG.md`
- [ ] TASK-22: 运行 `/opsx:archive`
