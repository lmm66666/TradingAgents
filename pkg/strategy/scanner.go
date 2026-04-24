package strategy

import "trading/model"

// Scanner 组合多个策略进行扫描
type Scanner struct {
	strategies []Strategy
}

// NewScanner 创建 Scanner
func NewScanner(strategies ...Strategy) *Scanner {
	return &Scanner{strategies: strategies}
}

// AddStrategy 添加策略
func (s *Scanner) AddStrategy(st Strategy) {
	s.strategies = append(s.strategies, st)
}

// Scan 对所有策略执行扫描
// 返回: strategyName -> signals
func (s *Scanner) Scan(klines []*model.StockKline) (map[string][]Signal, error) {
	result := make(map[string][]Signal, len(s.strategies))
	for _, st := range s.strategies {
		signals, err := st.Scan(klines)
		if err != nil {
			return nil, err
		}
		result[st.Name()] = signals
	}
	return result, nil
}
