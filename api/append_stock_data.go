package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AppendStockData 手动触发数据补全扫描
func (h *StockHandler) AppendStockData(c *gin.Context) {
	if err := h.scheduler.TriggerNow(c.Request.Context()); err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	respondSuccess(c, nil)
}
