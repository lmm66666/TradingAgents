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

#### 成功响应

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

#### 错误响应

```json
{
  "code": 400,
  "message": "code is required",
  "data": null
}
```

---

### 2. 查询股票分析数据

从数据库读取指定股票的日线、周线 K 线数据，并计算 MACD、KDJ 技术指标。

- **Method**: `GET`
- **Path**: `/api/stocks/analysis`

#### 查询参数

| 字段 | 类型   | 必填 | 说明                  |
|------|--------|------|-----------------------|
| code | string | 是   | 股票代码，如 `600312` |

#### 请求示例

```bash
curl "http://localhost:8080/api/stocks/analysis?code=600312"
```

#### 成功响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "daily": [
      {
        "date": "2025-04-21",
        "price": 24.9800,
        "volume": 12345678,
        "j": 78.7654,
        "dea": 0.4123,
        "ma5": 24.5000,
        "ma20": 23.8000,
        "ma60": 22.5000
      }
    ],
    "weekly": [
      {
        "date": "2025-04-18",
        "price": 24.9800,
        "volume": 56789012,
        "j": 66.6666,
        "dea": 0.2345,
        "ma5": 24.2000,
        "ma20": 23.5000,
        "ma60": 21.8000
      }
    ]
  }
}
```

#### 响应字段说明

| 字段   | 类型    | 说明                          |
|--------|---------|-------------------------------|
| daily  | array   | 日线分析数据（最近 100 条）   |
| weekly | array   | 周线分析数据（最近 50 条）    |

**分析数据字段**

| 字段   | 类型    | 说明                  |
|--------|---------|-----------------------|
| date   | string  | 日期                  |
| price  | float64 | 收盘价                |
| volume | int64   | 成交量                |
| j      | float64 | KDJ 的 J 值（3K - 2D）|
| dea    | float64 | MACD 的 DEA 线        |
| ma5    | float64 | 5 日均线              |
| ma20   | float64 | 20 日均线             |
| ma60   | float64 | 60 日均线             |

#### 错误响应

```json
{
  "code": 400,
  "message": "code is required",
  "data": null
}
```

---

### 3. 补全股票数据

手动触发扫描，检查 daily 和 weekly 表中所有股票代码的数据完整性，自动补充缺失的日线和周线数据。

- **Method**: `POST`
- **Path**: `/api/stocks/append`

#### 请求示例

```bash
curl -X POST http://localhost:8080/api/stocks/append
```

#### 成功响应

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

#### 错误响应

```json
{
  "code": 500,
  "message": "trigger failed",
  "data": null
}
```

---

### 4. 策略信号扫描

使用指定策略扫描股票，返回匹配信号列表。支持全市场扫描或指定股票代码列表。

- **Method**: `POST`
- **Path**: `/api/patterns/scan`
- **Content-Type**: `application/json`

#### 请求参数

| 字段      | 类型     | 必填 | 说明                                          |
|-----------|----------|------|-----------------------------------------------|
| strategy  | string   | 是   | 策略名称，如 `volume_surge_pullback`          |
| min_score | float64  | 否   | 最低匹配分数，默认 `70`                       |
| codes     | string[] | 否   | 指定股票代码列表，不传则扫描全市场            |

#### 请求示例

```bash
curl -X POST http://localhost:8080/api/patterns/scan \
  -H "Content-Type: application/json" \
  -d '{"strategy": "volume_surge_pullback", "min_score": 70}'
```

#### 成功响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "signals": [
      {
        "code": "600312",
        "date": "2026-04-22",
        "strategy": "volume_surge_pullback",
        "type": "watch",
        "phase": "pullback",
        "score": 82.5,
        "sub_scores": {
          "volume_surge": 90,
          "price_rally": 75,
          "pullback_volume": 85,
          "pullback_depth": 80,
          "time_rhythm": 100,
          "ma_support": 75
        },
        "context": {
          "surge_date": "2026-04-10",
          "peak_date": "2026-04-14",
          "surge_volume": 350000,
          "avg_pullback_vol": 98000,
          "max_pullback_pct": 6.5,
          "pullback_days": 4
        }
      }
    ]
  }
}
```

#### 支持的策略

| 策略名称                  | 说明                     |
|---------------------------|--------------------------|
| `volume_surge_pullback`   | 放量上涨后缩量回调形态   |

#### 错误响应

```json
{
  "code": 400,
  "message": "unknown strategy",
  "data": null
}
```

---

### 5. 策略历史回测

对指定股票使用指定策略进行历史回测，计算信号产生后持有 N 天的收益表现。

- **Method**: `GET`
- **Path**: `/api/patterns/backtest`

#### 查询参数

| 字段       | 类型    | 必填 | 说明                                      |
|------------|---------|------|-------------------------------------------|
| code       | string  | 是   | 股票代码，如 `600312`                     |
| strategy   | string  | 是   | 策略名称，如 `volume_surge_pullback`      |
| hold_days  | int     | 否   | 持有天数，默认 `5`                        |

#### 请求示例

```bash
curl "http://localhost:8080/api/patterns/backtest?code=600312&strategy=volume_surge_pullback&hold_days=5"
```

#### 成功响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "strategy_name": "volume_surge_pullback",
    "total_trades": 45,
    "win_rate": 0.62,
    "avg_return": 3.5,
    "max_drawdown": 12.0,
    "profit_factor": 1.8,
    "trades": [
      {
        "entry_date": "2026-01-15",
        "exit_date": "2026-01-22",
        "return_pct": 5.2
      }
    ]
  }
}
```

#### 响应字段说明

| 字段          | 类型          | 说明                                    |
|---------------|---------------|-----------------------------------------|
| strategy_name | string        | 策略名称                                |
| total_trades  | int           | 信号总数                                |
| win_rate      | float64       | 胜率（0~1）                             |
| avg_return    | float64       | 平均收益率（百分比）                    |
| max_drawdown  | float64       | 最大回撤（百分比）                      |
| profit_factor | float64       | 盈亏比（盈利总额/亏损总额）             |
| trades        | []TradeRecord | 每笔交易明细                            |

**TradeRecord 字段**

| 字段       | 类型    | 说明                |
|------------|---------|---------------------|
| entry_date | string  | 买入日期            |
| exit_date  | string  | 卖出日期            |
| return_pct | float64 | 收益率（百分比）    |

#### 错误响应

```json
{
  "code": 400,
  "message": "code and strategy are required",
  "data": null
}
```
