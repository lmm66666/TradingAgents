package backtest

import (
	"context"
	"fmt"
	"log"

	"trading/data"
	"trading/model"
	"trading/pkg/indicator"
	"trading/pkg/strategy"
)

// TradeRecord 单次交易记录
type TradeRecord struct {
	EntryDate string  `json:"entry_date"`
	ExitDate  string  `json:"exit_date"`
	ReturnPct float64 `json:"return_pct"`
}

// BacktestReport 回测报告
type BacktestReport struct {
	Strategies   []string      `json:"strategies"`
	TotalTrades  int           `json:"total_trades"`
	WinRate      float64       `json:"win_rate"`
	AvgReturn    float64       `json:"avg_return"`
	MaxDrawdown  float64       `json:"max_drawdown"`
	ProfitFactor float64       `json:"profit_factor"`
	Trades       []TradeRecord `json:"trades"`
}

// BacktestService 多策略组合扫描与回测
type BacktestService interface {
	// Scan 多策略交集扫描：所有策略都有信号的日子
	Scan(ctx context.Context, code string, strategies []strategy.Strategy, minScore float64) ([]strategy.Signal, error)
	// ScanAll 全市场多策略交集扫描
	ScanAll(ctx context.Context, strategies []strategy.Strategy, minScore float64) ([]strategy.Signal, error)
	// Run 组合回测：交集买入日 + holdDays 持仓
	Run(ctx context.Context, code string, strategies []strategy.Strategy, holdDays int) (*BacktestReport, error)
}

type backtestService struct {
	dailyRepo data.StockKlineDailyRepo
}

// NewBacktestService 创建 BacktestService
func NewBacktestService(dailyRepo data.StockKlineDailyRepo) BacktestService {
	return &backtestService{dailyRepo: dailyRepo}
}

func (s *backtestService) Scan(ctx context.Context, code string, strs []strategy.Strategy, minScore float64) ([]strategy.Signal, error) {
	dailies, err := s.dailyRepo.FindByCode(ctx, code, 0)
	if err != nil {
		return nil, fmt.Errorf("find daily failed: %w", err)
	}
	klines := dailyToKlines(dailies)
	return intersectSignals(klines, strs, minScore), nil
}

func (s *backtestService) ScanAll(ctx context.Context, strs []strategy.Strategy, minScore float64) ([]strategy.Signal, error) {
	codes, err := s.dailyRepo.FindAllCodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("find all codes failed: %w", err)
	}

	var result []strategy.Signal
	for _, code := range codes {
		dailies, findErr := s.dailyRepo.FindByCode(ctx, code, 0)
		if findErr != nil {
			log.Printf("[backtest] scan %s failed: %v", code, findErr)
			continue
		}
		klines := dailyToKlines(dailies)
		signals := intersectSignals(klines, strs, minScore)
		result = append(result, signals...)
	}
	return result, nil
}

func (s *backtestService) Run(ctx context.Context, code string, strs []strategy.Strategy, holdDays int) (*BacktestReport, error) {
	dailies, err := s.dailyRepo.FindByCode(ctx, code, 0)
	if err != nil {
		return nil, fmt.Errorf("find daily failed: %w", err)
	}
	klines := dailyToKlines(dailies)

	signals := intersectSignals(klines, strs, 0)

	names := make([]string, len(strs))
	for i, st := range strs {
		names[i] = st.Name()
	}

	report := &BacktestReport{
		Strategies: names,
		Trades:     make([]TradeRecord, 0),
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
			ReturnPct: indicator.Round4(ret * 100),
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

// intersectSignals 对多个策略的信号按日期取交集
func intersectSignals(klines []*model.StockKline, strs []strategy.Strategy, minScore float64) []strategy.Signal {
	n := len(strs)
	if n == 0 {
		return nil
	}

	// 运行每个策略
	allSignals := make([][]strategy.Signal, n)
	for i, st := range strs {
		signals, err := st.Scan(klines)
		if err != nil {
			log.Printf("[backtest] strategy %s scan failed: %v", st.Name(), err)
			continue
		}
		allSignals[i] = signals
	}

	// 第一个策略为空则直接返回
	if len(allSignals[0]) == 0 {
		return nil
	}

	// 用第一个策略的日期建立索引
	firstByDate := make(map[string]strategy.Signal, len(allSignals[0]))
	for _, s := range allSignals[0] {
		firstByDate[s.Date] = s
	}

	// 与其他策略取交集
	intersection := firstByDate
	for i := 1; i < n; i++ {
		nextByDate := make(map[string]strategy.Signal, len(allSignals[i]))
		for _, s := range allSignals[i] {
			nextByDate[s.Date] = s
		}

		newIntersection := make(map[string]strategy.Signal)
		for date, sig := range intersection {
			if otherSig, ok := nextByDate[date]; ok {
				// 取较低的 score
				if sig.Score <= otherSig.Score {
					newIntersection[date] = sig
				} else {
					newIntersection[date] = otherSig
				}
			}
		}
		intersection = newIntersection
	}

	// 收集结果，过滤 minScore
	var result []strategy.Signal
	for _, sig := range intersection {
		if sig.Score < minScore {
			continue
		}
		result = append(result, sig)
	}

	return result
}

func findDateIndex(klines []*model.StockKline, date string) int {
	for i, k := range klines {
		if k.Date == date {
			return i
		}
	}
	return -1
}

func dailyToKlines(dailies []*model.StockKlineDaily) []*model.StockKline {
	result := make([]*model.StockKline, 0, len(dailies))
	for _, d := range dailies {
		k := model.StockKline(*d)
		result = append(result, &k)
	}
	return result
}
