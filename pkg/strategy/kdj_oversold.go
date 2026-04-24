package strategy

import (
	"math"

	"trading/model"
	"trading/pkg/indicator"
)

// KDJOverSoldConfig KDJ 超卖策略配置
type KDJOverSoldConfig struct {
	Threshold float64 // J 值阈值，低于此值视为超卖
}

// DefaultKDJOverSoldConfig 返回默认配置
func DefaultKDJOverSoldConfig() KDJOverSoldConfig {
	return KDJOverSoldConfig{Threshold: 10.0}
}

// KDJOverSold KDJ 超卖策略：J 值低于阈值的天
type KDJOverSold struct {
	Config KDJOverSoldConfig
}

// NewKDJOverSold 创建策略实例
func NewKDJOverSold(cfg KDJOverSoldConfig) *KDJOverSold {
	return &KDJOverSold{Config: cfg}
}

func (k *KDJOverSold) Name() string        { return StrategyKDJOverSold }
func (k *KDJOverSold) Description() string { return "KDJ 超卖：J 值低于阈值" }

// Scan 返回所有 KDJ J 值低于阈值的日子
func (k *KDJOverSold) Scan(klines []*model.StockKline) ([]Signal, error) {
	if len(klines) == 0 {
		return nil, nil
	}

	kdjResults := indicator.ComputeKDJ(klines)
	threshold := k.Config.Threshold

	var signals []Signal
	for i, r := range kdjResults {
		if r.J >= threshold {
			continue
		}

		score := 0.0
		if r.J <= 0 {
			score = 100.0
		} else {
			score = (1.0 - r.J/threshold) * 100
		}

		signals = append(signals, Signal{
			Code:     klines[i].Code,
			Date:     klines[i].Date,
			Strategy: k.Name(),
			Type:     SignalWatch,
			Phase:    "oversold",
			Score:    math.Round(score*100) / 100,
			SubScores: map[string]float64{
				"kdj_j": math.Round(r.J*100) / 100,
			},
			Context: map[string]interface{}{
				"kdj_k": math.Round(r.K*100) / 100,
				"kdj_d": math.Round(r.D*100) / 100,
			},
		})
	}

	return signals, nil
}
