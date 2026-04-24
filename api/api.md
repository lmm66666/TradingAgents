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

### 3. 多策略信号扫描

使用多个策略扫描股票，返回所有策略交集（某天所有策略都满足）的信号列表。
单策略时返回该策略所有满足条件的日子。

- **Method**: `POST`
- **Path**: `/api/patterns/scan`
- **Content-Type**: `application/json`

#### 请求参数

| 字段       | 类型     | 必填 | 说明                                          |
|------------|----------|------|-----------------------------------------------|
| strategies | string[] | 是   | 策略名称列表，如 `["volume_surge", "kdj_oversold", "ma60_trend"]` |
| min_score  | float64  | 否   | 最低匹配分数，默认 `70`                       |
| codes      | string[] | 否   | 指定股票代码列表，不传则扫描全市场            |

#### 请求示例

```bash
# 三策略交集：放量回调 + KDJ超卖 + MA60向上
curl -X POST http://localhost:8080/api/patterns/scan \
  -H "Content-Type: application/json" \
  -d '{"strategies": ["volume_surge", "kdj_oversold", "ma60_trend"], "min_score": 70}'

# 单策略：只看 KDJ 超卖
curl -X POST http://localhost:8080/api/patterns/scan \
  -H "Content-Type: application/json" \
  -d '{"strategies": ["kdj_oversold"]}'
```

#### 支持的策略

| 策略名称           | 说明                           |
|--------------------|--------------------------------|
| `volume_surge`     | 放量上涨后缩量回调（四维评分） |
| `kdj_oversold`     | KDJ J 值低于阈值（默认 10）    |
| `ma60_trend`       | MA60 均线向上                  |
| `macd_divergence`  | MACD 底背离（骨架）            |

---

### 4. 多策略组合回测

对指定股票使用多个策略进行历史回测。所有策略都满足的日子视为买入日，
持有 N 天后卖出，计算收益表现。

- **Method**: `GET`
- **Path**: `/api/patterns/backtest`

#### 查询参数

| 字段       | 类型   | 必填 | 说明                                              |
|------------|--------|------|---------------------------------------------------|
| code       | string | 是   | 股票代码，如 `600312`                             |
| strategies | string | 是   | 逗号分隔的策略名，如 `volume_surge,kdj_oversold,ma60_trend` |
| hold_days  | int    | 否   | 持有天数，默认 `5`                                |

#### 请求示例

```bash
curl "http://localhost:8080/api/patterns/backtest?code=600312&strategies=volume_surge,kdj_oversold,ma60_trend&hold_days=5"
```

#### 成功响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "strategies": ["volume_surge", "kdj_oversold", "ma60_trend"],
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

| 字段         | 类型          | 说明                                    |
|--------------|---------------|-----------------------------------------|
| strategies   | string[]      | 参与回测的策略列表                      |
| total_trades | int           | 买入次数                                |
| win_rate     | float64       | 胜率（0~1）                             |
| avg_return   | float64       | 平均收益率（百分比）                    |
| max_drawdown | float64       | 最大回撤（百分比）                      |
| profit_factor| float64       | 盈亏比（盈利总额/亏损总额）             |
| trades       | []TradeRecord | 每笔交易明细                            |

**TradeRecord 字段**

| 字段       | 类型    | 说明                |
|------------|---------|---------------------|
| entry_date | string  | 买入日期            |
| exit_date  | string  | 卖出日期            |
| return_pct | float64 | 收益率（百分比）    |
