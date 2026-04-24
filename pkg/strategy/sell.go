package strategy

import (
	"trading/pkg/filter"
)

func NewDefaultSellStrategy() *Strategy {
	return NewStrategy("default_sell").
		AddFilter(filter.NewDateFilter(10))
}
