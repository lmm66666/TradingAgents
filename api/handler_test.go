package api

import (
	"context"

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

// mockScheduler 模拟调度器
type mockScheduler struct {
	triggerErr error
}

func (m *mockScheduler) Start(ctx context.Context, hour, minute int) {}
func (m *mockScheduler) Stop()                                      {}
func (m *mockScheduler) TriggerNow(ctx context.Context) error {
	return m.triggerErr
}

func setupTestRouter(svc business.StockService, scheduler business.Scheduler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewStockHandler(svc, scheduler)
	r.POST("/api/stocks/historical", h.SaveStockHistoricalData)
	r.GET("/api/stocks/analysis", h.GetStockAnalysisData)
	r.POST("/api/stocks/append", h.AppendStockData)
	return r
}
