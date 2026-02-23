# Requirements: Gap-Filling P0/P1

## FR-1: Velocity Fetcher
**Given** feature.Service 已注册 VelocityFetcher  
**When** DecisionRequest 到达，包含 UserID/DeviceID/IP  
**Then** 在 feature.Map 中写入 `velocity.pay_count_1min` / `velocity.pay_count_1hour` / `velocity.pay_count_24hour` 等键，超时（10ms）时写入默认值 0

## FR-2: Prometheus Metrics
**Given** 服务启动后 /metrics 端点暴露  
**When** 一次决策请求完成  
**Then** `riskengine_decision_duration_seconds` histogram 记录耗时，标签含 scene_code / decision；`riskengine_rule_hits_total` counter 按 rule_id 计数；`riskengine_feature_fetch_errors_total` 按 fetcher 计数

## FR-3: 限流中间件
**Given** 客户端以超额 QPS 调用 /api/v1/decision  
**When** 令牌桶耗尽  
**Then** 返回 HTTP 429，body `{"code":"RATE_LIMITED","message":"too many requests"}`；正常请求不受影响

## FR-4: 健康检查
**Given** Kubernetes 配置了 readinessProbe GET /readyz  
**When** Redis 不可达  
**Then** /readyz 返回 HTTP 503，body 含失败组件名称；/livez 始终返回 200（进程存活）

## FR-5: gRPC 服务
**Given** 内部服务通过 gRPC 调用决策引擎  
**When** 调用 DecisionService.Evaluate  
**Then** 返回与 HTTP API 等价的 DecisionResult；proto 已在 api/grpc/proto/decision.proto 定义

## FR-6: A/B 测试分流
**Given** PolicySet.ABTest.Enabled = true，SplitPct = 0.1  
**When** 请求到达  
**Then** 约 10% 流量走实验 Pipeline，90% 走对照 Pipeline；实验组决策结果标记 experiment_id

## FR-7: 多级降级 + 熔断器
**Given** 模型服务或 Redis 连续失败 5 次  
**When** 熔断器打开  
**Then** 后续请求直接走降级策略（跳过故障步骤），不再尝试调用；30 秒后半开探测

## FR-8: Policy YAML 加载
**Given** configs/policies/ 目录下存在 *.yaml 文件  
**When** 服务启动或调用 Reload  
**Then** 所有文件被解析为 []PolicySet 并加载到 Registry

## Non-Functional Requirements
| Attribute | Target |
|-----------|--------|
| Velocity Fetcher P99 | < 10ms |
| Metrics 开销 | < 0.1ms/request |
| 限流精度 | 误差 < 5% |
| 健康检查响应 | < 50ms |
| gRPC Evaluate P99 | < 100ms |

## Acceptance Criteria
- [ ] go test ./... 全通过
- [ ] make lint 无 warning
- [ ] CHANGELOG 更新
