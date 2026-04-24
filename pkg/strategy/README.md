# Strategy 规范

## 架构原则

**每个策略 = 一个条件**。策略只负责检测自己的条件是否满足，返回符合要求的日期和评分。
多策略组合（交集）由 `pkg.backtest` 负责，不在策略层做。

## 通用规定

1. **所有策略必须实现 `Strategy` 接口**，包含以下方法：
   - `Name() string` — 策略唯一标识
   - `Description() string` — 策略描述
   - `Scan(klines []*model.StockKline) ([]Signal, error)` — 扫描 K 线返回信号

2. **信号类型**由 `SignalType` 枚举定义：
   - `buy` — 买入信号（由 BacktestService 交集后生成）
   - `sell` — 卖出信号
   - `watch` — 观察信号（策略 Scan 返回此类型）

3. **策略文件命名**以策略功能命名，下划线连接：
   - `volume_surge.go` — 放量上涨+缩量回调
   - `kdj_oversold.go` — KDJ 超卖
   - `ma60_trend.go` — MA60 向上

4. **新增策略**步骤：
   - 在 `pkg/strategy/` 下新建文件
   - 实现 `Strategy` 接口
   - 在 `strategy.go` 的常量和 `ResolveStrategy` 中注册
   - 如需可选配置，实现 `Configurable` 接口
