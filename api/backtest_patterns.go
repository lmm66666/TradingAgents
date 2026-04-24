package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"trading/pkg/strategy"
)

// BacktestPatterns GET /api/patterns/backtest?code=xxx&strategies=volume_surge,kdj_oversold,ma60_trend&hold_days=5
func (h *StockHandler) BacktestPatterns(c *gin.Context) {
	code := c.Query("code")
	strategiesStr := c.Query("strategies")
	holdDaysStr := c.Query("hold_days")
	if code == "" || strategiesStr == "" {
		respondError(c, http.StatusBadRequest, "code and strategies are required")
		return
	}

	holdDays, _ := strconv.Atoi(holdDaysStr)
	if holdDays <= 0 {
		holdDays = 5
	}

	strs := make([]strategy.Strategy, 0)
	for _, name := range strings.Split(strategiesStr, ",") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		st, err := strategy.ResolveStrategy(name)
		if err != nil {
			respondError(c, http.StatusBadRequest, err.Error())
			return
		}
		strs = append(strs, st)
	}

	if len(strs) == 0 {
		respondError(c, http.StatusBadRequest, "at least one strategy is required")
		return
	}

	report, err := h.backtestSvc.Run(c.Request.Context(), code, strs, holdDays)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	respondSuccess(c, report)
}
