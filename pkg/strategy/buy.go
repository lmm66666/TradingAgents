package strategy

import (
	"trading/pkg/filter"
)

func NewDailyB1BuyStrategy() *Strategy {
	return NewStrategy("daily_b1_buy").
		AddFilter(filter.NewVolumeSurgeFilter(filter.DefaultVolumeSurgeConfig())).
		AddFilter(filter.NewKDJOverSold(10)).
		AddFilter(filter.NewMATrendUp(20, 10))
}

// NewWeeklyB1BuyStrategy 周线 B1 买点策略
// 条件：周线 KDJ < 10（超卖）且周线 MA20 趋势向上
func NewWeeklyB1BuyStrategy() *Strategy {
	return NewStrategy("weekly_b1_buy").
		AddFilter(filter.NewKDJOverSold(10)).
		AddFilter(filter.NewMATrendUp(20, 1))
}
