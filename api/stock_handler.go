package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"trading/business"
)

// StockHandler 股票数据 HTTP 处理器
type StockHandler struct {
	svc business.StockService
}

// NewStockHandler 创建 StockHandler
func NewStockHandler(svc business.StockService) *StockHandler {
	return &StockHandler{svc: svc}
}

// SaveStockHistoricalData 从 broker 获取历史数据并保存
func (h *StockHandler) SaveStockHistoricalData(c *gin.Context) {
	var req saveHistoricalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Code == "" {
		respondError(c, http.StatusBadRequest, "code is required")
		return
	}

	if err := h.svc.SaveHistoricalData(c.Request.Context(), req.Code); err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	respondSuccess(c, nil)
}

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

// saveHistoricalRequest 保存历史数据请求体
type saveHistoricalRequest struct {
	Code string `json:"code" binding:"required"`
}

// response 统一 JSON 响应结构
type response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func respondSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, response{Code: 0, Message: "success", Data: data})
}

func respondError(c *gin.Context, status int, message string) {
	c.JSON(status, response{Code: status, Message: message, Data: nil})
}
