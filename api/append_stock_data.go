package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AppendStockData 手动触发数据补全扫描（异步）
func (h *StockHandler) AppendStockData(c *gin.Context) {
	if err := h.scheduler.TriggerNow(c.Request.Context()); err != nil {
		if strings.Contains(err.Error(), "already running") {
			respondError(c, http.StatusTooManyRequests, err.Error())
			return
		}
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	respondSuccess(c, nil)
}
