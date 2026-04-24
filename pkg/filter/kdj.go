package filter

import (
	"trading/model"
	"trading/pkg/indicator"
)

type KDJFilter struct {
	Type      string // OverBuy 或 OverSold
	Threshold float64
}

func NewKDJOverBuy(threshold float64) *KDJFilter {
	return &KDJFilter{Type: "OverBuy", Threshold: threshold}
}

func NewKDJOverSold(threshold float64) *KDJFilter {
	return &KDJFilter{Type: "OverSold", Threshold: threshold}
}

func (f *KDJFilter) Filter(klines []*model.StockKline) []Result {
	if len(klines) == 0 {
		return nil
	}

	kdjResults := indicator.ComputeKDJ(klines)
	results := make([]Result, len(kdjResults))
	for i, r := range kdjResults {
		var valid bool
		switch f.Type {
		case "OverBuy":
			valid = r.J > f.Threshold
		case "OverSold":
			valid = r.J < f.Threshold
		}
		results[i] = Result{Date: r.Date, Valid: valid}
	}
	return results
}
