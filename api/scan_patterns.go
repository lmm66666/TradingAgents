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

	var st strategy.Strategy
	switch req.Strategy {
	case "volume_surge_pullback":
		st = strategy.NewVolumeSurgePullback(strategy.DefaultVolumeSurgePullbackConfig())
	default:
		respondError(c, http.StatusBadRequest, "unknown strategy")
		return
	}

	if req.MinScore == 0 {
		req.MinScore = 70
	}

	var signals []strategy.Signal
	var err error

	if len(req.Codes) > 0 {
		signals, err = h.patternSvc.ScanAll(c.Request.Context(), st, req.MinScore)
		if err != nil {
			respondError(c, http.StatusInternalServerError, err.Error())
			return
		}
		// Filter to requested codes
		codeSet := make(map[string]struct{}, len(req.Codes))
		for _, code := range req.Codes {
			codeSet[code] = struct{}{}
		}
		filtered := make([]strategy.Signal, 0, len(signals))
		for _, s := range signals {
			if _, ok := codeSet[s.Code]; ok {
				filtered = append(filtered, s)
			}
		}
		signals = filtered
	} else {
		signals, err = h.patternSvc.ScanAll(c.Request.Context(), st, req.MinScore)
		if err != nil {
			respondError(c, http.StatusInternalServerError, err.Error())
			return
		}
	}

	respondSuccess(c, gin.H{"signals": signals})
}
