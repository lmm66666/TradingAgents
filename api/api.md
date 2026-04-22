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
        "code": "600312",
        "date": "2025-04-21",
        "open": 24.5000,
        "high": 25.1200,
        "low": 24.3000,
        "close": 24.9800,
        "volume": 12345678
      }
    ],
    "weekly": [
      {
        "code": "600312",
        "date": "2025-04-18",
        "open": 23.8000,
        "high": 25.5000,
        "low": 23.5000,
        "close": 24.9800,
        "volume": 56789012
      }
    ],
    "daily_macd": [
      {
        "date": "2025-04-21",
        "dif": 0.5234,
        "dea": 0.4123,
        "bar": 0.2222
      }
    ],
    "weekly_macd": [
      {
        "date": "2025-04-18",
        "dif": 0.3456,
        "dea": 0.2345,
        "bar": 0.2222
      }
    ],
    "daily_kdj": [
      {
        "date": "2025-04-21",
        "k": 65.4321,
        "d": 58.7654,
        "j": 78.7654
      }
    ],
    "weekly_kdj": [
      {
        "date": "2025-04-18",
        "k": 55.5555,
        "d": 50.0000,
        "j": 66.6666
      }
    ]
  }
}
```

#### 响应字段说明

| 字段        | 类型    | 说明                          |
|-------------|---------|-------------------------------|
| daily       | array   | 日线 K 线数据（最近 100 条）  |
| weekly      | array   | 周线 K 线数据（最近 50 条）   |
| daily_macd  | array   | 日线 MACD 指标                |
| weekly_macd | array   | 周线 MACD 指标                |
| daily_kdj   | array   | 日线 KDJ 指标                 |
| weekly_kdj  | array   | 周线 KDJ 指标                 |

**K 线字段**

| 字段   | 类型    | 说明     |
|--------|---------|----------|
| code   | string  | 股票代码 |
| date   | string  | 日期     |
| open   | float64 | 开盘价   |
| high   | float64 | 最高价   |
| low    | float64 | 最低价   |
| close  | float64 | 收盘价   |
| volume | int64   | 成交量   |

**MACD 字段**

| 字段 | 类型    | 说明                  |
|------|---------|-----------------------|
| date | string  | 日期                  |
| dif  | float64 | DIF 线（EMA12 - EMA26）|
| dea  | float64 | DEA 线（DIF 的 EMA9）  |
| bar  | float64 | 柱状图（2 * (DIF - DEA)）|

**KDJ 字段**

| 字段 | 类型    | 说明          |
|------|---------|---------------|
| date | string  | 日期          |
| k    | float64 | K 值          |
| d    | float64 | D 值          |
| j    | float64 | J 值（3K - 2D）|

#### 错误响应

```json
{
  "code": 400,
  "message": "code is required",
  "data": null
}
```
