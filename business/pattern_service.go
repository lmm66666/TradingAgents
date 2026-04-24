package business

import (
	"context"
	"fmt"
	"log"

	"trading/data"
	"trading/model"
	"trading/pkg/strategy"
	"trading/pkg/utils"
)

// TradeRecord 单次交易记录
type TradeRecord struct {
	EntryDate string  `json:"entry_date"`
	ExitDate  string  `json:"exit_date"`
	ReturnPct float64 `json:"return_pct"`
}

// BacktestReport 回测报告
type BacktestReport struct {
	StrategyName string        `json:"strategy_name"`
	TotalTrades  int           `json:"total_trades"`
	WinRate      float64       `json:"win_rate"`
	AvgReturn    float64       `json:"avg_return"`
	MaxDrawdown  float64       `json:"max_drawdown"`
	ProfitFactor float64       `json:"profit_factor"`
	Trades       []TradeRecord `json:"trades"`
}

// PatternService 策略扫描服务
type PatternService interface {
	Scan(ctx context.Context, code string, st strategy.Strategy) ([]strategy.Signal, error)
	ScanAll(ctx context.Context, st strategy.Strategy, minScore float64) ([]strategy.Signal, error)
	Backtest(ctx context.Context, code string, st strategy.Strategy, holdDays int) (*BacktestReport, error)
}

type patternService struct {
	dailyRepo data.StockKlineDailyRepo
}

// NewPatternService 创建 PatternService
func NewPatternService(dailyRepo data.StockKlineDailyRepo) PatternService {
	return &patternService{dailyRepo: dailyRepo}
}

func (p *patternService) Scan(ctx context.Context, code string, st strategy.Strategy) ([]strategy.Signal, error) {
	dailies, err := p.dailyRepo.FindByCode(ctx, code, 0)
	if err != nil {
		return nil, fmt.Errorf("find daily failed: %w", err)
	}
	klines := dailyToKlines(dailies)
	return st.Scan(klines)
}

func (p *patternService) ScanAll(ctx context.Context, st strategy.Strategy, minScore float64) ([]strategy.Signal, error) {
	codes, err := p.dailyRepo.FindAllCodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("find all codes failed: %w", err)
	}

	var result []strategy.Signal
	for _, code := range codes {
		signals, err := p.Scan(ctx, code, st)
		if err != nil {
			log.Printf("[pattern] scan %s failed: %v", code, err)
			continue
		}
		for _, s := range signals {
			if s.Score >= minScore {
				result = append(result, s)
			}
		}
	}
	return result, nil
}

func (p *patternService) Backtest(ctx context.Context, code string, st strategy.Strategy, holdDays int) (*BacktestReport, error) {
	dailies, err := p.dailyRepo.FindByCode(ctx, code, 0)
	if err != nil {
		return nil, fmt.Errorf("find daily failed: %w", err)
	}
	klines := dailyToKlines(dailies)
	signals, err := st.Scan(klines)
	if err != nil {
		return nil, err
	}

	report := &BacktestReport{
		StrategyName: st.Name(),
		Trades:       make([]TradeRecord, 0),
	}

	for _, sig := range signals {
		entryIdx := findDateIndex(klines, sig.Date)
		if entryIdx < 0 || entryIdx+holdDays >= len(klines) {
			continue
		}
		entryPrice := klines[entryIdx].Close
		exitPrice := klines[entryIdx+holdDays].Close
		ret := (exitPrice - entryPrice) / entryPrice

		report.Trades = append(report.Trades, TradeRecord{
			EntryDate: sig.Date,
			ExitDate:  klines[entryIdx+holdDays].Date,
			ReturnPct: utils.Round4(ret * 100),
		})
	}

	report.TotalTrades = len(report.Trades)
	if report.TotalTrades > 0 {
		wins := 0
		var totalRet, grossProfit, grossLoss float64
		maxDD := 0.0
		peak := 0.0
		cumRet := 0.0
		for _, tr := range report.Trades {
			totalRet += tr.ReturnPct
			if tr.ReturnPct > 0 {
				wins++
				grossProfit += tr.ReturnPct
			} else {
				grossLoss += -tr.ReturnPct
			}
			cumRet += tr.ReturnPct
			if cumRet > peak {
				peak = cumRet
			}
			dd := peak - cumRet
			if dd > maxDD {
				maxDD = dd
			}
		}
		report.WinRate = float64(wins) / float64(report.TotalTrades)
		report.AvgReturn = totalRet / float64(report.TotalTrades)
		report.MaxDrawdown = maxDD
		if grossLoss > 0 {
			report.ProfitFactor = grossProfit / grossLoss
		}
	}

	return report, nil
}

func findDateIndex(klines []*model.StockKline, date string) int {
	for i, k := range klines {
		if k.Date == date {
			return i
		}
	}
	return -1
}
