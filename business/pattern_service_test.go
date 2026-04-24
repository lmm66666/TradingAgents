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
	klines := make([]*model.StockKlineDaily, 70)
	for i := 0; i < 70; i++ {
		price := 10.0 + float64(i)*0.01
		klines[i] = &model.StockKlineDaily{
			Code:   "600312",
			Date:   fmt.Sprintf("2026-%02d-%02d", (i/30)+1, (i%30)+1),
			Open:   price - 0.05,
			High:   price + 0.1,
			Low:    price - 0.1,
			Close:  price,
			Volume: 100000,
		}
	}
	// 放量上涨
	klines[30] = &model.StockKlineDaily{
		Code: "600312", Date: "2026-02-01",
		Open: 10.3, High: 10.7, Low: 10.2, Close: 10.6, Volume: 350000,
	}
	klines[31] = &model.StockKlineDaily{
		Code: "600312", Date: "2026-02-02",
		Open: 10.6, High: 11.0, Low: 10.5, Close: 10.9, Volume: 280000,
	}
	// 急跌压低 KDJ
	for i := 32; i <= 46; i++ {
		prevClose := klines[i-1].Close
		close := prevClose - 0.12
		klines[i] = &model.StockKlineDaily{
			Code: "600312", Date: fmt.Sprintf("2026-02-%02d", i-29),
			Open: prevClose, High: prevClose + 0.02, Low: close - 0.02, Close: close, Volume: 60000,
		}
	}
	// 缓慢回升保持 MA60 向上
	for i := 47; i < 70; i++ {
		prevClose := klines[i-1].Close
		close := prevClose + 0.03
		klines[i] = &model.StockKlineDaily{
			Code: "600312", Date: fmt.Sprintf("2026-03-%02d", i-46),
			Open: prevClose, High: close + 0.05, Low: prevClose - 0.05, Close: close, Volume: 100000,
		}
	}
	return klines
}

func testConfig() strategy.VolumeSurgePullbackConfig {
	cfg := strategy.DefaultVolumeSurgePullbackConfig()
	cfg.MaxPullbackPct = 20.0
	cfg.MaxPullbackDays = 15
	return cfg
}

func TestPatternServiceScan(t *testing.T) {
	repo := &mockDailyRepoForPattern{klines: buildDailyKlines()}
	svc := NewPatternService(repo)
	st := strategy.NewVolumeSurgePullback(testConfig())

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
	st := strategy.NewVolumeSurgePullback(testConfig())

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
	st := strategy.NewVolumeSurgePullback(testConfig())

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
