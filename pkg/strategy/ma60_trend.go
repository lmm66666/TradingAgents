package strategy

import (
	"trading/model"
	"trading/pkg/utils"
)

// MA60Trend MA60 趋势策略：均线向上的天
type MA60Trend struct{}

func (m *MA60Trend) Name() string        { return StrategyMA60Trend }
func (m *MA60Trend) Description() string { return "MA60 均线向上" }

// Scan 返回所有 MA60 向上（当日 > 昨日）的日子
func (m *MA60Trend) Scan(klines []*model.StockKline) ([]Signal, error) {
	if len(klines) < 2 {
		return nil, nil
	}

	closes := make([]float64, len(klines))
	for i, k := range klines {
		closes[i] = k.Close
	}

	ma60 := utils.ComputeMA(closes, []int{60})[60]

	var signals []Signal
	for i := 1; i < len(klines); i++ {
		if ma60[i] <= ma60[i-1] {
			continue
		}

		signals = append(signals, Signal{
			Code:     klines[i].Code,
			Date:     klines[i].Date,
			Strategy: m.Name(),
			Type:     SignalWatch,
			Phase:    "trend_up",
			Score:    100,
			SubScores: map[string]float64{
				"ma60": utils.Round4(ma60[i]),
			},
		})
	}

	return signals, nil
}
