package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"trading/pkg/strategy"
)

type scanPatternsRequest struct {
	Strategies []string `json:"strategies" binding:"required"`
	MinScore   float64  `json:"min_score"`
	Codes      []string `json:"codes"`
}

// ScanPatterns POST /api/patterns/scan
func (h *StockHandler) ScanPatterns(c *gin.Context) {
	var req scanPatternsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	strs := make([]strategy.Strategy, 0, len(req.Strategies))
	for _, name := range req.Strategies {
		st, err := resolveStrategy(name)
		if err != nil {
			respondError(c, http.StatusBadRequest, err.Error())
			return
		}
		strs = append(strs, st)
	}

	if req.MinScore == 0 {
		req.MinScore = 70
	}

	var signals []strategy.Signal
	var err error

	if len(req.Codes) > 0 {
		for _, code := range req.Codes {
			result, scanErr := h.backtestSvc.Scan(c.Request.Context(), code, strs, req.MinScore)
			if scanErr != nil {
				continue
			}
			signals = append(signals, result...)
		}
	} else {
		signals, err = h.backtestSvc.ScanAll(c.Request.Context(), strs, req.MinScore)
		if err != nil {
			respondError(c, http.StatusInternalServerError, err.Error())
			return
		}
	}

	respondSuccess(c, gin.H{"signals": signals})
}
