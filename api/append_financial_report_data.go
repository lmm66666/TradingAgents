package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AppendFinancialReportData 手动触发财报数据补全扫描（异步）
func (h *StockHandler) AppendFinancialReportData(c *gin.Context) {
	if err := h.financialScheduler.TriggerNow(c.Request.Context()); err != nil {
		if strings.Contains(err.Error(), "already running") {
			respondError(c, http.StatusTooManyRequests, err.Error())
			return
		}
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	respondSuccess(c, nil)
}
