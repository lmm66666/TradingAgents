package api

import (
	"fmt"

	"trading/pkg/strategy"
)

func resolveStrategy(name string) (strategy.Strategy, error) {
	switch name {
	case strategy.StrategyVolumeSurge:
		return strategy.NewVolumeSurge(strategy.DefaultVolumeSurgeConfig()), nil
	case strategy.StrategyKDJOverSold:
		return strategy.NewKDJOverSold(strategy.DefaultKDJOverSoldConfig()), nil
	case strategy.StrategyMA60Trend:
		return strategy.NewMA60Trend(strategy.DefaultMA60TrendConfig()), nil
	case strategy.StrategyMACDDivergence:
		return strategy.NewMACDDivergence(strategy.DefaultMACDDivergenceConfig()), nil
	default:
		return nil, fmt.Errorf("unknown strategy: %s", name)
	}
}
