# 日线 B1 股票筛选

## 核心目标
定位专业价值投资者角色，以夏普比率最大化为核心目标，兼顾投资收益与风险控制，筛选具备长期投资价值的周线 B1 形态个股

## 流程
1. 根据 api.md 调用 /api/stocks/signal 获取处于日线 b1 的股票
2. 对于符合条件的股票，根据 api.md 调用 /api/stocks/price 获取股价信息，/api/stocks/financial-report 获取财报信息
3. 根据短线走势 + 长线发展综合进行打分。短线走势占比 60%，长线价值占比 40%。结合技术形态与财务基本面双重逻辑综合评估投资价值，标准化输出个股分级买入建议
4. 短线走势要求有一个完整的放量大幅上涨，缩量回调，不破支撑的走势
5. 长线发展从估值价格、盈利质量、所处行业、成长能力、现金流健康五大核心维度开展基本面深度分析
6. 股票名称可以从 ./shell/code 中获取

## 财报数据格式
```
type FinancialReport struct {
	gorm.Model
	Code       string `gorm:"size:16;index;uniqueIndex:idx_code_report_date" json:"code"`
	ReportDate string `gorm:"size:20;index;uniqueIndex:idx_code_report_date" json:"report_date"` // 如 20250930
	ReportType int    `json:"report_type"`                                                       // 1一季报 2半年报 3三季报 4年报

	// 利润表（decimal(20,4) 支持百万亿级别，如中石化年营收超 3 万亿）
	TotalRevenue float64 `gorm:"type:decimal(20,4)" json:"total_revenue"`  // 营业总收入
	TotalCost    float64 `gorm:"type:decimal(20,4)" json:"total_cost"`     // 营业成本
	NetProfit    float64 `gorm:"type:decimal(20,4)" json:"net_profit"`     // 归母净利润
	NetProfitCut float64 `gorm:"type:decimal(20,4)" json:"net_profit_cut"` // 扣非净利润

	// 盈利能力
	GrossMargin     float64 `gorm:"type:decimal(10,4)" json:"gross_margin"`      // 毛利率
	NetMargin       float64 `gorm:"type:decimal(10,4)" json:"net_margin"`        // 销售净利率
	OperatingMargin float64 `gorm:"type:decimal(10,4)" json:"operating_margin"`  // 营业利润率
	EBITMargin      float64 `gorm:"type:decimal(10,4)" json:"ebit_margin"`       // 息税前利润率
	CostProfitRatio float64 `gorm:"type:decimal(10,4)" json:"cost_profit_ratio"` // 成本费用利润率
	ROE             float64 `gorm:"type:decimal(10,4)" json:"roe"`               // 净资产收益率
	ROA             float64 `gorm:"type:decimal(10,4)" json:"roa"`               // 总资产报酬率

	// 偿债能力
	AssetLiabilityRatio float64 `gorm:"type:decimal(10,4)" json:"asset_liability_ratio"` // 资产负债率
	CurrentRatio        float64 `gorm:"type:decimal(10,4)" json:"current_ratio"`         // 流动比率
	QuickRatio          float64 `gorm:"type:decimal(10,4)" json:"quick_ratio"`           // 速动比率

	// 运营效率（decimal(12,4) 防止茅台等极端周转率溢出）
	TotalAssetTurnover  float64 `gorm:"type:decimal(12,4)" json:"total_asset_turnover"` // 总资产周转率
	InventoryTurnover   float64 `gorm:"type:decimal(12,4)" json:"inventory_turnover"`   // 存货周转率
	ReceivablesTurnover float64 `gorm:"type:decimal(12,4)" json:"receivables_turnover"` // 应收账款周转率

	// 现金流（decimal(20,4) 支持大型国企经营现金流）
	OperatingCashFlow         float64 `gorm:"type:decimal(20,4)" json:"operating_cash_flow"`          // 经营现金流量净额
	OperatingCashFlowPerShare float64 `gorm:"type:decimal(10,4)" json:"operating_cash_flow_per_share"` // 每股经营现金流

	// 每股指标
	EPS float64 `gorm:"type:decimal(8,4)" json:"eps"` // 基本每股收益
	BPS float64 `gorm:"type:decimal(8,4)" json:"bps"` // 每股净资产

}
```

## 分析结果要求
### 输出地址
固定放在 ./docs/analysis
### 文件名称
日期-日线b1（如：2026-04-14日线b1分析.md）
### 文件内容（固定表格格式）
| 股票名称 | 股票代码 | 打分 | 原因   |
| -------- | -------- | ---- |------|
| 贵州茅台 | 600519   | 88   | 短线走势反弹概率大，股票基本面不错 |
| 宁德时代 | 300750   | 76   |      |