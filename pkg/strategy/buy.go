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
