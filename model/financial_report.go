package model

import "gorm.io/gorm"

// FinancialReport 季度财报核心指标（DB 实体）
// 从新浪财报 API 的 81 个原始字段中提取最有分析价值的指标
type FinancialReport struct {
	gorm.Model
	Code       string `gorm:"size:16;index;uniqueIndex:idx_code_report_date" json:"code"`
	ReportDate string `gorm:"size:20;index;uniqueIndex:idx_code_report_date" json:"report_date"` // 如 20250930
	ReportType int    `json:"report_type"`                                                       // 1一季报 2半年报 3三季报 4年报

	// 利润表
	TotalRevenue float64 `gorm:"type:decimal(16,4)" json:"total_revenue"`  // 营业总收入
	TotalCost    float64 `gorm:"type:decimal(16,4)" json:"total_cost"`     // 营业成本
	NetProfit    float64 `gorm:"type:decimal(16,4)" json:"net_profit"`     // 归母净利润
	NetProfitCut float64 `gorm:"type:decimal(16,4)" json:"net_profit_cut"` // 扣非净利润

	// 盈利能力
	GrossMargin     float64 `gorm:"type:decimal(8,4)" json:"gross_margin"`      // 毛利率
	NetMargin       float64 `gorm:"type:decimal(8,4)" json:"net_margin"`        // 销售净利率
	OperatingMargin float64 `gorm:"type:decimal(8,4)" json:"operating_margin"`  // 营业利润率
	EBITMargin      float64 `gorm:"type:decimal(8,4)" json:"ebit_margin"`       // 息税前利润率
	CostProfitRatio float64 `gorm:"type:decimal(8,4)" json:"cost_profit_ratio"` // 成本费用利润率
	ROE             float64 `gorm:"type:decimal(8,4)" json:"roe"`               // 净资产收益率
	ROA             float64 `gorm:"type:decimal(8,4)" json:"roa"`               // 总资产报酬率

	// 偿债能力
	AssetLiabilityRatio float64 `gorm:"type:decimal(8,4)" json:"asset_liability_ratio"` // 资产负债率
	CurrentRatio        float64 `gorm:"type:decimal(8,4)" json:"current_ratio"`         // 流动比率
	QuickRatio          float64 `gorm:"type:decimal(8,4)" json:"quick_ratio"`           // 速动比率

	// 运营效率
	TotalAssetTurnover  float64 `gorm:"type:decimal(8,4)" json:"total_asset_turnover"` // 总资产周转率
	InventoryTurnover   float64 `gorm:"type:decimal(8,4)" json:"inventory_turnover"`   // 存货周转率
	ReceivablesTurnover float64 `gorm:"type:decimal(8,4)" json:"receivables_turnover"` // 应收账款周转率

	// 现金流
	OperatingCashFlow         float64 `gorm:"type:decimal(16,4)" json:"operating_cash_flow"`          // 经营现金流量净额
	OperatingCashFlowPerShare float64 `gorm:"type:decimal(8,4)" json:"operating_cash_flow_per_share"` // 每股经营现金流

	// 每股指标
	EPS float64 `gorm:"type:decimal(8,4)" json:"eps"` // 基本每股收益
	BPS float64 `gorm:"type:decimal(8,4)" json:"bps"` // 每股净资产

}
