package score

import (
	"fmt"

	"trading/model"
	"trading/pkg/indicator"
)

const (
	overBuyScorer  = "OverBuy"
	overSoldScorer = "OverSold"
)

// KDJScorer KDJ 超买/超卖过滤器
type KDJScorer struct {
	Type      string
	Threshold float64 // J 值阈值
}

func NewOverBuyKDJScorer(threshold float64) *KDJScorer {
	return &KDJScorer{
		Type:      overBuyScorer,
		Threshold: threshold,
	}
}

func NewOverSoldKDJScorer(threshold float64) *KDJScorer {
	return &KDJScorer{
		Type:      overSoldScorer,
		Threshold: threshold,
	}
}

func (f *KDJScorer) Name() string {
	return fmt.Sprintf("%s %s", f.Type, "KDJScorer")
}

func (f *KDJScorer) Score(klines []*model.StockKline) []*Result {
	if len(klines) == 0 {
		return nil
	}

	kdjResults := indicator.ComputeKDJ(klines)

	var results []*Result
	for _, r := range kdjResults {
		var diff float64
		switch f.Type {
		case overBuyScorer:
			if r.J <= f.Threshold {
				continue
			}
			diff = r.J - f.Threshold
		case overSoldScorer:
			if r.J >= f.Threshold {
				continue
			}
			diff = f.Threshold - r.J
		}

		score := 50 + int(diff*5)
		if score > 100 {
			score = 100
		}
		results = append(results, &Result{
			Date:  r.Date,
			Score: score,
		})
	}

	return results
}
