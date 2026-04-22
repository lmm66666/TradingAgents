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
