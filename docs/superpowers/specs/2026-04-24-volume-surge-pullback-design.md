# 放量上涨 + 缩量回调 策略设计文档

## 1. 背景与目标

在现有股票数据分析系统的基础上，新增**策略框架层**，支持可扩展的技术分析策略。首期实现 **VolumeSurgePullback（放量上涨后缩量回调）** 策略，同时预留 MACD 背离、KDJ 金叉等策略的接入能力。

目标：
- 实时选股：扫描全市场，找出当前处于缩量回调阶段的股票
- 历史回测：验证策略在历史数据上的胜率与收益表现
- 可扩展：后续策略只需实现统一接口即可无缝接入

## 2. 核心概念定义

### 2.1 模式结构

对单只股票的 K 线序列，识别三段式结构：

```
[放量上涨日] → [可选续涨] → [缩量回调期]
     ↑              ↑            ↑
   成交量突增    价格持续新高    成交量萎缩
   价格大涨                      价格温和回落
```

### 2.2 判定约束

| 阶段 | 约束条件 |
|------|----------|
| 放量日 | 成交量 >= 20 日均量 × 1.2，涨幅 >= 2% |
| 回调期 | 从放量后最高点算起，成交量 <= 放量日 × 0.7，回调幅度 <= 15% |
| 时间窗 | 回调持续 2~8 个交易日 |

## 3. 策略框架接口设计

### 3.1 统一接口

```go
package strategy

// SignalType 信号类型
const (
    SignalBuy   SignalType = "buy"
    SignalSell  SignalType = "sell"
    SignalWatch SignalType = "watch"
)

// Signal 策略产生的单个信号
type Signal struct {
    Code      string                 `json:"code"`
    Date      string                 `json:"date"`         // 信号产生日期
    Strategy  string                 `json:"strategy"`     // 策略名称
    Type      SignalType             `json:"type"`         // buy / sell / watch
    Phase     string                 `json:"phase"`        // 阶段描述
    Score     float64                `json:"score"`        // 0~100 匹配质量分
    SubScores map[string]float64     `json:"sub_scores"`   // 各维度明细
    Context   map[string]interface{} `json:"context"`      // 策略-specific 上下文
}

// Strategy 策略接口，所有具体策略必须实现
type Strategy interface {
    Name() string
    Description() string
    Scan(klines []*model.StockKline) ([]Signal, error)
}

// Configurable 可选接口，支持参数调整的策略实现
type Configurable interface {
    Strategy
    DefaultConfig() interface{}
    ValidateConfig(cfg interface{}) error
}
```

### 3.2 组合扫描器

```go
// Scanner 组合多个策略进行扫描
type Scanner struct {
    strategies []Strategy
}

func NewScanner(strategies ...Strategy) *Scanner
func (s *Scanner) AddStrategy(st Strategy)
func (s *Scanner) Scan(klines []*model.StockKline) (map[string][]Signal, error)
```

## 4. VolumeSurgePullback 策略详细设计

### 4.1 配置结构

```go
type VolumeSurgePullbackConfig struct {
    VolumeMAPeriod  int
    MinVolumeRatio  float64  // 放量阈值，默认 1.2
    MinRallyPct     float64  // 最小涨幅，默认 2.0
    MaxPullbackPct  float64  // 最大回撤，默认 15.0
    MaxPullbackDays int      // 最长回调天数，默认 8
    MinScore        float64  // 有效模式最低分，默认 70
    Weights         struct {
        VolumeSurge     float64 // 放量强度权重，默认 0.20
        PriceRally      float64 // 上涨强度权重，默认 0.15
        PullbackVolume  float64 // 缩量程度权重，默认 0.25
        PullbackDepth   float64 // 回调健康度权重，默认 0.20
        TimeRhythm      float64 // 时间节奏权重，默认 0.10
        MASupport       float64 // 均线支撑权重，默认 0.10
    }
}
```

### 4.2 评分模型（0~100 分）

| 维度 | 计算方法 | 权重 |
|------|----------|------|
| **放量强度** | `min(当日量/20日均量, 3) / 3 * 100` | 20% |
| **上涨强度** | `min(涨幅%, 8) / 8 * 100` | 15% |
| **缩量程度** | `(1 - 回调期均量/放量日量) * 100`，最低 0 | 25% |
| **回调健康度** | `(1 - 回撤幅度/15%) * 100` | 20% |
| **时间节奏** | 回调 2~5 天 = 100 分，6~8 天 = 60 分，其他 = 0 | 10% |
| **均线支撑** | 回调低点 >= MA20 得 100 分，>= MA60 得 50 分 | 10% |

总分 >= 70 视为有效模式。

### 4.3 状态机

```
none → surge → rally → pullback → complete
                        ↓
                     invalid
```

