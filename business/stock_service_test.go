package business

import (
	"context"
	"errors"
	"testing"

	"trading/model"
)

// mockBroker 模拟行情数据提供者
type mockBroker struct {
	historicalData []model.StockKline
	historicalErr  error
}

func (m *mockBroker) GetStockTodayInBatch(ctx context.Context, codes []string) (map[string]*model.StockKline, error) {
	return nil, nil
}

func (m *mockBroker) GetStockToday(ctx context.Context, code string) (*model.StockKline, error) {
	return nil, nil
}

func (m *mockBroker) GetStockHistorical(ctx context.Context, symbol string, scale int, length int) ([]model.StockKline, error) {
	return m.historicalData, m.historicalErr
}

// mockRepo 模拟 StockKline 数据仓库
type mockRepo struct {
	k       []*model.StockKline
	findErr error
	upErr   error
}

func (m *mockRepo) Create(ctx context.Context, kline *model.StockKline) error         { return nil }
func (m *mockRepo) CreateBatch(ctx context.Context, klines []*model.StockKline) error { return nil }
func (m *mockRepo) Upsert(ctx context.Context, klines []*model.StockKline) error      { return m.upErr }
func (m *mockRepo) FindByID(ctx context.Context, id uint) (*model.StockKline, error)  { return nil, nil }
func (m *mockRepo) FindByCode(ctx context.Context, code string, limit int) ([]*model.StockKline, error) {
	return m.k, m.findErr
}
func (m *mockRepo) Update(ctx context.Context, kline *model.StockKline) error { return nil }
func (m *mockRepo) Delete(ctx context.Context, id uint) error                 { return nil }
func (m *mockRepo) List(ctx context.Context, limit, offset int) ([]*model.StockKline, error) {
	return nil, nil
}

// TestStockServiceSaveHistoricalDataSuccess 成功保存历史数据
func TestStockServiceSaveHistoricalDataSuccess(t *testing.T) {
	broker := &mockBroker{
		historicalData: []model.StockKline{
			{Code: "sh000001", Date: "2025-04-20", Open: 1, High: 2, Low: 0.5, Close: 1.5, Volume: 100},
			{Code: "sh000001", Date: "2025-04-21", Open: 1.5, High: 3, Low: 1, Close: 2, Volume: 200},
		},
	}
	repo := &mockRepo{}

	svc := NewStockService(broker, repo)
	err := svc.SaveHistoricalData(context.Background(), "000001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// TestStockServiceSaveHistoricalDataInvalidCode 非法股票代码
func TestStockServiceSaveHistoricalDataInvalidCode(t *testing.T) {
	svc := NewStockService(&mockBroker{}, &mockRepo{})
	err := svc.SaveHistoricalData(context.Background(), "999999")
	if err == nil {
		t.Fatal("expected error for invalid code, got nil")
	}
}

// TestStockServiceSaveHistoricalDataBrokerError broker 失败
func TestStockServiceSaveHistoricalDataBrokerError(t *testing.T) {
	broker := &mockBroker{historicalErr: errors.New("broker down")}
	svc := NewStockService(broker, &mockRepo{})

	err := svc.SaveHistoricalData(context.Background(), "000001")
	if err == nil {
		t.Fatal("expected error when broker fails")
	}
}

// TestStockServiceSaveHistoricalDataRepoError repo upsert 失败
func TestStockServiceSaveHistoricalDataRepoError(t *testing.T) {
	broker := &mockBroker{
		historicalData: []model.StockKline{
			{Code: "sh000001", Date: "2025-04-20", Open: 1, High: 2, Low: 0.5, Close: 1.5, Volume: 100},
		},
	}
	repo := &mockRepo{upErr: errors.New("db down")}
	svc := NewStockService(broker, repo)

	err := svc.SaveHistoricalData(context.Background(), "000001")
	if err == nil {
		t.Fatal("expected error when repo upsert fails")
	}
}

// TestStockServiceGetStockDataSuccess 成功获取数据
func TestStockServiceGetStockDataSuccess(t *testing.T) {
	repo := &mockRepo{
		k: []*model.StockKline{
			{Code: "000001", Date: "2025-04-20", Open: 1, High: 2, Low: 0.5, Close: 1.5, Volume: 100},
			{Code: "000001", Date: "2025-04-21", Open: 1.5, High: 3, Low: 1, Close: 2, Volume: 200},
		},
	}
	svc := NewStockService(&mockBroker{}, repo)

	data, err := svc.GetStockData(context.Background(), "000001", 240, 2)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(data) != 2 {
		t.Fatalf("expected 2 items, got %d", len(data))
	}
}

// TestStockServiceGetStockDataRepoError repo 查询失败
func TestStockServiceGetStockDataRepoError(t *testing.T) {
	repo := &mockRepo{findErr: errors.New("db down")}
	svc := NewStockService(&mockBroker{}, repo)

	_, err := svc.GetStockData(context.Background(), "000001", 240, 10)
	if err == nil {
		t.Fatal("expected error when repo fails")
	}
}

// TestStockServiceGetStockDataAggregation scale 聚合
func TestStockServiceGetStockDataAggregation(t *testing.T) {
	repo := &mockRepo{
		k: []*model.StockKline{
			{Code: "000001", Date: "2025-04-18", Open: 1, High: 2, Low: 0.5, Close: 1.5, Volume: 100},
			{Code: "000001", Date: "2025-04-19", Open: 1.5, High: 3, Low: 1, Close: 2, Volume: 200},
			{Code: "000001", Date: "2025-04-20", Open: 2, High: 4, Low: 1.5, Close: 3, Volume: 300},
			{Code: "000001", Date: "2025-04-21", Open: 3, High: 5, Low: 2, Close: 4, Volume: 400},
		},
	}
	svc := NewStockService(&mockBroker{}, repo)

	// scale=480, 每 2 条聚合成 1 条，取 1 条
	data, err := svc.GetStockData(context.Background(), "000001", 480, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(data) != 1 {
		t.Fatalf("expected 1 aggregated item, got %d", len(data))
	}

	agg := data[0]
	if agg.Open != 2 { // 第 3 条的 Open
		t.Errorf("Open = %.2f, want 2.00", agg.Open)
	}
	if agg.High != 5 { // max High
		t.Errorf("High = %.2f, want 5.00", agg.High)
	}
	if agg.Low != 1.5 { // min Low
		t.Errorf("Low = %.2f, want 1.50", agg.Low)
	}
	if agg.Close != 4 { // 第 4 条的 Close
		t.Errorf("Close = %.2f, want 4.00", agg.Close)
	}
	if agg.Volume != 700 { // 300 + 400
		t.Errorf("Volume = %d, want 700", agg.Volume)
	}
	if agg.Date != "2025-04-21" {
		t.Errorf("Date = %s, want 2025-04-21", agg.Date)
	}
}

// TestStockServiceGetStockDataDefaultLen len 默认值为 240
func TestStockServiceGetStockDataDefaultLen(t *testing.T) {
	repo := &mockRepo{
		k: []*model.StockKline{
			{Code: "000001", Date: "2025-04-20", Open: 1, High: 2, Low: 0.5, Close: 1.5, Volume: 100},
		},
	}
	svc := NewStockService(&mockBroker{}, repo)

	// len <= 0 时使用默认值 240
	data, err := svc.GetStockData(context.Background(), "000001", 240, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(data) != 1 {
		t.Fatalf("expected 1 item, got %d", len(data))
	}
}
