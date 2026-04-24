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

	var st strategy.Strategy
	switch strategyName {
	case "volume_surge_pullback":
		st = strategy.NewVolumeSurgePullback(strategy.DefaultVolumeSurgePullbackConfig())
	default:
		respondError(c, http.StatusBadRequest, "unknown strategy")
		return
	}

	report, err := h.patternSvc.Backtest(c.Request.Context(), code, st, holdDays)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	respondSuccess(c, report)
}
