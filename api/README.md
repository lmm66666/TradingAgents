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

### 2. 查询股票数据

从数据库读取指定股票的 K 线数据，支持按时间粒度聚合。

- **Method**: `GET`
- **Path**: `/api/stocks/data`

#### 查询参数

| 字段   | 类型   | 必填 | 默认值 | 说明                                                          |
|--------|--------|------|--------|---------------------------------------------------------------|
| code   | string | 是   | -      | 股票代码，如 `600312`                                         |
| scale  | int    | 否   | 240    | 时间粒度（分钟），必须是 240 的倍数，如 240、480、720 等       |
| len    | int    | 否   | 240    | 返回最大条数                                                  |

> **说明**：数据源中的原始数据为 240 分钟（日线）级别，`scale` 通过将多条原始数据聚合实现。例如 `scale=480` 表示每 2 条原始数据聚合成一条返回。

#### 请求示例

```bash
# 默认参数
curl "http://localhost:8080/api/stocks/data?code=600312"

# 指定聚合粒度与返回条数
curl "http://localhost:8080/api/stocks/data?code=600312&scale=480&len=100"
```

#### 成功响应

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "code": "600312",
      "date": "2025-04-21 15:00:00",
      "open": 24.5000,
      "high": 25.1200,
      "low": 24.3000,
      "close": 24.9800,
      "volume": 12345678
    }
  ]
}
```

#### 响应字段说明

| 字段   | 类型    | 说明               |
|--------|---------|--------------------|
| code   | string  | 股票代码           |
| date   | string  | 日期时间           |
| open   | float64 | 开盘价             |
| high   | float64 | 最高价             |
| low    | float64 | 最低价             |
| close  | float64 | 收盘价             |
| volume | int64   | 成交量             |

#### 错误响应

```json
{
  "code": 400,
  "message": "code is required",
  "data": null
}
```
