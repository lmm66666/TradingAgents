package filter

import (
	"trading/model"
	"trading/pkg/indicator"
)

type MATrendFilter struct {
	Period int
	Up     bool
}

func NewMATrendUp(period int) *MATrendFilter {
	return &MATrendFilter{Period: period, Up: true}
}

func NewMATrendDown(period int) *MATrendFilter {
	return &MATrendFilter{Period: period, Up: false}
}

func (f *MATrendFilter) Filter(klines []*model.StockKline) []Result {
	if len(klines) == 0 {
		return nil
	}

	prices := make([]float64, len(klines))
	for i, k := range klines {
		prices[i] = k.Close
	}

	maResults := indicator.ComputeMA(prices, f.Period)
	results := make([]Result, len(klines))
	for i := range maResults {
		var valid bool
		if i > 0 {
			if f.Up {
				valid = maResults[i] > maResults[i-1]
			} else {
				valid = maResults[i] < maResults[i-1]
			}
		}
		results[i] = Result{Date: klines[i].Date, Valid: valid}
	}
	return results
}
