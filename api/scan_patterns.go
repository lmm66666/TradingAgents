package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"trading/pkg/strategy"
)

type scanPatternsRequest struct {
	Strategy string   `json:"strategy" binding:"required"`
	MinScore float64  `json:"min_score"`
	Codes    []string `json:"codes"`
}

// ScanPatterns POST /api/patterns/scan
func (h *StockHandler) ScanPatterns(c *gin.Context) {
	var req scanPatternsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	st, err := strategy.ResolveStrategy(req.Strategy)
	if err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	if req.MinScore == 0 {
		req.MinScore = 70
	}

	var signals []strategy.Signal

	if len(req.Codes) > 0 {
		for _, code := range req.Codes {
			result, scanErr := h.patternSvc.Scan(c.Request.Context(), code, st)
			if scanErr != nil {
				continue
			}
			for _, s := range result {
				if s.Score >= req.MinScore {
					signals = append(signals, s)
				}
			}
		}
	} else {
		signals, err = h.patternSvc.ScanAll(c.Request.Context(), st, req.MinScore)
		if err != nil {
			respondError(c, http.StatusInternalServerError, err.Error())
			return
		}
	}

	respondSuccess(c, gin.H{"signals": signals})
}
