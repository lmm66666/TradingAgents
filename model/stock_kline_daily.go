package model

// StockKlineDaily 日线数据模型
type StockKlineDaily StockKline

func (StockKlineDaily) TableName() string {
	return "t_stock_kline_daily"
}
