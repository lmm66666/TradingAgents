package api

import (
	"fmt"

	"trading/pkg/score"
	"trading/pkg/strategy"
)

func resolveStrategy(name string) (strategy.Strategy, error) {
	switch name {
	case strategy.StrategyVolumeSurge:
		return score.NewVolumeSurge(score.DefaultVolumeSurgeConfig()), nil
	case strategy.StrategyKDJOverSold:
		return score.NewKDJOverSold(score.DefaultKDJFilterConfig()), nil
	case strategy.StrategyMA60Trend:
		return score.NewMA60Trend(score.DefaultMA60TrendConfig()), nil
	case strategy.StrategyMACDDivergence:
		return strategy.NewMACDDivergence(strategy.DefaultMACDDivergenceConfig()), nil
	default:
		return nil, fmt.Errorf("unknown strategy: %s", name)
	}
}
