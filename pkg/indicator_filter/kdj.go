package indicator_filter

import (
	"trading/model"
	"trading/pkg/indicator"
)

const (
	overBuyFilter  = "OverBuy"
	overSoldFilter = "OverSold"
)

// KDJFilter KDJ 超买/超卖过滤器
type KDJFilter struct {
	Type      string
	Threshold float64 // J 值阈值
}

func NewOverBuyKDJFilter(threshold float64) *KDJFilter {
	return &KDJFilter{
		Type:      overBuyFilter,
		Threshold: threshold,
	}
}

func NewOverSoldKDJFilter(threshold float64) *KDJFilter {
	return &KDJFilter{
		Type:      overSoldFilter,
		Threshold: threshold,
	}
}

func (f *KDJFilter) Filter(klines []*model.StockKline) []string {
	if len(klines) == 0 {
		return nil
	}

	kdjResults := indicator.ComputeKDJ(klines)

	var dates []string
	for i, r := range kdjResults {
		match := false
		switch f.Type {
		case overBuyFilter:
			match = r.J > f.Threshold
		case overSoldFilter:
			match = r.J < f.Threshold
		}

		if match {
			dates = append(dates, klines[i].Date)
		}
	}

	return dates
}
