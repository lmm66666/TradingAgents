# API 接口文档

## 通用说明

所有接口均采用统一的 JSON 响应格式：

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

| 字段    | 类型   | 说明                        |
|---------|--------|-----------------------------|
| code    | int    | 0 表示成功，非 0 表示错误   |
| message | string | 提示信息                    |
| data    | any    | 业务数据，错误时为 null     |

---

## 接口列表

### 1. 保存股票历史数据

从行情数据源（Broker）获取指定股票的历史 K 线数据，清洗后写入数据库。

- **Method**: `POST`
- **Path**: `/api/stocks/historical`
- **Content-Type**: `application/json`

#### 请求参数

| 字段 | 类型   | 必填 | 说明                     |
|------|--------|------|--------------------------|
| code | string | 是   | 股票代码，如 `600312`    |

#### 请求示例

```bash
curl -X POST http://localhost:8080/api/stocks/historical \
  -H "Content-Type: application/json" \
  -d '{"code": "600312"}'
```

---

### 2. 补全股票数据

手动触发扫描，检查 daily 和 weekly 表中所有股票代码的数据完整性，自动补充缺失的日线和周线数据。

- **Method**: `POST`
- **Path**: `/api/stocks/append`

#### 请求示例

```bash
curl -X POST http://localhost:8080/api/stocks/append
```

---

### 3. 股票买点扫描

按指定策略名称扫描所有股票，判断最新数据日期是否为买点，返回符合条件的股票代码列表。

- **Method**: `GET`
- **Path**: `/api/stocks/signal`

#### 请求参数

| 字段     | 类型   | 必填 | 说明                                 |
|----------|--------|------|--------------------------------------|
| strategy | string | 是   | 策略名称，如 `daily_b1_buy`          |

#### 请求示例

```bash
# 日线 B1 策略
curl "http://localhost:8080/api/stocks/signal?strategy=daily_b1_buy"

# 周线 B1 策略
curl "http://localhost:8080/api/stocks/signal?strategy=weekly_b1_buy"
```

#### 成功响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "strategy": "daily_b1_buy",
    "codes": ["600312", "000001"]
  }
}
```

#### 响应字段说明

| 字段     | 类型     | 说明                   |
|----------|----------|------------------------|
| strategy | string   | 策略名称               |
| codes    | []string | 符合该策略的股票代码   |

#### 支持的策略名称

| 策略名称        | 说明                          |
|-----------------|-------------------------------|
| daily_b1_buy   | 日线 B1：放量回调 + KDJ超卖 + MA20向上 |
| weekly_b1_buy  | 周线 B1：KDJ超卖 + MA20向上          |
