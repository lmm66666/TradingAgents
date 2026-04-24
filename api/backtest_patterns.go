package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"trading/pkg/strategy"
)

// BacktestPatterns GET /api/patterns/backtest
func (h *StockHandler) BacktestPatterns(c *gin.Context) {
	code := c.Query("code")
	strategyName := c.Query("strategy")
	holdDaysStr := c.Query("hold_days")
	if code == "" || strategyName == "" {
		respondError(c, http.StatusBadRequest, "code and strategy are required")
		return
	}

	holdDays, _ := strconv.Atoi(holdDaysStr)
	if holdDays <= 0 {
		holdDays = 5
	}

	st, err := strategy.ResolveStrategy(strategyName)
	if err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	report, err := h.patternSvc.Backtest(c.Request.Context(), code, st, holdDays)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	respondSuccess(c, report)
}
