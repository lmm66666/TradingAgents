package indicator_filter

import (
	"trading/model"
	"trading/pkg/indicator"
)

const (
	trendUpFilter   = "TrendUp"
	trendDownFilter = "TrendDown"
)

type MAFilter struct {
	Type   string
	Period int
}

func NewTrendUpMAFilter(period int) *MAFilter {
	return &MAFilter{
		Type:   trendUpFilter,
		Period: period,
	}
}

func NewTrendDownMAFilter(period int) *MAFilter {
	return &MAFilter{
		Type:   trendDownFilter,
		Period: period,
	}
}

func (f *MAFilter) Filter(klines []*model.StockKline) []string {
	if len(klines) == 0 {
		return nil
	}

	var prices []float64
	for _, kline := range klines {
		prices = append(prices, kline.Close)
	}

	maResults := indicator.ComputeMA(prices, f.Period)

	var dates []string
	for i := range maResults {
		if i == 0 {
			continue
		}

		match := false
		switch f.Type {
		case trendUpFilter:
			match = maResults[i] > maResults[i-1]
		case trendDownFilter:
			match = maResults[i] < maResults[i-1]
		}

		if match {
			dates = append(dates, klines[i].Date)
		}
	}

	return dates
}
