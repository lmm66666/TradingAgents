package api

import (
	"github.com/gin-gonic/gin"

	"trading/business"
	"trading/pkg/backtest"
)

// NewRouter 创建 gin 路由
func NewRouter(svc business.StockService, scheduler business.Scheduler, backtestSvc backtest.BacktestService) *gin.Engine {
	r := gin.Default()
	h := NewStockHandler(svc, scheduler, backtestSvc)

	r.POST("/api/stocks/historical", h.SaveStockHistoricalData)
	r.GET("/api/stocks/analysis", h.GetStockAnalysisData)
	r.POST("/api/stocks/append", h.AppendStockData)
	r.POST("/api/patterns/scan", h.ScanPatterns)
	r.GET("/api/patterns/backtest", h.BacktestPatterns)

	return r
}
