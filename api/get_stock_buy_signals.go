package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetStockBuySignals GET /api/stocks/signal
// 扫描所有股票，按策略分组返回今日出现买点的股票代码列表
func (h *StockHandler) GetStockBuySignals(c *gin.Context) {
	signals, err := h.analysisSvc.FindBuySignals(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	respondSuccess(c, gin.H{"strategies": signals})
}
