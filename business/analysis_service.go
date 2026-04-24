package business

import (
	"context"
	"fmt"
	"time"

	"trading/data"
	"trading/model"
	"trading/pkg/strategy"
)

// StrategySignal 单个策略的扫描结果
type StrategySignal struct {
	Name  string   `json:"name"`
	Codes []string `json:"codes"`
}

// AnalysisService 股票分析服务
type AnalysisService interface {
	// FindBuySignals 扫描所有股票，返回每个策略对应的买点股票列表
	FindBuySignals(ctx context.Context) ([]StrategySignal, error)
}

type analysisService struct {
	dailyRepo data.StockKlineDailyRepo
}

// NewAnalysisService 创建 AnalysisService 实例
func NewAnalysisService(dailyRepo data.StockKlineDailyRepo) AnalysisService {
	return &analysisService{dailyRepo: dailyRepo}
}

func (s *analysisService) FindBuySignals(ctx context.Context) ([]StrategySignal, error) {
	codes, err := s.dailyRepo.FindAllCodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("find all codes failed: %w", err)
	}

	strategies := []*strategy.Strategy{
		strategy.NewDailyB1BuyStrategy(),
	}

	today := time.Now().Format("2006-01-02")
	var results []StrategySignal

	for _, st := range strategies {
		var matched []string
		for _, code := range codes {
			dailies, findErr := s.dailyRepo.FindByCode(ctx, code, 70)
			if findErr != nil {
				continue
			}

			klines := dailyToKlines(dailies)
			sig := st.Scan(klines)
			if sig != nil && sig.Date == today {
				matched = append(matched, code)
			}
		}
		if len(matched) > 0 {
			results = append(results, StrategySignal{
				Name:  st.Name(),
				Codes: matched,
			})
		}
	}

	return results, nil
}

func dailyToKlines(dailies []*model.StockKlineDaily) []*model.StockKline {
	result := make([]*model.StockKline, 0, len(dailies))
	for _, d := range dailies {
		k := model.StockKline(*d)
		result = append(result, &k)
	}
	return result
}
