package model

// StockKlineWeekly 周线数据模型
type StockKlineWeekly StockKline

func (StockKlineWeekly) TableName() string {
	return "t_stock_kline_weekly"
}
