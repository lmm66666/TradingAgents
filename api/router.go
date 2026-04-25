package api

import (
	"github.com/gin-gonic/gin"

	"trading/business"
)

// NewRouter 创建 gin 路由
func NewRouter(svc business.StockService, scheduler business.Scheduler, financialScheduler business.FinancialScheduler, analysisSvc business.AnalysisService) *gin.Engine {
	r := gin.Default()
	h := NewStockHandler(svc, scheduler, financialScheduler, analysisSvc)

	r.POST("/api/stocks/historical", h.SaveStockHistoricalData)
	r.POST("/api/stocks/append", h.AppendStockData)
	r.POST("/api/stocks/financial-report", h.SaveFinancialReportData)
	r.POST("/api/stocks/financial-report/append", h.AppendFinancialReportData)
	r.GET("/api/stocks/signal", h.GetStockBuySignals)
	r.GET("/api/stocks/price", h.GetStockPrice)
	r.GET("/api/stocks/financial-report", h.GetFinancialReport)
	return r
}
