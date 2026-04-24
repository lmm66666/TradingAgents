package strategy

import (
	"fmt"

	"trading/model"
)

// SignalType 信号类型
type SignalType string

const (
	SignalBuy   SignalType = "buy"
	SignalSell  SignalType = "sell"
	SignalWatch SignalType = "watch"
)

// Signal 策略产生的单个信号
type Signal struct {
	Code      string                 `json:"code"`
	Date      string                 `json:"date"`
	Strategy  string                 `json:"strategy"`
	Type      SignalType             `json:"type"`
	Phase     string                 `json:"phase"`
	Score     float64                `json:"score"`
	SubScores map[string]float64     `json:"sub_scores"`
	Context   map[string]interface{} `json:"context"`
}

// Strategy 策略接口，所有具体策略必须实现
type Strategy interface {
	Name() string
	Description() string
	Scan(klines []*model.StockKline) ([]Signal, error)
}

// Configurable 可选接口，支持参数调整的策略实现
type Configurable interface {
	Strategy
	DefaultConfig() interface{}
	ValidateConfig(cfg interface{}) error
}

// ResolveStrategy 根据策略名称创建策略实例
func ResolveStrategy(name string) (Strategy, error) {
	switch name {
	case "volume_surge_pullback":
		return NewVolumeSurgePullback(DefaultVolumeSurgePullbackConfig()), nil
	case "macd_divergence":
		return &MACDDivergence{}, nil
	default:
		return nil, fmt.Errorf("unknown strategy: %s", name)
	}
}
