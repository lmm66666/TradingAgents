package api

import (
	"github.com/gin-gonic/gin"

	"trading/business"
)

// NewRouter 创建 gin 路由
func NewRouter(svc business.StockService, scheduler business.Scheduler, analysisSvc business.AnalysisService) *gin.Engine {
	r := gin.Default()
	h := NewStockHandler(svc, scheduler, analysisSvc)

	r.POST("/api/stocks/historical", h.SaveStockHistoricalData)
	r.POST("/api/stocks/append", h.AppendStockData)
	r.POST("/api/stocks/financial-report", h.SaveFinancialReportData)
	r.GET("/api/stocks/signal", h.GetStockBuySignals)
	return r
}
