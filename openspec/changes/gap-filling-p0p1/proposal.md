# Proposal: Gap-Filling P0/P1 — 生产就绪补全

## Motivation

当前引擎核心骨架已就位，但对照设计文档（docs/风控决策引擎深度设计方案.md）
和业界（字节/美团/B站）风控系统实践，存在多个影响生产可用性的关键缺口。
本次变更一次性补齐 P0（影响系统可用）和 P1（影响生产质量）缺口。

## Goals

- [ ] G1: Velocity Fetcher — 接通 sliding.Window 到 feature.Service，提供实时速率特征
- [ ] G2: Prometheus Metrics — 四个黄金信号 + 风控业务指标埋点
- [ ] G3: 令牌桶限流中间件 — 防流量击穿（per-IP + 全局两级）
- [ ] G4: Bloom Filter 接入 List Service — L1 快速过滤，减少 Redis 压力
- [ ] G5: 健康检查深化 — Redis/DB 探活 + k8s /readyz /livez probe
- [ ] G6: gRPC 服务实现 — 补全 proto 已定义的 DecisionService
- [ ] G7: A/B 测试执行逻辑 — ABTestConfig 字段已有，补执行分流
- [ ] G8: 多级降级策略 — 4 级降级 + 熔断器（gobreaker）
- [ ] G9: Policy YAML 加载修复 — loadPolicies() 实际读文件

## Non-Goals

- ONNX 模型推理（P3，第二阶段）
- 图风险查询（P3）
- Kafka 真实 producer（保持 log 模拟）
- 联邦学习 / LLM（P3）

## Success Criteria

- Performance: Velocity Fetcher P99 < 10ms（含 Redis RTT）
- Observability: `riskengine_decision_duration_seconds` histogram 可在 /metrics 查询
- Correctness: 限流正确拒绝超额请求，返回 HTTP 429
- Reliability: Redis 宕机时健康检查返回 503，决策引擎降级到纯规则模式

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Bloom Filter FP 导致漏查 | 低（已有精确 Redis 回查） | FP rate 0.1%，可接受 |
| gobreaker 误触发 | 中（短暂熔断正常服务） | 初始阈值保守配置 |
| gRPC 接入增加接口复杂度 | 低 | 复用 engine.Engine 接口 |
