package business

import (
	"context"
	"errors"
	"testing"
	"time"

	"trading/model"
)

// mockDailyRepoForScheduler 模拟日线数据仓库（支持 FindLatestByCode / FindAllCodes）
type mockDailyRepoForScheduler struct {
	codes       []string
	latest      map[string]*model.StockKlineDaily
	latestErr   map[string]error
	findAllErr  error
}

func (m *mockDailyRepoForScheduler) Create(ctx context.Context, kline *model.StockKlineDaily) error         { return nil }
func (m *mockDailyRepoForScheduler) CreateBatch(ctx context.Context, klines []*model.StockKlineDaily) error { return nil }
func (m *mockDailyRepoForScheduler) Upsert(ctx context.Context, klines []*model.StockKlineDaily) error      { return nil }
func (m *mockDailyRepoForScheduler) FindByID(ctx context.Context, id uint) (*model.StockKlineDaily, error)  { return nil, nil }
func (m *mockDailyRepoForScheduler) FindByCode(ctx context.Context, code string, limit int) ([]*model.StockKlineDaily, error) {
	return nil, nil
}
func (m *mockDailyRepoForScheduler) FindLatestByCode(ctx context.Context, code string) (*model.StockKlineDaily, error) {
	if err, ok := m.latestErr[code]; ok {
		return nil, err
	}
	if k, ok := m.latest[code]; ok {
		return k, nil
	}
	return nil, errors.New("not found")
}
func (m *mockDailyRepoForScheduler) FindAllCodes(ctx context.Context) ([]string, error) {
	if m.findAllErr != nil {
		return nil, m.findAllErr
	}
	return m.codes, nil
}
func (m *mockDailyRepoForScheduler) Update(ctx context.Context, kline *model.StockKlineDaily) error { return nil }
func (m *mockDailyRepoForScheduler) Delete(ctx context.Context, id uint) error                      { return nil }
func (m *mockDailyRepoForScheduler) List(ctx context.Context, limit, offset int) ([]*model.StockKlineDaily, error) {
	return nil, nil
}

// mockWeeklyRepoForScheduler 模拟周线数据仓库（支持 FindLatestByCode / FindAllCodes）
type mockWeeklyRepoForScheduler struct {
	codes       []string
	latest      map[string]*model.StockKlineWeekly
	latestErr   map[string]error
	findAllErr  error
}

func (m *mockWeeklyRepoForScheduler) Create(ctx context.Context, kline *model.StockKlineWeekly) error         { return nil }
func (m *mockWeeklyRepoForScheduler) CreateBatch(ctx context.Context, klines []*model.StockKlineWeekly) error { return nil }
func (m *mockWeeklyRepoForScheduler) Upsert(ctx context.Context, klines []*model.StockKlineWeekly) error      { return nil }
func (m *mockWeeklyRepoForScheduler) FindByID(ctx context.Context, id uint) (*model.StockKlineWeekly, error)  { return nil, nil }
func (m *mockWeeklyRepoForScheduler) FindByCode(ctx context.Context, code string, limit int) ([]*model.StockKlineWeekly, error) {
	return nil, nil
}
func (m *mockWeeklyRepoForScheduler) FindLatestByCode(ctx context.Context, code string) (*model.StockKlineWeekly, error) {
	if err, ok := m.latestErr[code]; ok {
		return nil, err
	}
	if k, ok := m.latest[code]; ok {
		return k, nil
	}
	return nil, errors.New("not found")
}
func (m *mockWeeklyRepoForScheduler) FindAllCodes(ctx context.Context) ([]string, error) {
	if m.findAllErr != nil {
		return nil, m.findAllErr
	}
	return m.codes, nil
}
func (m *mockWeeklyRepoForScheduler) Update(ctx context.Context, kline *model.StockKlineWeekly) error { return nil }
func (m *mockWeeklyRepoForScheduler) Delete(ctx context.Context, id uint) error                      { return nil }
func (m *mockWeeklyRepoForScheduler) List(ctx context.Context, limit, offset int) ([]*model.StockKlineWeekly, error) {
	return nil, nil
}

// mockSvcForScheduler 模拟 StockService
type mockSvcForScheduler struct {
	saveErr error
}

func (m *mockSvcForScheduler) SaveHistoricalData(ctx context.Context, code string) error {
	return m.saveErr
}
func (m *mockSvcForScheduler) GetStockAnalysisData(ctx context.Context, code string) (*StockAnalysisData, error) {
	return nil, nil
}

