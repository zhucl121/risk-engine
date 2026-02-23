# Requirements: 自研风控规则 DSL（ANTLR4-Go）

## Functional Requirements

### FR-1: DSL 编译器（Compiler）
**Given** 一条合法的 DSL condition 字符串和已注册的 FunctionRegistry  
**When** 调用 `dsl.Compile(condition, env)` （规则加载阶段）  
**Then** 返回一个 `*dsl.Program`（Go 闭包树），后续调用 `Program.Run(env)` 无需 ANTLR4 介入

### FR-2: 类型安全
**Given** 一条 DSL 字符串，其中某个操作数类型不兼容（如 `amount > "abc"`）  
**When** 调用 `dsl.Compile()`  
**Then** 返回带清晰位置信息的 `TypeError`，规则加载失败，不影响已运行规则集

### FR-3: 语法错误早期报告
**Given** 一条语法非法的 DSL 字符串（如 `amount > > 5`）  
**When** 调用 `dsl.Compile()` 或管理 API `POST /admin/v1/rules/validate`  
**Then** 返回带行列号的 `SyntaxError`，不写入数据库

### FR-4: 风控语义函数
**Given** DSL 字符串中使用 `inList('blacklist.phone', phone)` 等风控函数  
**When** Compile 和 Run 时  
**Then** 函数通过 `FunctionRegistry` 解析，Compile 时类型检查，Run 时执行对应 Go 实现

内置必须支持的函数：

| 函数签名 | 语义 |
|---|---|
| `inList(listName string, value any) bool` | 命中名单检查（调用 list.Service） |
| `velocity(event string, window string) int` | 速率窗口计数（调用 sliding window） |
| `modelScore(modelName string) float` | ML 模型分数（调用 model.Registry） |
| `geoIP(ip string) GeoInfo` | IP 归属地（返回可访问 `.country` `.isp` 的对象） |
| `within(v, lo, hi number) bool` | 范围判断 `lo <= v <= hi` |

### FR-5: 向后兼容现有表达式
**Given** 现有 `payment_rules.yaml` 中的所有 condition 字符串  
**When** 用新 DSL Compiler 编译  
**Then** 全部编译通过，行为等价

### FR-6: 数据库规则存储
**Given** 管理员通过 API 创建/更新规则  
**When** 调用 `POST /admin/v1/rules`  
**Then** 规则持久化到 DB，`condition` 字段存 DSL 字符串，`condition_ast` 字段存可视化 JSON（可选）

### FR-7: 管理 API
**Given** 合法的管理员请求  
**When** 调用管理 API（见下表）  
**Then** 返回对应操作结果，写操作记录审计日志

必须实现的管理 API：

| 方法 | 路径 | 功能 |
|---|---|---|
| GET | `/admin/v1/rules` | 列表（支持 scene_code、group_name、status 过滤，分页） |
| GET | `/admin/v1/rules/:id` | 详情 |
| POST | `/admin/v1/rules` | 创建（DSL 先 validate 再写入） |
| PUT | `/admin/v1/rules/:id` | 更新（乐观锁 version 字段） |
| DELETE | `/admin/v1/rules/:id` | 软删除（status=0） |
| POST | `/admin/v1/rules/validate` | 仅校验 DSL 语法，不写入 |
| POST | `/admin/v1/rules/:id/enable` | 启用 |
| POST | `/admin/v1/rules/:id/disable` | 禁用 |
| POST | `/admin/v1/rules/reload` | 手动触发热加载 |

### FR-8: DB 热加载
**Given** DB 中某条规则的 `updated_at` 发生变化  
**When** 热加载 watcher 检测到（轮询间隔 ≤ 30s）  
**Then** 仅重新编译变更的规则集，通过 `atomic.Pointer.Store` 替换，in-flight 请求不受影响

### FR-9: 可视化 JSON 双向转换
**Given** 前端提交可视化构建器产生的 `condition_ast` JSON  
**When** 管理 API 接收到请求  
**Then** 服务端将 `condition_ast` 序列化为 DSL 字符串并校验，两者同时存入 DB

## Non-Functional Requirements

| Attribute | Target |
|-----------|--------|
| 单规则 `Program.Run()` P99 | < 500ns |
| 100 条规则 Evaluate P99 | < 10ms |
| 热加载传播延迟 | < 30s |
| DSL Compile（单条） | < 10ms（load time，不计入请求延迟） |
| 管理 API 写入 P99 | < 200ms |
| 内存每条规则 Program 开销 | < 50KB |

## Acceptance Criteria

- [ ] 所有 FR 覆盖表驱动单元测试
- [ ] `BenchmarkRunProgram` 结果 < 500ns/op（GitHub Actions CI 中记录基线）
- [ ] `BenchmarkEvaluate100Rules` 结果 < 10ms/op
- [ ] `make lint` 零 warning
- [ ] 现有 `payment_rules.yaml` 条件全部可被新 Compiler 编译通过（兼容性测试）
- [ ] CHANGELOG.md 更新
- [ ] `openspec/specs/rule-engine.md` 中 DSL 章节更新为新语法