| 状态 | 说明 |
|------|------|
| `none` | 无模式 |
| `surge` | 刚出现放量（1~2 天内） |
| `rally` | 放量后继续上涨 |
| `pullback` | 正在缩量回调（买入观察窗口） |
| `complete` | 回调结束，放量突破回调前高 |
| `invalid` | 回调过深或超时，模式失效 |

### 4.4 扫描算法流程

1. **预处理**：计算成交量 MA、价格 MA
2. **寻找放量日**：遍历 K 线，找满足放量条件的日期
3. **以放量日为锚点，向后追踪**：
   - 若价格继续创新高，更新高点，进入 `rally` 状态
   - 若出现回调，进入 `pullback` 状态
4. **计算回调期评分**：持续跟踪回调天数、成交量、回撤幅度
5. **状态判定**：
   - 回调 <= 15% 且天数 <= 8 天：保持 `pullback`
   - 放量突破回调前高：`complete`
   - 回调 > 15% 或天数 > 8 天：`invalid`
6. **生成 Signal**：对每个有效模式生成 Signal，包含当前阶段和评分

## 5. 业务层设计

### 5.1 PatternService 接口

```go
type PatternService interface {
    Scan(ctx context.Context, code string, st strategy.Strategy) ([]strategy.Signal, error)
    ScanAll(ctx context.Context, st strategy.Strategy, minScore float64) ([]strategy.Signal, error)
    MultiScan(ctx context.Context, codes []string, scanner *strategy.Scanner) (map[string]map[string][]strategy.Signal, error)
    Backtest(ctx context.Context, code string, st strategy.Strategy, holdDays int) (*BacktestReport, error)
}

type BacktestReport struct {
    StrategyName string
    TotalTrades  int
    WinRate      float64
    AvgReturn    float64
    MaxDrawdown  float64
    ProfitFactor float64
    Trades       []TradeRecord
}
```

### 5.2 实时选股流程

1. 从 DB 获取所有股票最新 60 天日线数据
2. 对每只股票调用 `Strategy.Scan`
3. 过滤 `Phase == "pullback"` 且 `Score >= 70` 的结果
4. 按 `Score` 降序排列，取 Top N
5. 输出：股票代码、放量日、当前回调天数、得分、均线支撑情况

### 5.3 历史回测流程

1. 对单只股票，取全部历史 K 线
2. 扫描所有 `PatternMatch`（不仅看最新的）
3. 对每个匹配，计算"模式结束后 holdDays 天的收益率"
4. 汇总统计：胜率、平均收益、最大回撤、盈亏比
5. 调整 `PatternConfig` 参数，重复步骤 2~4，寻找最优参数组合

## 6. API 设计

### 6.1 实时选股

```http
POST /api/patterns/scan
Content-Type: application/json

{
  "strategy": "volume_surge_pullback",
  "min_score": 70,
  "codes": ["600312", "000001"]  // 可选，不传则扫描全市场
}
```

响应：
```json
{
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
        "pullback_days": 4,
        "max_pullback_pct": 6.5
      }
    }
  ]
}
```

### 6.2 历史回测

```http
GET /api/patterns/backtest?code=600312&strategy=volume_surge_pullback&hold_days=5
```

响应：
```json
{
  "strategy_name": "volume_surge_pullback",
  "total_trades": 45,
  "win_rate": 0.62,
  "avg_return": 0.035,
  "max_drawdown": 0.12,
  "profit_factor": 1.8,
  "trades": [
    {
      "entry_date": "2026-01-15",
      "exit_date": "2026-01-22",
      "return_pct": 0.052
    }
  ]
}
```

## 7. 文件清单

| 文件 | 内容 |
|------|------|
| `pkg/strategy/strategy.go` | Signal、Strategy 接口、SignalType 枚举 |
| `pkg/strategy/scanner.go` | Scanner 组合扫描器 |
| `pkg/strategy/volume_surge_pullback.go` | 放量上涨缩量回调策略 + 评分算法 |
| `pkg/strategy/volume_surge_pullback_test.go` | 单元测试 |
| `pkg/strategy/macd_divergence.go` | 预留：MACD 背离策略骨架 |
| `business/pattern_service.go` | PatternService 接口与实现（策略无关） |
| `business/pattern_service_test.go` | 服务层测试 |
| `api/scan_patterns.go` | `POST /api/patterns/scan` 接口 |
| `api/backtest_patterns.go` | `GET /api/patterns/backtest` 接口 |

## 8. 扩展预留

后续策略接入步骤：

1. 在 `pkg/strategy/` 下新建文件
2. 实现 `Strategy` 接口（`Name`、`Description`、`Scan`）
3. 如需参数配置，额外实现 `Configurable` 接口
4. 在 API 层通过策略名称字符串匹配到具体实现

预留策略：
- `macd_divergence`：MACD 底背离识别
- `kdj_cross`：KDJ 金叉/死叉识别
- `breakout`：平台突破识别
