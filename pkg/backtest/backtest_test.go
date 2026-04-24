package backtest

import (
	"context"
	"fmt"
	"testing"

	"trading/model"
	"trading/pkg/score"
	"trading/pkg/strategy"
)

type mockDailyRepoForBacktest struct {
	klines []*model.StockKlineDaily
}

func (m *mockDailyRepoForBacktest) Create(ctx context.Context, kline *model.StockKlineDaily) error {
	return nil
}
func (m *mockDailyRepoForBacktest) CreateBatch(ctx context.Context, klines []*model.StockKlineDaily) error {
	return nil
}
func (m *mockDailyRepoForBacktest) Upsert(ctx context.Context, klines []*model.StockKlineDaily) error {
	return nil
}
func (m *mockDailyRepoForBacktest) FindByID(ctx context.Context, id uint) (*model.StockKlineDaily, error) {
	return nil, nil
}
func (m *mockDailyRepoForBacktest) FindByCode(ctx context.Context, code string, limit int) ([]*model.StockKlineDaily, error) {
	return m.klines, nil
}
func (m *mockDailyRepoForBacktest) FindLatestByCode(ctx context.Context, code string) (*model.StockKlineDaily, error) {
	return nil, nil
}
func (m *mockDailyRepoForBacktest) FindAllCodes(ctx context.Context) ([]string, error) {
	return []string{"600312"}, nil
}
func (m *mockDailyRepoForBacktest) Update(ctx context.Context, kline *model.StockKlineDaily) error {
	return nil
}
func (m *mockDailyRepoForBacktest) Delete(ctx context.Context, id uint) error {
	return nil
}
func (m *mockDailyRepoForBacktest) List(ctx context.Context, limit, offset int) ([]*model.StockKlineDaily, error) {
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
	klines[30] = &model.StockKlineDaily{
		Code: "600312", Date: "2026-02-01",
		Open: 10.3, High: 10.7, Low: 10.2, Close: 10.6, Volume: 350000,
	}
	klines[31] = &model.StockKlineDaily{
		Code: "600312", Date: "2026-02-02",
		Open: 10.6, High: 11.0, Low: 10.5, Close: 10.9, Volume: 280000,
	}
	for i := 32; i <= 46; i++ {
		prevClose := klines[i-1].Close
		close := prevClose - 0.12
		klines[i] = &model.StockKlineDaily{
			Code: "600312", Date: fmt.Sprintf("2026-02-%02d", i-29),
			Open: prevClose, High: prevClose + 0.02, Low: close - 0.02, Close: close, Volume: 60000,
		}
	}
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

func testStrategies() []strategy.Strategy {
	vsCfg := score.DefaultVolumeSurgeConfig()
	vsCfg.MaxPullbackPct = 20.0
	vsCfg.MaxPullbackDays = 15
	return []strategy.Strategy{
		score.NewVolumeSurge(vsCfg),
		score.NewKDJOverSold(score.DefaultKDJFilterConfig()),
		score.NewMA60Trend(score.DefaultMA60TrendConfig()),
	}
}

func TestBacktestServiceScan(t *testing.T) {
	repo := &mockDailyRepoForBacktest{klines: buildDailyKlines()}
	svc := NewBacktestService(repo)
	strs := testStrategies()

	signals, err := svc.Scan(context.Background(), "600312", strs, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 三策略交集应有信号
	if len(signals) == 0 {
		t.Fatal("expected at least one signal from intersection")
	}
}

func TestBacktestServiceScanSingleStrategy(t *testing.T) {
	repo := &mockDailyRepoForBacktest{klines: buildDailyKlines()}
	svc := NewBacktestService(repo)

	// 单策略：只要 KDJ 超卖
	kdj := score.NewKDJOverSold(score.DefaultKDJFilterConfig())
	signals, err := svc.Scan(context.Background(), "600312", []strategy.Strategy{kdj}, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(signals) == 0 {
		t.Fatal("expected signals from single KDJ strategy")
	}
}

func TestBacktestServiceRun(t *testing.T) {
	repo := &mockDailyRepoForBacktest{klines: buildDailyKlines()}
	svc := NewBacktestService(repo)
	strs := testStrategies()

	report, err := svc.Run(context.Background(), "600312", strs, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report == nil {
		t.Fatal("expected report")
	}
	if len(report.Strategies) != 3 {
		t.Fatalf("expected 3 strategies, got %d", len(report.Strategies))
	}
}
