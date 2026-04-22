package api

import (
	"github.com/gin-gonic/gin"

	"trading/business"
)

// NewRouter 创建 gin 路由
func NewRouter(svc business.StockService) *gin.Engine {
	r := gin.Default()
	h := NewStockHandler(svc)

	r.POST("/api/stocks/historical", h.SaveStockHistoricalData)
	r.GET("/api/stocks/data", h.GetStockData)

	return r
}
