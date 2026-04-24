package filter

import (
	"trading/model"
	"trading/pkg/indicator"
)

type MATrendFilter struct {
	Period      int
	Up          bool
	Consecutive int // 连续天数
}

func NewMATrendUp(period, consecutive int) *MATrendFilter {
	return &MATrendFilter{Period: period, Up: true, Consecutive: consecutive}
}

func NewMATrendDown(period, consecutive int) *MATrendFilter {
	return &MATrendFilter{Period: period, Up: false, Consecutive: consecutive}
}

func (f *MATrendFilter) WithConsecutive(n int) *MATrendFilter {
	f.Consecutive = n
	return f
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

	streak := 0
	for i := 1; i < len(maResults); i++ {
		if (f.Up && maResults[i] > maResults[i-1]) || (!f.Up && maResults[i] < maResults[i-1]) {
			streak++
		} else {
			streak = 0
		}
		results[i] = Result{
			Date:  klines[i].Date,
			Valid: streak >= f.Consecutive,
		}
	}
	return results
}
