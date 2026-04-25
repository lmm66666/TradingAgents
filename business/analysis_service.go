package business

import (
	"context"
	"fmt"

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
	// FindBuySignalsByStrategy 按策略名称扫描，只返回该策略的结果
	FindBuySignalsByStrategy(ctx context.Context, name string) (*StrategySignal, error)
}

type analysisService struct {
	dailyRepo  data.StockKlineDailyRepo
	weeklyRepo data.StockKlineWeeklyRepo
}

// NewAnalysisService 创建 AnalysisService 实例
func NewAnalysisService(dailyRepo data.StockKlineDailyRepo, weeklyRepo data.StockKlineWeeklyRepo) AnalysisService {
	return &analysisService{dailyRepo: dailyRepo, weeklyRepo: weeklyRepo}
}

func (s *analysisService) FindBuySignals(ctx context.Context) ([]StrategySignal, error) {
	var results []StrategySignal

	dailySigs, err := s.scanDailyStrategy(ctx, strategy.NewDailyB1BuyStrategy())
	if err != nil {
		return nil, err
	}
	if dailySigs != nil {
		results = append(results, *dailySigs)
	}

	weeklySigs, err := s.scanWeeklyStrategy(ctx, strategy.NewWeeklyB1BuyStrategy())
	if err != nil {
		return nil, err
	}
	if weeklySigs != nil {
		results = append(results, *weeklySigs)
	}

	return results, nil
}

func (s *analysisService) FindBuySignalsByStrategy(ctx context.Context, name string) (*StrategySignal, error) {
	switch name {
	case "daily_b1_buy":
		return s.scanDailyStrategy(ctx, strategy.NewDailyB1BuyStrategy())
	case "weekly_b1_buy":
		return s.scanWeeklyStrategy(ctx, strategy.NewWeeklyB1BuyStrategy())
	default:
		return nil, fmt.Errorf("unknown strategy: %s", name)
	}
}

func (s *analysisService) scanDailyStrategy(ctx context.Context, st *strategy.Strategy) (*StrategySignal, error) {
	codes, err := s.dailyRepo.FindAllCodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("find all daily codes failed: %w", err)
	}

	var matched []string
	for _, code := range codes {
		dailies, findErr := s.dailyRepo.FindByCode(ctx, code, 70)
		if findErr != nil || len(dailies) == 0 {
			continue
		}
		lastDate := dailies[len(dailies)-1].Date
		klines := dailyToKlines(dailies)
		signals := st.ScanAll(klines)
		if len(signals) > 0 && signals[len(signals)-1].Date == lastDate {
			matched = append(matched, code)
		}
	}

	if len(matched) == 0 {
		return nil, nil
	}
	return &StrategySignal{Name: st.Name(), Codes: matched}, nil
}

func (s *analysisService) scanWeeklyStrategy(ctx context.Context, st *strategy.Strategy) (*StrategySignal, error) {
	codes, err := s.weeklyRepo.FindAllCodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("find all weekly codes failed: %w", err)
	}

	var matched []string
	for _, code := range codes {
		weeklies, findErr := s.weeklyRepo.FindByCode(ctx, code, 70)
		if findErr != nil || len(weeklies) == 0 {
			continue
		}
		lastDate := weeklies[len(weeklies)-1].Date
		klines := weeklyToKlines(weeklies)
		signals := st.ScanAll(klines)
		if len(signals) > 0 && signals[len(signals)-1].Date == lastDate {
			matched = append(matched, code)
		}
	}

	if len(matched) == 0 {
		return nil, nil
	}
	return &StrategySignal{Name: st.Name(), Codes: matched}, nil
}

func dailyToKlines(dailies []*model.StockKlineDaily) []*model.StockKline {
	result := make([]*model.StockKline, 0, len(dailies))
	for _, d := range dailies {
		k := model.StockKline(*d)
		result = append(result, &k)
	}
	return result
}

func weeklyToKlines(weeklies []*model.StockKlineWeekly) []*model.StockKline {
	result := make([]*model.StockKline, 0, len(weeklies))
	for _, w := range weeklies {
		k := model.StockKline(*w)
		result = append(result, &k)
	}
	return result
}
