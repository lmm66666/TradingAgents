package api

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"

	"trading/business"
)

// mockStockService 模拟 StockService
type mockStockService struct {
	saveErr  error
	analysis *business.StockAnalysisData
	getErr   error
}

func (m *mockStockService) SaveHistoricalData(ctx context.Context, code string) error {
	return m.saveErr
}

func (m *mockStockService) GetStockAnalysisData(ctx context.Context, code string) (*business.StockAnalysisData, error) {
	return m.analysis, m.getErr
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

func setupTestRouter(svc business.StockService, scheduler business.Scheduler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewStockHandler(svc, scheduler, nil)
	r.POST("/api/stocks/historical", h.SaveStockHistoricalData)
	r.GET("/api/stocks/analysis", h.GetStockAnalysisData)
	r.POST("/api/stocks/append", h.AppendStockData)
	return r
}
