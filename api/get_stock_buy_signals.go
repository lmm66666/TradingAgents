package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetStockBuySignals GET /api/stocks/analysis
// 扫描所有股票，返回今日出现买点的股票代码列表
func (h *StockHandler) GetStockBuySignals(c *gin.Context) {
	codes, err := h.analysisSvc.FindBuySignals(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	respondSuccess(c, gin.H{"codes": codes})
}
