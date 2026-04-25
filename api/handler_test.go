package api

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"

	"trading/business"
)

// mockStockService 模拟 StockService
type mockStockService struct {
	saveErr error
}

func (m *mockStockService) SaveHistoricalData(ctx context.Context, code string) error {
	return m.saveErr
}

func (m *mockStockService) AppendStockData(ctx context.Context, code string) error {
	return m.saveErr
}

// mockScheduler 模拟调度器
type mockScheduler struct {
	triggerErr     error
	alreadyRunning bool
}

func (m *mockScheduler) Start(ctx context.Context, hour, minute int) {}
func (m *mockScheduler) Stop()                                      {}
func (m *mockScheduler) TriggerNow(ctx context.Context) error {
	if m.alreadyRunning {
		return errors.New("another task is already running")
	}
	return m.triggerErr
}

func setupTestRouter(svc business.StockService, scheduler business.Scheduler, analysisSvc business.AnalysisService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewStockHandler(svc, scheduler, analysisSvc)
	r.POST("/api/stocks/historical", h.SaveStockHistoricalData)
	r.GET("/api/stocks/signal", h.GetStockBuySignals)
	r.POST("/api/stocks/append", h.AppendStockData)
	return r
}
