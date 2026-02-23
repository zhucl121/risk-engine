# RiskDSL 语法使用指南

RiskDSL 是风控决策引擎内置的**领域专用语言**（DSL），专为风控规则条件表达式设计。  
它基于 ANTLR4 自研语法，在规则加载时编译为 Go 闭包，运行时**零解析开销**。

---

## 目录

1. [设计原则](#设计原则)
2. [数据类型](#数据类型)
3. [内置变量（请求字段）](#内置变量请求字段)
4. [特征访问](#特征访问)
5. [运算符](#运算符)
6. [内置函数参考](#内置函数参考)
7. [规则配置示例](#规则配置示例)
8. [扩展自定义函数](#扩展自定义函数)
9. [错误处理与降级](#错误处理与降级)
10. [运算符优先级表](#运算符优先级表)

---

## 设计原则

| 约束 | 说明 |
|------|------|
| **非图灵完备** | 无循环、无赋值、无副作用，表达式必须终止 |
| **纯函数求值** | 每次执行结果只依赖输入，适合并发场景 |
| **短路求值** | `&&` 左侧为 `false` 时不计算右侧；`\|\|` 左侧为 `true` 时不计算右侧 |
| **失败降级** | 外部服务（名单、模型、GeoIP）不可用时返回安全默认值，不中断决策 |
| **零运行时开销** | 规则加载时编译为 Go 闭包，运行时只调用闭包，无 ANTLR4 代码 |
| **可扩展** | 通过 `FunctionRegistry` 注册自定义函数，无需修改语法文件 |

---

## 数据类型

DSL 支持以下基础类型：

| 类型 | 字面量示例 | 说明 |
|------|-----------|------|
| `int` | `100`、`0`、`9999` | 64 位整数 |
| `float` | `3.14`、`0.5`、`1000.0` | 64 位浮点数 |
| `string` | `'hello'`、`"world"` | 单引号或双引号均可 |
| `bool` | `true`、`false` | 布尔值 |
| `null` | `null`、`nil` | 空值，数值比较时视为 `0` |
| `array` | `[1, 'a', true]` | 仅用于 `in`/`not in` 右侧 |
| `object` | 仅由函数返回 | 如 `geoIP(ip)` 返回的地理信息对象 |

> **注意**：DSL 不支持负数字面量。如需负数，用 `0 - n`，或在特征预处理阶段注入。

---

## 内置变量（请求字段）

以下标识符直接映射到 `DecisionRequest` 的字段，无需前缀即可使用：

| 变量名 | 类型 | 说明 |
|--------|------|------|
| `amount` | `int` | 交易金额（分） |
| `userID` | `string` | 用户唯一标识 |
| `deviceID` | `string` | 设备指纹 |
| `ip` | `string` | 请求来源 IP |
| `phone` | `string` | 手机号（来自 `Extra["phone"]`） |

**示例：**
```
amount > 100000
userID == 'u_12345'
ip != '127.0.0.1'
```

---

## 特征访问

引擎拉取特征后，所有特征值通过 `features['key']` 访问：

```
features['user.register_days'] > 7
features['device.is_rooted'] == true
features['payment.channel'] == 'WECHAT'
features['extra.merchant_id'] == 'M001'
```

> `Extra` 字段会自动注入为 `extra.<key>`，例如 `features['extra.merchant_id']`。

---

## 运算符

### 比较运算符

| 运算符 | 说明 | 适用类型 |
|--------|------|---------|
| `==` | 等于 | int、float、string、bool |
| `!=` | 不等于 | int、float、string、bool |
| `>` | 大于 | int、float |
| `<` | 小于 | int、float |
| `>=` | 大于等于 | int、float |
| `<=` | 小于等于 | int、float |

```
amount >= 500
features['user.credit_score'] < 600
```

### 逻辑运算符

| 运算符 | 说明 | 短路规则 |
|--------|------|---------|
| `&&` | 逻辑与 | 左侧为 `false` 时不计算右侧 |
| `\|\|` | 逻辑或 | 左侧为 `true` 时不计算右侧 |
| `!` | 逻辑非 | — |

```
amount > 10000 && features['user.register_days'] < 3
geoIP(ip).country == 'CN' || inList('whitelist_ip', ip)
!inList('blacklist_device', deviceID)
```

### 成员运算符（in / not in）

检查值是否在数组中，右侧必须为数组字面量：

```
features['payment.channel'] in ['ALIPAY', 'WECHAT', 'UNIONPAY']
ip not in ['192.168.0.1', '10.0.0.1']
amount in [100, 500, 1000]
```

### 三元运算符（? :）

条件表达式，结果通常内嵌在比较中使用：

```
(amount > 50000 ? amount : 0) > 30000
(features['user.vip_level'] == 'gold' ? 1 : 0) == 1
```

> 顶层规则条件**必须返回 bool**，三元表达式若作为顶层表达式需确保两侧类型一致为 bool。

### 括号分组

括号用于覆盖默认优先级：

```
(amount > 1000 || features['user.register_days'] < 1) && !inList('whitelist', userID)
```

---

## 内置函数参考

### 风控核心函数

#### `inList(listName, value)` → bool

查询命中名单。名单数据存储在 Redis List Service 中，异步加载后缓存。

```
inList('blacklist_user', userID)
inList('high_risk_ip', ip)
inList('whitelist_merchant', features['extra.merchant_id'])
```

| 参数 | 类型 | 说明 |
|------|------|------|
| `listName` | string | 名单名称 |
| `value` | string | 待查询的值 |

> 服务不可用时**降级为 `false`**（fail-open）。

---

#### `velocity(event, window)` → int

查询当前用户在时间窗口内的事件计数（基于 Redis 滑动窗口）。

```
velocity('payment', '1m') > 5
velocity('login_fail', '5m') >= 3
velocity('register', '24h') > 10
```

| 参数 | 类型 | 说明 |
|------|------|------|
| `event` | string | 事件名称 |
| `window` | string | 时间窗口：`1m`、`5m`、`1h`、`24h` 等 |

> 服务不可用时降级为 `0`。

---

#### `modelScore(modelName)` → float

调用 ML 模型评分接口。

```
modelScore('device_risk_v2') > 0.85
modelScore('transaction_fraud') >= 0.7
```

| 参数 | 类型 | 说明 |
|------|------|------|
| `modelName` | string | 模型注册名称 |

> 服务不可用时降级为 `0.0`。

---

#### `geoIP(ip)` → object

IP 地理位置查询，返回可字段访问的对象。

```
geoIP(ip).country == 'CN'
geoIP(ip).isProxy == true
geoIP(ip).country != 'US' && geoIP(ip).country != 'CN'
```

返回对象字段：

| 字段名 | 类型 | 说明 |
|--------|------|------|
| `.country` | string | 国家代码（ISO 3166-1 alpha-2，如 `CN`、`US`） |
| `.isp` | string | ISP 名称 |
| `.asn` | string | ASN 编号 |
| `.isProxy` | bool | 是否为代理/VPN/Tor |

---

#### `within(value, center, radiusKm)` → bool

地理围栏判断（值是否在以 center 为圆心、radiusKm 公里内）。

```
within(features['user.location'], features['merchant.location'], 50.0)
```

---

### 字符串函数

| 函数 | 签名 | 说明 |
|------|------|------|
| `contains` | `(s, substr) → bool` | 是否包含子串 |
| `startsWith` | `(s, prefix) → bool` | 是否以 prefix 开头 |
| `endsWith` | `(s, suffix) → bool` | 是否以 suffix 结尾 |
| `match` | `(s, pattern) → bool` | 正则匹配（RE2 语法） |
| `lower` | `(s) → string` | 转小写 |
| `upper` | `(s) → string` | 转大写 |
| `trim` | `(s) → string` | 去除首尾空白 |
| `strlen` | `(s) → int` | 字符串长度（Unicode 字符数） |
| `isEmpty` | `(s) → bool` | 是否为空字符串 |

**示例：**
```
contains(features['user.email'], '@gmail.com')
match(phone, '^1[3-9]\\d{9}$')
startsWith(features['device.model'], 'iPhone')
strlen(features['extra.remark']) > 200
isEmpty(features['extra.coupon_code'])
```

---

### 数学函数

| 函数 | 签名 | 说明 |
|------|------|------|
| `abs` | `(n) → float` | 绝对值 |
| `ceil` | `(n) → float` | 向上取整 |
| `floor` | `(n) → float` | 向下取整 |
| `round` | `(n) → float` | 四舍五入 |
| `sqrt` | `(n) → float` | 平方根 |
| `min` | `(a, b) → float` | 两数取小 |
| `max` | `(a, b) → float` | 两数取大 |
| `clamp` | `(v, lo, hi) → float` | 将 v 限制在 [lo, hi] 区间内 |

**示例：**
```
abs(features['account.balance']) < 100.0
max(modelScore('fraud_v1'), modelScore('fraud_v2')) > 0.9
clamp(velocity('payment', '1h'), 0.0, 100.0) > 50.0
```

---

### 时间函数

| 函数 | 签名 | 说明 |
|------|------|------|
| `now` | `() → int` | 当前 Unix 时间戳（秒） |
| `nowMs` | `() → int` | 当前 Unix 时间戳（毫秒） |
| `daysSince` | `(timeStr) → int` | 距指定时间的天数 |
| `hoursSince` | `(timeStr) → int` | 距指定时间的小时数 |
| `secondsSince` | `(timeStr) → int` | 距指定时间的秒数 |
| `toUnix` | `(timeStr) → int` | 时间字符串转 Unix 时间戳 |
| `hour` | `() → int` | 当前 UTC 小时（0-23） |
| `weekday` | `() → int` | 当前星期几（0=周日，6=周六） |

**时间字符串支持格式：**
```
2006-01-02T15:04:05
2006-01-02 15:04:05
2006-01-02
20060102
```

**示例：**
```
daysSince(features['user.register_time']) < 7
hoursSince(features['account.last_login']) > 720
hour() >= 0 && hour() <= 6
weekday() == 0 || weekday() == 6
```

---

### 类型转换与条件函数

| 函数 | 签名 | 说明 |
|------|------|------|
| `toInt` | `(v) → int` | 转整数（字符串/浮点/布尔） |
| `toFloat` | `(v) → float` | 转浮点数 |
| `toString` | `(v) → string` | 转字符串 |
| `toBool` | `(v) → bool` | 转布尔（`"true"`/`"1"`/`"yes"` → true） |
| `isNull` | `(v) → bool` | 是否为 null/nil |
| `coalesce` | `(a, b, ...) → any` | 返回第一个非 null 值 |
| `ifThen` | `(cond, trueVal, falseVal) → any` | 函数式三元 |

**示例：**
```
toInt(features['extra.level']) >= 3
isNull(features['user.id_number']) == false
coalesce(features['extra.amount'], amount) > 50000
ifThen(features['user.is_vip'] == true, 1, 0) == 1
```

---

## 规则配置示例

规则条件在数据库或管理页面中直接配置，每条规则的 `condition` 字段填写 DSL 表达式。

### 基础风险规则

```yaml
# 大额新用户规则
- name: high_amount_new_user
  group: payment_risk
  condition: "amount > 500000 && daysSince(features['user.register_time']) < 3"
  action: MANUAL_REVIEW
  weight: 80

# 黑名单设备
- name: blacklist_device
  group: device_risk
  condition: "inList('blacklist_device', deviceID)"
  action: REJECT
  weight: 100

# 境外代理 IP
- name: overseas_proxy
  group: ip_risk
  condition: "geoIP(ip).isProxy == true || (geoIP(ip).country != 'CN' && amount > 10000)"
  action: MANUAL_REVIEW
  weight: 70
```

### 频控规则

```yaml
# 1 分钟内支付超 5 次
- name: payment_velocity_1m
  group: velocity_control
  condition: "velocity('payment', '1m') > 5"
  action: REJECT
  weight: 90

# 24 小时内注册设备登录失败超 10 次
- name: login_fail_device_24h
  group: velocity_control
  condition: "velocity('login_fail', '24h') >= 10"
  action: MANUAL_REVIEW
  weight: 75
```

### 模型集成规则

```yaml
# 设备风险模型 + 交易金额双重过滤
- name: device_model_high_amount
  group: model_risk
  condition: "modelScore('device_risk_v2') > 0.85 && amount > 100000"
  action: REJECT
  weight: 95

# 多模型取最高分
- name: ensemble_model_risk
  group: model_risk
  condition: "max(modelScore('fraud_v1'), modelScore('fraud_v2')) > 0.9"
  action: REJECT
  weight: 95
```

### 字符串与类型处理

```yaml
# 手机号格式校验（格式异常拦截）
- name: invalid_phone
  group: data_quality
  condition: "!match(phone, '^1[3-9]\\d{9}$') && !isEmpty(phone)"
  action: REJECT
  weight: 60

# 邮箱域名白名单
- name: corporate_email
  group: user_profile
  condition: "contains(features['user.email'], '@company.com') == false"
  action: MANUAL_REVIEW
  weight: 40

# 注册时间较新且高风险渠道
- name: new_user_risky_channel
  group: channel_risk
  condition: "daysSince(features['user.register_time']) < 1 && features['payment.channel'] in ['CRYPTO', 'GIFT_CARD']"
  action: REJECT
  weight: 90
```

### 时间窗口规则

```yaml
# 凌晨高额交易
- name: midnight_high_amount
  group: time_risk
  condition: "hour() >= 1 && hour() <= 5 && amount > 50000"
  action: MANUAL_REVIEW
  weight: 65

# 周末大额充值
- name: weekend_large_topup
  group: time_risk
  condition: "(weekday() == 0 || weekday() == 6) && amount > 200000"
  action: MANUAL_REVIEW
  weight: 55
```

### 复合条件

```yaml
# 多维度综合判断
- name: comprehensive_fraud
  group: comprehensive
  condition: |
    (inList('blacklist_user', userID) || velocity('payment', '5m') > 10)
    && amount > 100000
    && geoIP(ip).isProxy == true
  action: REJECT
  weight: 100

# 会员豁免高额限制
- name: vip_large_payment
  group: vip_policy
  condition: |
    amount > 500000
    && features['user.vip_level'] not in ['gold', 'platinum']
    && daysSince(features['user.register_time']) < 30
  action: MANUAL_REVIEW
  weight: 70
```

---

## 扩展自定义函数

当内置函数不满足需求时，可通过 `FunctionRegistry` 注册自定义 Go 函数，**无需修改语法文件**。

### 注册方式

```go
package main

import (
    "context"
    "github.com/yourorg/riskengine/pkg/dsl"
)

func main() {
    reg := dsl.NewFunctionRegistry()

    // 方式一：使用 RegisterFunc（快速注册，无类型检查）
    _ = reg.RegisterFunc("isCorporateUser", func(ctx context.Context, rt *dsl.Runtime, args []dsl.Value) (dsl.Value, error) {
        if len(args) != 1 {
            return dsl.NilValue(), fmt.Errorf("isCorporateUser: expects 1 arg")
        }
        userID := args[0].Str()
        // 自定义业务逻辑
        isCorp := strings.HasPrefix(userID, "corp_")
        return dsl.BoolValue(isCorp), nil
    })

    // 方式二：使用 Register + FuncDef（完整签名，支持静态类型检查）
    _ = reg.Register(dsl.FuncDef{
        Name:       "riskLevel",
        Args:       []dsl.ArgKind{dsl.ArgKindString},  // 1 个 string 参数
        ReturnKind: dsl.KindInt,                         // 返回 int
        Impl: func(ctx context.Context, rt *dsl.Runtime, args []dsl.Value) (dsl.Value, error) {
            level := queryRiskLevelFromCache(args[0].Str())
            return dsl.IntValue(int64(level)), nil
        },
    })
}
```

### 在 DSL 中使用自定义函数

```
isCorporateUser(userID) == true && amount > 1000000
riskLevel(features['extra.merchant_id']) >= 3
```

### 访问请求上下文

自定义函数通过 `rt *dsl.Runtime` 访问请求信息：

```go
func myFunc(ctx context.Context, rt *dsl.Runtime, args []dsl.Value) (dsl.Value, error) {
    userID   := rt.Request.UserID    // 用户 ID
    deviceID := rt.Request.DeviceID  // 设备 ID
    ip       := rt.Request.IP        // IP 地址
    amount   := rt.Request.Amount    // 交易金额
    extras   := rt.Request.Extra     // Extra 扩展字段 map[string]string
    features := rt.Features          // feature.Map — 所有特征值
    
    // 使用 List Service
    hit, _ := rt.ListChecker.InList(ctx, "my_list", userID)
    return dsl.BoolValue(hit), nil
}
```

---

## 错误处理与降级

RiskDSL 采用**失败降级**策略，外部服务故障不会导致规则评估中断：

| 场景 | 降级行为 |
|------|---------|
| `inList` 名单服务不可用 | 返回 `false`（不命中） |
| `velocity` Redis 不可用 | 返回 `0` |
| `modelScore` 模型服务不可用 | 返回 `0.0` |
| `geoIP` 地理服务不可用 | 返回空对象（`country=""`, `isProxy=false`） |
| 特征键不存在（`features['key']`） | 返回 `null`（数值比较视为 `0`） |
| 编译时发现未知函数 | 编译失败，拒绝加载该规则 |
| 运行时类型不匹配 | 返回 `false`，记录错误日志 |

> 可通过 Prometheus 指标 `riskengine_rule_eval_errors_total` 监控运行时降级次数。

---

## 运算符优先级表

从高到低（数字越小优先级越高）：

| 优先级 | 运算符/元素 | 说明 |
|--------|------------|------|
| 1 | 函数调用、`features[]`、字段访问、字面量、标识符 | 原子元素 |
| 2 | `!` | 逻辑非（一元） |
| 3 | `>` `<` `>=` `<=` `==` `!=` | 比较运算 |
| 4 | `in` `not in` | 成员测试 |
| 5 | `&&` `\|\|` | 逻辑与/或（短路） |
| 6 | `? :` | 三元条件（右结合） |

**括号可覆盖任意优先级：**
```
amount > 1000 || (velocity('payment', '5m') > 3 && !inList('whitelist', userID))
```

---

## 常见问题

**Q: DSL 表达式里能写多行吗？**  
A: 可以。空白字符（空格、换行、Tab）都会被忽略，YAML 中使用 `|` 块字符串即可：
```yaml
condition: |
  amount > 100000
  && daysSince(features['user.register_time']) < 7
  && !inList('whitelist_user', userID)
```

**Q: 字符串比较区分大小写吗？**  
A: 区分。如需忽略大小写，用 `lower()` 或 `upper()` 先转换：
```
lower(features['user.email']) == 'admin@example.com'
```

**Q: features 里的值类型是什么？**  
A: 特征值类型由 Feature Store 返回时决定，支持 int、float、string、bool。  
`Extra` 字段来自请求参数，默认注入为 string，可通过 `extraSchema` 或 DB 配置类型转换。

**Q: 规则条件最终必须是 bool 类型吗？**  
A: 是。规则顶层表达式必须返回 `bool`，否则运行时报类型错误、该规则不命中。  
三元表达式 `? :` 用作子表达式（包裹在比较或逻辑运算中）是合法的。

**Q: 如何验证一条规则语法是否正确？**  
A: 调用管理 API 进行编译检查：
```bash
curl -X POST http://localhost:8080/admin/v1/rules/validate \
  -H 'Content-Type: application/json' \
  -d '{"condition": "amount > 100 && inList(\"blacklist\", userID)"}'
```
