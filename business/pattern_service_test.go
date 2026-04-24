package business

import (
	"context"
	"fmt"
	"testing"

	"trading/model"
	"trading/pkg/strategy"
)

type mockDailyRepoForPattern struct {
	klines []*model.StockKlineDaily
}

func (m *mockDailyRepoForPattern) Create(ctx context.Context, kline *model.StockKlineDaily) error {
	return nil
}
func (m *mockDailyRepoForPattern) CreateBatch(ctx context.Context, klines []*model.StockKlineDaily) error {
	return nil
}
func (m *mockDailyRepoForPattern) Upsert(ctx context.Context, klines []*model.StockKlineDaily) error {
	return nil
}
func (m *mockDailyRepoForPattern) FindByID(ctx context.Context, id uint) (*model.StockKlineDaily, error) {
	return nil, nil
}
func (m *mockDailyRepoForPattern) FindByCode(ctx context.Context, code string, limit int) ([]*model.StockKlineDaily, error) {
	return m.klines, nil
}
func (m *mockDailyRepoForPattern) FindLatestByCode(ctx context.Context, code string) (*model.StockKlineDaily, error) {
	return nil, nil
}
func (m *mockDailyRepoForPattern) FindAllCodes(ctx context.Context) ([]string, error) {
	return []string{"600312"}, nil
}
func (m *mockDailyRepoForPattern) Update(ctx context.Context, kline *model.StockKlineDaily) error {
	return nil
}
func (m *mockDailyRepoForPattern) Delete(ctx context.Context, id uint) error {
	return nil
}
func (m *mockDailyRepoForPattern) List(ctx context.Context, limit, offset int) ([]*model.StockKlineDaily, error) {
	return nil, nil
}

func buildDailyKlines() []*model.StockKlineDaily {
	klines := make([]*model.StockKlineDaily, 25)
	for i := 0; i < 25; i++ {
		vol := int64(100000)
		close := 10.0
		open := 9.9
		high := 10.1
		low := 9.8

		switch i {
		case 20:
			vol = 300000
			open = 10.2
			close = 10.5
			high = 10.6
			low = 10.1
		case 21:
			vol = 250000
			open = 10.5
			close = 10.8
			high = 10.9
			low = 10.4
		case 22, 23, 24:
			vol = 80000
			open = close
			close = 10.8 - float64(i-21)*0.15
			high = open + 0.1
			low = close - 0.1
		}

		klines[i] = &model.StockKlineDaily{
			Code:   "600312",
			Date:   fmt.Sprintf("2026-03-%02d", i+1),
			Open:   open,
			High:   high,
			Low:    low,
			Close:  close,
			Volume: vol,
		}
	}
	return klines
}

func TestPatternServiceScan(t *testing.T) {
	repo := &mockDailyRepoForPattern{klines: buildDailyKlines()}
	svc := NewPatternService(repo)
	st := strategy.NewVolumeSurgePullback(strategy.DefaultVolumeSurgePullbackConfig())

	signals, err := svc.Scan(context.Background(), "600312", st)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(signals) == 0 {
		t.Fatal("expected signals")
	}
}

func TestPatternServiceScanAll(t *testing.T) {
	repo := &mockDailyRepoForPattern{klines: buildDailyKlines()}
	svc := NewPatternService(repo)
	st := strategy.NewVolumeSurgePullback(strategy.DefaultVolumeSurgePullbackConfig())

	signals, err := svc.ScanAll(context.Background(), st, 70)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(signals) == 0 {
		t.Fatal("expected signals from scan all")
	}
}

func TestPatternServiceBacktest(t *testing.T) {
	repo := &mockDailyRepoForPattern{klines: buildDailyKlines()}
	svc := NewPatternService(repo)
	st := strategy.NewVolumeSurgePullback(strategy.DefaultVolumeSurgePullbackConfig())

	report, err := svc.Backtest(context.Background(), "600312", st, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report == nil {
		t.Fatal("expected report")
	}
	if report.StrategyName != strategy.StrategyVolumeSurgePullback {
		t.Fatalf("unexpected strategy name: %s", report.StrategyName)
	}
	if report.TotalTrades == 0 {
		t.Fatal("expected at least one trade")
	}
}
