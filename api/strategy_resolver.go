package api

import (
	"fmt"

	"trading/pkg/indicator_filter"
	"trading/pkg/strategy"
)

func resolveStrategy(name string) (strategy.Strategy, error) {
	switch name {
	case strategy.StrategyVolumeSurge:
		return indicator_filter.NewVolumeSurge(indicator_filter.DefaultVolumeSurgeConfig()), nil
	case strategy.StrategyKDJOverSold:
		return indicator_filter.NewKDJOverSold(indicator_filter.DefaultKDJFilterConfig()), nil
	case strategy.StrategyMA60Trend:
		return indicator_filter.NewMA60Trend(indicator_filter.DefaultMA60TrendConfig()), nil
	case strategy.StrategyMACDDivergence:
		return strategy.NewMACDDivergence(strategy.DefaultMACDDivergenceConfig()), nil
	default:
		return nil, fmt.Errorf("unknown strategy: %s", name)
	}
}
