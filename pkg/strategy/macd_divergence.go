package strategy

import "trading/model"

// MACDDivergence MACD背离策略
type MACDDivergence struct{}

func (m *MACDDivergence) Name() string        { return StrategyMACDDivergence }
func (m *MACDDivergence) Description() string { return "识别MACD底背离形态" }

func (m *MACDDivergence) Scan(klines []*model.StockKline) ([]Signal, error) {
	// TODO: implement MACD divergence detection
	return nil, nil
}