func TestLastFridayDate(t *testing.T) {
	tests := []struct {
		input    time.Time
		expected string
	}{
		{time.Date(2025, 4, 21, 0, 0, 0, 0, time.UTC), "2025-04-18"}, // Monday -> last Friday
		{time.Date(2025, 4, 22, 0, 0, 0, 0, time.UTC), "2025-04-18"}, // Tuesday
		{time.Date(2025, 4, 23, 0, 0, 0, 0, time.UTC), "2025-04-18"}, // Wednesday
		{time.Date(2025, 4, 24, 0, 0, 0, 0, time.UTC), "2025-04-18"}, // Thursday
		{time.Date(2025, 4, 25, 0, 0, 0, 0, time.UTC), "2025-04-25"}, // Friday -> today
		{time.Date(2025, 4, 26, 0, 0, 0, 0, time.UTC), "2025-04-25"}, // Saturday -> last Friday
		{time.Date(2025, 4, 27, 0, 0, 0, 0, time.UTC), "2025-04-25"}, // Sunday -> last Friday
	}

	for _, tt := range tests {
		got := lastFridayDate(tt.input)
		if got != tt.expected {
			t.Errorf("lastFridayDate(%v) = %s, want %s", tt.input, got, tt.expected)
		}
	}
}

func TestSchedulerScanAllUpToDate(t *testing.T) {
	today := time.Now().Format("2006-01-02")
	lastFriday := lastFridayDate(time.Now())

	dailyRepo := &mockDailyRepoForScheduler{
		codes: []string{"000001"},
		latest: map[string]*model.StockKlineDaily{
			"000001": {Code: "000001", Date: today},
		},
	}
	weeklyRepo := &mockWeeklyRepoForScheduler{
		codes: []string{"000001"},
		latest: map[string]*model.StockKlineWeekly{
			"000001": {Code: "000001", Date: lastFriday},
		},
	}

	svc := &mockSvcForScheduler{}
	sched := NewScheduler(svc, dailyRepo, weeklyRepo).(*stockScheduler)

	tasks, err := sched.scan(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("expected 0 tasks, got %d", len(tasks))
	}
}

func TestSchedulerScanMissingDaily(t *testing.T) {
	lastFriday := lastFridayDate(time.Now())

	dailyRepo := &mockDailyRepoForScheduler{
		codes: []string{"000001"},
		latest: map[string]*model.StockKlineDaily{
			"000001": {Code: "000001", Date: "2020-01-01"},
		},
	}
	weeklyRepo := &mockWeeklyRepoForScheduler{
		codes: []string{"000001"},
		latest: map[string]*model.StockKlineWeekly{
			"000001": {Code: "000001", Date: lastFriday},
		},
	}

	svc := &mockSvcForScheduler{}
	sched := NewScheduler(svc, dailyRepo, weeklyRepo).(*stockScheduler)

	tasks, err := sched.scan(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if !tasks[0].needDaily {
		t.Fatal("expected needDaily to be true")
	}
	if tasks[0].needWeekly {
		t.Fatal("expected needWeekly to be false")
	}
}

func TestSchedulerScanMissingWeekly(t *testing.T) {
	today := time.Now().Format("2006-01-02")

	dailyRepo := &mockDailyRepoForScheduler{
		codes: []string{"000001"},
		latest: map[string]*model.StockKlineDaily{
			"000001": {Code: "000001", Date: today},
		},
	}
	weeklyRepo := &mockWeeklyRepoForScheduler{
		codes: []string{"000001"},
		latest: map[string]*model.StockKlineWeekly{
			"000001": {Code: "000001", Date: "2020-01-01"},
		},
	}

	svc := &mockSvcForScheduler{}
	sched := NewScheduler(svc, dailyRepo, weeklyRepo).(*stockScheduler)

	tasks, err := sched.scan(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].needDaily {
		t.Fatal("expected needDaily to be false")
	}
	if !tasks[0].needWeekly {
		t.Fatal("expected needWeekly to be true")
	}
}

func TestSchedulerScanUnionCodes(t *testing.T) {
	today := time.Now().Format("2006-01-02")
	lastFriday := lastFridayDate(time.Now())

	// daily has code A, weekly has code B -> union should include both
	dailyRepo := &mockDailyRepoForScheduler{
		codes: []string{"000001"},
		latest: map[string]*model.StockKlineDaily{
			"000001": {Code: "000001", Date: today},
		},
	}
	weeklyRepo := &mockWeeklyRepoForScheduler{
		codes: []string{"000002"},
		latest: map[string]*model.StockKlineWeekly{
			"000002": {Code: "000002", Date: lastFriday},
		},
	}

	svc := &mockSvcForScheduler{}
	sched := NewScheduler(svc, dailyRepo, weeklyRepo).(*stockScheduler)

	tasks, err := sched.scan(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	// 000001 missing in weekly, 000002 missing in daily -> 2 tasks
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
}

func TestSchedulerProcessSuccess(t *testing.T) {
	svc := &mockSvcForScheduler{}
	sched := NewScheduler(svc, &mockDailyRepoForScheduler{}, &mockWeeklyRepoForScheduler{}).(*stockScheduler)

	err := sched.process(context.Background(), task{code: "000001", needDaily: true})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestSchedulerProcessError(t *testing.T) {
	svc := &mockSvcForScheduler{saveErr: errors.New("save failed")}
	sched := NewScheduler(svc, &mockDailyRepoForScheduler{}, &mockWeeklyRepoForScheduler{}).(*stockScheduler)

	err := sched.process(context.Background(), task{code: "000001", needDaily: true})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
