package api

import (
	"github.com/gin-gonic/gin"

	"trading/business"
)

// NewRouter 创建 gin 路由
func NewRouter(svc business.StockService, scheduler Scheduler) *gin.Engine {
	r := gin.Default()
	h := NewStockHandler(svc, scheduler)

	r.POST("/api/stocks/historical", h.SaveStockHistoricalData)
	r.GET("/api/stocks/analysis", h.GetStockAnalysisData)
	r.POST("/api/stocks/append", h.AppendStockData)

	return r
}
