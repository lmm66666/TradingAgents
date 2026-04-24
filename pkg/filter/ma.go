package filter

import (
	"trading/model"
	"trading/pkg/indicator"
)

type MATrendFilter struct {
	Period      int
	Up          bool
	Consecutive int // 连续 N 天保持趋势才算有效，默认 1
}

func NewMATrendUp(period int) *MATrendFilter {
	return &MATrendFilter{Period: period, Up: true, Consecutive: 1}
}

func NewMATrendDown(period int) *MATrendFilter {
	return &MATrendFilter{Period: period, Up: false, Consecutive: 1}
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
	for i := range maResults {
		valid := false
		if i >= f.Consecutive {
			valid = true
			for j := 0; j < f.Consecutive; j++ {
				if f.Up {
					if maResults[i-j] <= maResults[i-j-1] {
						valid = false
						break
					}
				} else {
					if maResults[i-j] >= maResults[i-j-1] {
						valid = false
						break
					}
				}
			}
		}
		results[i] = Result{Date: klines[i].Date, Valid: valid}
	}
	return results
}
