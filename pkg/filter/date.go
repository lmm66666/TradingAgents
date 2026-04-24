package filter

import "trading/model"

// DateFilter 持有天数过滤器
// 输入一段 K 线，索引 >= HoldDays 的天返回 true，否则返回 false
// 常用于与其他条件组合，表示“已持有至少 N 天”
type DateFilter struct {
	HoldDays int
}

func NewDateFilter(holdDays int) *DateFilter {
	return &DateFilter{HoldDays: holdDays}
}

func (f *DateFilter) Filter(klines []*model.StockKline) []Result {
	if len(klines) == 0 {
		return nil
	}

	results := make([]Result, len(klines))
	for i, k := range klines {
		results[i] = Result{
			Date:  k.Date,
			Valid: i >= f.HoldDays,
		}
	}
	return results
}
