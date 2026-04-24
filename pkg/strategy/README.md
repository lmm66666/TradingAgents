# Strategy 规范

## 通用规定

1. **所有策略必须实现 `Strategy` 接口**，包含以下方法：
   - `Name() string` — 策略唯一标识
   - `Description() string` — 策略描述
   - `Scan(klines []*model.StockKline) ([]Signal, error)` — 扫描 K 线返回信号

2. **信号类型**由 `SignalType` 枚举定义：
   - `buy` — 买入信号
   - `sell` — 卖出信号
   - `watch` — 观察信号

3. **策略文件命名**以策略功能命名，下划线连接：
   - `volume_surge_pullback.go` — 放量上涨缩量回调
   - `macd_divergence.go` — MACD 背离

4. **新增策略**步骤：
   - 在 `pkg/strategy/` 下新建文件
   - 实现 `Strategy` 接口
   - 在 `strategy.go` 的 `ResolveStrategy` 中注册
