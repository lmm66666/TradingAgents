package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetStockAnalysisData 获取股票分析数据
func (h *StockHandler) GetStockAnalysisData(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		respondError(c, http.StatusBadRequest, "code is required")
		return
	}

	data, err := h.svc.GetStockAnalysisData(c.Request.Context(), code)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	respondSuccess(c, data)
}
