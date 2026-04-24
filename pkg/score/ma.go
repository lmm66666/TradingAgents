package score

import (
	"trading/model"
	"trading/pkg/indicator"
)

const (
	trendUpScorer   = "TrendUp"
	trendDownScorer = "TrendDown"
)

type MAScorer struct {
	Type   string
	Period int
}

func NewTrendUpMAScorer(period int) *MAScorer {
	return &MAScorer{
		Type:   trendUpScorer,
		Period: period,
	}
}

func NewTrendDownMAScorer(period int) *MAScorer {
	return &MAScorer{
		Type:   trendDownScorer,
		Period: period,
	}
}

func (f *MAScorer) Score(klines []*model.StockKline) []*Result {
	if len(klines) == 0 {
		return nil
	}

	var prices []float64
	for _, kline := range klines {
		prices = append(prices, kline.Close)
	}

	maResults := indicator.ComputeMA(prices, f.Period)

	var results []*Result
	for i := range maResults {
		if i == 0 {
			continue
		}

		match := false
		switch f.Type {
		case trendUpScorer:
			match = maResults[i] > maResults[i-1]
		case trendDownScorer:
			match = maResults[i] < maResults[i-1]
		}

		if match {
			results = append(results, &Result{
				Date:  klines[i].Date,
				Score: 100,
			})
		}
	}

	return results
}
